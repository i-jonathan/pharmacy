package service

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"pharmacy/httperror"
	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"strings"
	"unicode/utf8"

	"github.com/jmoiron/sqlx"
)

const (
	minAuditReason = 3
	maxAuditReason = 300
)

// change is a before and after pair in an audit record.
type change struct {
	From any `json:"from"`
	To   any `json:"to"`
}

// normalizeAuditReason requires a short explanation for actions that abandon or reassign what
// was ordered, so a later dispute can see why.
func normalizeAuditReason(reason string) (string, error) {
	reason = strings.TrimSpace(reason)
	if utf8.RuneCountInString(reason) < minAuditReason {
		return "", badRequest("give a reason")
	}
	if utf8.RuneCountInString(reason) > maxAuditReason {
		return "", badRequest(
			fmt.Sprintf("the reason is limited to %d characters", maxAuditReason),
		)
	}
	return reason, nil
}

// auditPurchaseOrder appends an audit record inside the caller's transaction.
func (s *inventoryService) auditPurchaseOrder(
	ctx context.Context, tx *sqlx.Tx, userID, orderID int, action, reason string, details any,
) error {
	return s.audit(
		ctx, tx, userID, constant.AuditEntityPurchaseOrder, orderID, action, reason, details,
	)
}

func auditItemSummary(item types.PurchaseOrderItem) map[string]any {
	summary := map[string]any{"name": item.Name, "quantity": item.Quantity}
	if item.ProductID != nil {
		summary["product_id"] = *item.ProductID
	}
	if item.Manufacturer != "" {
		summary["manufacturer"] = item.Manufacturer
	}
	if item.Note != "" {
		summary["note"] = item.Note
	}
	return summary
}

func auditItemSummaries(items []types.PurchaseOrderItem) []map[string]any {
	summaries := make([]map[string]any, len(items))
	for i, item := range items {
		summaries[i] = auditItemSummary(item)
	}
	return summaries
}

// describePurchaseOrderEdit lists what an edit changes: header fields, added and removed lines,
// and for kept lines each field that moved with its before and after value. It returns nil
// when the edit changes nothing.
func describePurchaseOrderEdit(
	stored types.PurchaseOrder, header types.NewPurchaseOrder, plan purchaseOrderPlan,
) map[string]any {
	changes := map[string]any{}
	if stored.Supplier != header.SupplierName {
		changes["supplier"] = change{From: stored.Supplier, To: header.SupplierName}
	}
	if stored.Note != header.Note {
		changes["note"] = change{From: stored.Note, To: header.Note}
	}
	if len(plan.inserts) > 0 {
		changes["added"] = auditItemSummaries(plan.inserts)
	}

	byID := make(map[int]types.PurchaseOrderItem, len(stored.Items))
	for _, item := range stored.Items {
		byID[item.ID] = item
	}
	if len(plan.deletes) > 0 {
		removed := make([]types.PurchaseOrderItem, 0, len(plan.deletes))
		for _, id := range plan.deletes {
			removed = append(removed, byID[id])
		}
		changes["removed"] = auditItemSummaries(removed)
	}

	var updated []map[string]any
	for _, item := range plan.updates {
		before := byID[item.ID]
		fields := map[string]change{}
		if before.Quantity != item.Quantity {
			fields["quantity"] = change{From: before.Quantity, To: item.Quantity}
		}
		if before.Note != item.Note {
			fields["note"] = change{From: before.Note, To: item.Note}
		}
		if before.Name != item.Name {
			fields["name"] = change{From: before.Name, To: item.Name}
		}
		if before.Manufacturer != item.Manufacturer {
			fields["manufacturer"] = change{From: before.Manufacturer, To: item.Manufacturer}
		}
		if !sameProduct(before.ProductID, item.ProductID) {
			fields["product_id"] = change{From: before.ProductID, To: item.ProductID}
		}
		if len(fields) > 0 {
			updated = append(updated, map[string]any{"item": before.Name, "changes": fields})
		}
	}
	if len(updated) > 0 {
		changes["updated"] = updated
	}

	if len(changes) == 0 {
		return nil
	}
	return changes
}

func sameProduct(a, b *int) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

// GetPurchaseOrderHistory returns the audit records of one purchase order, newest first.
func (s *inventoryService) GetPurchaseOrderHistory(
	ctx context.Context, id, userID int,
) ([]types.AuditRecord, error) {
	if canApprovePurchaseOrders(ctx) {
		exists, err := s.repo.PurchaseOrderExists(ctx, id)
		if err != nil {
			return nil, httperror.ServerError("failed to fetch purchase order", err)
		}
		if !exists {
			return nil, purchaseOrderNotFound(sql.ErrNoRows)
		}
	} else {
		// Everyone else may read the activity only of orders they wrote.
		order, err := s.GetPurchaseOrder(ctx, id, userID)
		if err != nil {
			return nil, err
		}
		if order.CreatedByID == nil || *order.CreatedByID != userID {
			return nil, purchaseOrderNotFound(sql.ErrNoRows)
		}
	}
	records, err := s.repo.FetchAuditLog(ctx, constant.AuditEntityPurchaseOrder, id)
	if err != nil {
		return nil, httperror.ServerError("failed to fetch purchase order history", err)
	}
	return records, nil
}

func purchaseOrderChangedElsewhere() error {
	return httperror.NewHttpError(
		http.StatusConflict,
		"this purchase order was changed by someone else; reload it and try again",
		nil,
	)
}
