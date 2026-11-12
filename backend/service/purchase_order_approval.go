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
	return s.inTx(ctx, "failed to save purchase order", func(tx *sqlx.Tx) error {
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
		return fn(tx, order)
	})
}

// SubmitPurchaseOrder is a drafter saying "this is done": the draft goes to the managers for
// approval. Any drafter can submit any draft, so a draft someone else fixed need not wait for its
// author; the audit entry names who submitted. It can be pulled back with WithdrawPurchaseOrder.
func (s *inventoryService) SubmitPurchaseOrder(ctx context.Context, id, version, userID int) error {
	if !canWorkOnPurchaseOrders(ctx) {
		return httperror.Forbidden("you cannot submit purchase orders", nil)
	}
	return s.inPurchaseOrderTx(
		ctx,
		id,
		version,
		func(tx *sqlx.Tx, order types.PurchaseOrder) error {
			if order.Status != constant.PurchaseOrderDraft {
				return httperror.NewHttpError(
					http.StatusConflict, "only a draft can be submitted for approval", nil,
				)
			}
			if len(order.Items) == 0 {
				return badRequest("add at least one item before submitting")
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

// WithdrawPurchaseOrder takes a submitted order back to a draft, so a mistake found after
// submitting can be fixed without asking a manager to reject the order. Anyone who can submit can
// pull it back; approval is the point after which only managers act.
func (s *inventoryService) WithdrawPurchaseOrder(
	ctx context.Context,
	id, version, userID int,
) error {
	if !canWorkOnPurchaseOrders(ctx) {
		return httperror.Forbidden("you cannot change purchase orders", nil)
	}
	return s.inPurchaseOrderTx(
		ctx,
		id,
		version,
		func(tx *sqlx.Tx, order types.PurchaseOrder) error {
			if order.Status != constant.PurchaseOrderPending {
				return httperror.NewHttpError(
					http.StatusConflict,
					"only a submitted order can be pulled back to a draft",
					nil,
				)
			}
			err := s.repo.SetPurchaseOrderStatusTx(ctx, tx, id, constant.PurchaseOrderDraft)
			if err != nil {
				return httperror.ServerError("failed to pull back purchase order", err)
			}
			return s.auditPurchaseOrder(
				ctx,
				tx,
				userID,
				id,
				constant.AuditWithdrawn,
				"",
				map[string]any{},
			)
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
				return badRequest("a purchase order needs at least one item")
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
