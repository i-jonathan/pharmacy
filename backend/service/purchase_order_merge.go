package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"pharmacy/httperror"
	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"strings"
	"unicode/utf8"

	"github.com/jmoiron/sqlx"
)

// purchaseOrderMergedLine is a line of the order being merged into, as it will be saved.
type purchaseOrderMergedLine struct {
	item           types.PurchaseOrderItem
	quantityBefore int
	noteBefore     string
}

// purchaseOrderMergePlan is what folding a draft into an open order changes: the combined order
// note, the open order's lines that change, and the draft's lines that are added.
type purchaseOrderMergePlan struct {
	note    string
	updates []purchaseOrderMergedLine
	inserts []types.PurchaseOrderItem
}

// purchaseOrderMergeKey says which line of the target a draft line is the same item as: the
// product when both are products, otherwise the typed name and manufacturer. A typed line never
// matches a product line, as in the editor. The frontend preview applies the same rule.
func purchaseOrderMergeKey(item types.PurchaseOrderItem) string {
	if item.ProductID != nil {
		return fmt.Sprintf("p%d", *item.ProductID)
	}
	return "n" + strings.ToLower(item.Name+"\x00"+item.Manufacturer)
}

// joinMergedNotes keeps the first note, and adds the second after it when it says something
// new.
func joinMergedNotes(first, second, separator string) string {
	first, second = strings.TrimSpace(first), strings.TrimSpace(second)
	switch {
	case second == "" || second == first:
		return first
	case first == "":
		return second
	}
	return first + separator + second
}

// planPurchaseOrderMerge works out how a draft's lines and note fold into the open order that
// receives them. A line on both orders keeps the higher quantity (they are two writings of the
// same need, so they are not added together) and both line notes; a line only on the draft is
// added. It writes nothing.
func planPurchaseOrderMerge(target, draft types.PurchaseOrder) (purchaseOrderMergePlan, error) {
	plan := purchaseOrderMergePlan{note: joinMergedNotes(target.Note, draft.Note, "\n")}
	if utf8.RuneCountInString(plan.note) > maxPurchaseOrderNote {
		return plan, badRequest(fmt.Sprintf(
			"the two notes together pass the %d character limit; shorten one before merging",
			maxPurchaseOrderNote,
		))
	}

	existing := make(map[string]types.PurchaseOrderItem, len(target.Items))
	for _, item := range target.Items {
		existing[purchaseOrderMergeKey(item)] = item
	}
	for _, line := range draft.Items {
		current, ok := existing[purchaseOrderMergeKey(line)]
		if !ok {
			line.ID, line.ReceivedQuantity = 0, 0
			plan.inserts = append(plan.inserts, line)
			continue
		}
		merged := current
		merged.Quantity = max(current.Quantity, line.Quantity)
		merged.Note = joinMergedNotes(current.Note, line.Note, "; ")
		if utf8.RuneCountInString(merged.Note) > maxPurchaseOrderItemNote {
			return plan, badRequest(fmt.Sprintf(
				"%s: the two notes together pass the %d character limit; shorten one before merging",
				current.Name,
				maxPurchaseOrderItemNote,
			))
		}
		if merged.Quantity != current.Quantity || merged.Note != current.Note {
			plan.updates = append(plan.updates, purchaseOrderMergedLine{
				item: merged, quantityBefore: current.Quantity, noteBefore: current.Note,
			})
		}
	}
	return plan, nil
}

// MergePurchaseOrder folds a draft into an open order that has received nothing. The open order
// comes back as a draft, so it is approved again before it is sent, and the old draft is closed.
// Only a manager may do it: it undoes the approval of an order that may already be with the
// supplier, which a drafter cannot do. The draft being merged may be anyone's.
func (s *inventoryService) MergePurchaseOrder(
	ctx context.Context, id int, req types.MergePurchaseOrderRequest, userID int,
) error {
	if !canApprovePurchaseOrders(ctx) {
		return httperror.Forbidden("only a manager can merge purchase orders", nil)
	}
	if req.DraftID <= 0 || req.DraftID == id {
		return badRequest("choose a different draft to merge")
	}
	return s.inTx(ctx, "failed to merge purchase orders", func(tx *sqlx.Tx) error {
		// Lock the lower id first so two merges touching the same orders cannot deadlock.
		locked := make(map[int]types.PurchaseOrder, 2)
		for _, orderID := range []int{min(id, req.DraftID), max(id, req.DraftID)} {
			order, err := s.repo.GetPurchaseOrderForUpdateTx(ctx, tx, orderID)
			if errors.Is(err, sql.ErrNoRows) {
				return purchaseOrderNotFound(err)
			}
			if err != nil {
				return httperror.ServerError("failed to fetch purchase order", err)
			}
			locked[orderID] = order
		}
		target, draft := locked[id], locked[req.DraftID]
		if target.Version != req.Version || draft.Version != req.DraftVersion {
			return purchaseOrderChangedElsewhere()
		}
		if target.Status != constant.PurchaseOrderOpen || purchaseOrderHasReceipts(target) {
			return conflict(
				"only an open order with nothing received can have a draft merged into it",
			)
		}
		if draft.Status != constant.PurchaseOrderDraft {
			return conflict("only a draft can be merged into another order")
		}
		plan, err := planPurchaseOrderMerge(target, draft)
		if err != nil {
			return err
		}

		header := types.NewPurchaseOrder{
			SupplierID: target.SupplierID, SupplierName: target.Supplier, Note: plan.note,
		}
		if err = s.repo.UpdatePurchaseOrderHeaderTx(ctx, tx, target.ID, header); err != nil {
			return httperror.ServerError("failed to update purchase order", err)
		}
		for _, line := range plan.updates {
			if err = s.repo.UpdatePurchaseOrderItemTx(ctx, tx, target.ID, line.item); err != nil {
				return purchaseOrderItemError(err)
			}
		}
		for _, line := range plan.inserts {
			if err = s.repo.InsertPurchaseOrderItemTx(ctx, tx, target.ID, line); err != nil {
				return purchaseOrderItemError(err)
			}
		}
		if err = s.repo.RevertPurchaseOrderToDraftTx(ctx, tx, target.ID); err != nil {
			return httperror.ServerError("failed to return the order to a draft", err)
		}
		if err = s.repo.ClosePurchaseOrderTx(ctx, tx, draft.ID); err != nil {
			return httperror.ServerError("failed to close the merged draft", err)
		}

		raised := make([]map[string]any, 0, len(plan.updates))
		lineNotes := make([]map[string]any, 0, len(plan.updates))
		for _, line := range plan.updates {
			if line.item.Quantity != line.quantityBefore {
				raised = append(raised, map[string]any{
					"name":     line.item.Name,
					"quantity": change{From: line.quantityBefore, To: line.item.Quantity},
				})
			}
			if line.item.Note != line.noteBefore {
				lineNotes = append(lineNotes, map[string]any{
					"name": line.item.Name,
					"note": change{From: line.noteBefore, To: line.item.Note},
				})
			}
		}
		onTarget := map[string]any{
			"merged_purchase_order_id": draft.ID,
			"status_before":            target.Status,
			"added":                    auditItemSummaries(plan.inserts),
			"raised":                   raised,
		}
		if plan.note != target.Note {
			onTarget["note"] = change{From: target.Note, To: plan.note}
		}
		if len(lineNotes) > 0 {
			onTarget["line_notes"] = lineNotes
		}
		// The draft's lines now sit on the open order's supplier; keep a trace of who they were
		// written for.
		draftSupplier := strings.TrimSpace(draft.Supplier)
		targetSupplier := strings.TrimSpace(target.Supplier)
		if draftSupplier != "" && !strings.EqualFold(draftSupplier, targetSupplier) {
			onTarget["draft_supplier"] = draftSupplier
		}
		err = s.auditPurchaseOrder(ctx, tx, userID, target.ID, constant.AuditMerged, "", onTarget)
		if err != nil {
			return err
		}
		onDraft := map[string]any{"merged_into_purchase_order_id": target.ID}
		return s.auditPurchaseOrder(ctx, tx, userID, draft.ID, constant.AuditMerged, "", onDraft)
	})
}

// purchaseOrderHasReceipts reports whether any line has received stock.
func purchaseOrderHasReceipts(order types.PurchaseOrder) bool {
	for _, item := range order.Items {
		if item.ReceivedQuantity > 0 {
			return true
		}
	}
	return false
}
