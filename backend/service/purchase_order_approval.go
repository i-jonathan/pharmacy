package service

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"pharmacy/httperror"
	"pharmacy/internal/constant"
	"pharmacy/internal/types"

	"github.com/jmoiron/sqlx"
)

// inPurchaseOrderTx runs fn in a transaction that holds the order's row lock, after checking the
// caller is acting on the version they saw. It commits when fn succeeds.
func (s *inventoryService) inPurchaseOrderTx(
	ctx context.Context, id, version int, fn func(tx *sqlx.Tx, order types.PurchaseOrder) error,
) error {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return httperror.ServerError("failed to start transaction", err)
	}
	committed := false
	defer func() {
		if !committed {
			s.repo.RollbackTx(tx)
		}
	}()

	order, err := s.repo.GetPurchaseOrderForUpdateTx(ctx, tx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return purchaseOrderNotFound(err)
	}
	if err != nil {
		return httperror.ServerError("failed to fetch purchase order", err)
	}
	if order.Version != version {
		return purchaseOrderChangedElsewhere()
	}
	if err = fn(tx, order); err != nil {
		return err
	}
	if err = s.repo.CommitTx(tx); err != nil {
		return httperror.ServerError("failed to save purchase order", err)
	}
	committed = true
	return nil
}

// SubmitPurchaseOrder is the drafter saying "I'm done": the draft goes to the managers for
// approval and can no longer be changed by its author.
func (s *inventoryService) SubmitPurchaseOrder(ctx context.Context, id, version, userID int) error {
	if !canDraftPurchaseOrders(ctx) && !canApprovePurchaseOrders(ctx) {
		return httperror.Forbidden("you cannot submit purchase orders", nil)
	}
	return s.inPurchaseOrderTx(
		ctx,
		id,
		version,
		func(tx *sqlx.Tx, order types.PurchaseOrder) error {
			if order.CreatedByID == nil || *order.CreatedByID != userID {
				return httperror.Forbidden("only the author can submit a draft", nil)
			}
			if order.Status != constant.PurchaseOrderDraft {
				return httperror.NewHttpError(
					http.StatusConflict, "only a draft can be submitted for approval", nil,
				)
			}
			if len(order.Items) == 0 {
				return purchaseOrderBadRequest("add at least one item before submitting")
			}
			err := s.repo.SetPurchaseOrderStatusTx(ctx, tx, id, constant.PurchaseOrderPending)
			if err != nil {
				return httperror.ServerError("failed to submit purchase order", err)
			}
			details := map[string]any{"items": auditItemSummaries(order.Items)}
			return s.auditPurchaseOrder(ctx, tx, userID, id, constant.AuditSubmitted, "", details)
		},
	)
}

// ApprovePurchaseOrder turns a draft or submitted order into an open one, ready to be sent to
// the supplier and received against.
func (s *inventoryService) ApprovePurchaseOrder(
	ctx context.Context,
	id, version, userID int,
) error {
	if err := requireApprover(ctx); err != nil {
		return err
	}
	return s.inPurchaseOrderTx(
		ctx,
		id,
		version,
		func(tx *sqlx.Tx, order types.PurchaseOrder) error {
			if !purchaseOrderAwaitingApproval(order.Status) {
				return httperror.NewHttpError(
					http.StatusConflict, "this purchase order does not need approval", nil,
				)
			}
			if len(order.Items) == 0 {
				return purchaseOrderBadRequest("a purchase order needs at least one item")
			}
			if err := s.repo.ApprovePurchaseOrderTx(ctx, tx, id, userID); err != nil {
				return httperror.ServerError("failed to approve purchase order", err)
			}
			details := map[string]any{
				"status_before": order.Status,
				"items":         auditItemSummaries(order.Items),
			}
			return s.auditPurchaseOrder(ctx, tx, userID, id, constant.AuditApproved, "", details)
		},
	)
}

// RejectPurchaseOrder declines a draft or submitted order. It is closed, not deleted, so the
// record and the reason stay.
func (s *inventoryService) RejectPurchaseOrder(
	ctx context.Context, id, version int, reason string, userID int,
) error {
	if err := requireApprover(ctx); err != nil {
		return err
	}
	reason, err := normalizeAuditReason(reason)
	if err != nil {
		return err
	}
	return s.inPurchaseOrderTx(
		ctx,
		id,
		version,
		func(tx *sqlx.Tx, order types.PurchaseOrder) error {
			if !purchaseOrderAwaitingApproval(order.Status) {
				return httperror.NewHttpError(
					http.StatusConflict, "only a draft or submitted order can be rejected", nil,
				)
			}
			if err := s.repo.ClosePurchaseOrderTx(ctx, tx, id); err != nil {
				return httperror.ServerError("failed to reject purchase order", err)
			}
			details := map[string]any{"status_before": order.Status}
			return s.auditPurchaseOrder(
				ctx,
				tx,
				userID,
				id,
				constant.AuditRejected,
				reason,
				details,
			)
		},
	)
}

// CountPendingPurchaseOrders feeds the managers' badge: how many orders wait for approval and
// how many drafts are still being written.
func (s *inventoryService) CountPendingPurchaseOrders(
	ctx context.Context,
) (types.PendingPurchaseOrders, error) {
	if err := requireApprover(ctx); err != nil {
		return types.PendingPurchaseOrders{}, err
	}
	counts, err := s.repo.CountPendingPurchaseOrders(ctx)
	if err != nil {
		return types.PendingPurchaseOrders{}, httperror.ServerError(
			"failed to count purchase orders",
			err,
		)
	}
	return types.PendingPurchaseOrders{
		Pending: counts[constant.PurchaseOrderPending],
		Drafts:  counts[constant.PurchaseOrderDraft],
	}, nil
}
