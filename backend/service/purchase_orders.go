package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"pharmacy/httperror"
	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"pharmacy/repository"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/jmoiron/sqlx"
)

const (
	maxPurchaseOrderLines    = 200
	maxPurchaseOrderQuantity = 1_000_000
	maxPurchaseOrderText     = 200
	maxPurchaseOrderNote     = 1000
	maxPurchaseOrderItemNote = 500

	purchaseOrderViewActive  = "active"
	purchaseOrderViewHistory = "history"
	purchaseOrderHistoryMax  = 100
)

// purchaseOrderPlan is the set of line changes that turns an existing order into the
// requested one.
type purchaseOrderPlan struct {
	inserts []types.PurchaseOrderItem
	updates []types.PurchaseOrderItem
	deletes []int
}

func purchaseOrderBadRequest(msg string) error {
	return httperror.BadRequest(msg, errors.New(msg))
}

func purchaseOrderEditable(status string) bool {
	return status == constant.PurchaseOrderOpen || status == constant.PurchaseOrderPartial
}

func purchaseOrderNotFound(err error) error {
	return httperror.NotFound("purchase order was not found", err)
}

// normalizePurchaseOrderRequest trims text and rejects requests that can never be valid,
// independent of what is stored.
func normalizePurchaseOrderRequest(
	order types.PurchaseOrderRequest,
) (types.PurchaseOrderRequest, error) {
	order.Supplier = strings.TrimSpace(order.Supplier)
	if utf8.RuneCountInString(order.Supplier) > maxPurchaseOrderText {
		return order, purchaseOrderBadRequest("supplier name is too long")
	}
	order.Note = strings.TrimSpace(order.Note)
	if utf8.RuneCountInString(order.Note) > maxPurchaseOrderNote {
		return order, purchaseOrderBadRequest(
			fmt.Sprintf("notes are limited to %d characters", maxPurchaseOrderNote),
		)
	}
	if len(order.Items) == 0 {
		return order, purchaseOrderBadRequest("add at least one item")
	}
	if len(order.Items) > maxPurchaseOrderLines {
		return order, purchaseOrderBadRequest("too many items on one purchase order")
	}

	items := make([]types.PurchaseOrderItem, len(order.Items))
	seenProducts := make(map[int]bool)
	seenNewItems := make(map[string]bool)
	for i, item := range order.Items {
		item.Name = strings.TrimSpace(item.Name)
		item.Manufacturer = strings.TrimSpace(item.Manufacturer)
		item.Note = strings.TrimSpace(item.Note)
		if utf8.RuneCountInString(item.Note) > maxPurchaseOrderItemNote {
			return order, purchaseOrderBadRequest(fmt.Sprintf(
				"item %d: notes are limited to %d characters", i+1, maxPurchaseOrderItemNote,
			))
		}
		if item.Name == "" {
			return order, purchaseOrderBadRequest(fmt.Sprintf("item %d needs a name", i+1))
		}
		if utf8.RuneCountInString(item.Name) > maxPurchaseOrderText ||
			utf8.RuneCountInString(item.Manufacturer) > maxPurchaseOrderText {
			return order, purchaseOrderBadRequest(fmt.Sprintf("item %d: text is too long", i+1))
		}
		if item.Quantity < 1 || item.Quantity > maxPurchaseOrderQuantity {
			return order, purchaseOrderBadRequest(
				fmt.Sprintf(
					"item %d: quantity must be between 1 and %d",
					i+1,
					maxPurchaseOrderQuantity,
				),
			)
		}
		if item.ProductID != nil {
			if seenProducts[*item.ProductID] {
				return order, purchaseOrderBadRequest(
					fmt.Sprintf("%s is listed more than once", item.Name),
				)
			}
			seenProducts[*item.ProductID] = true
		} else {
			key := strings.ToLower(item.Name + "\x00" + item.Manufacturer)
			if seenNewItems[key] {
				return order, purchaseOrderBadRequest(
					fmt.Sprintf("%s is listed more than once", item.Name),
				)
			}
			seenNewItems[key] = true
		}
		items[i] = item
	}
	order.Items = items
	return order, nil
}

// planPurchaseOrderEdit compares the stored lines with the requested ones. Lines that have
// already received stock keep their identity (product, name, manufacturer) and can only
// have their quantity changed, never below what has been received, and cannot be removed.
func planPurchaseOrderEdit(
	existing, requested []types.PurchaseOrderItem,
) (purchaseOrderPlan, error) {
	var plan purchaseOrderPlan
	stored := make(map[int]types.PurchaseOrderItem, len(existing))
	for _, item := range existing {
		stored[item.ID] = item
	}

	kept := make(map[int]bool, len(requested))
	for _, item := range requested {
		if item.ID == 0 {
			plan.inserts = append(plan.inserts, item)
			continue
		}
		current, ok := stored[item.ID]
		if !ok || kept[item.ID] {
			return plan, purchaseOrderBadRequest("an item does not belong to this purchase order")
		}
		kept[item.ID] = true
		if item.Quantity < current.ReceivedQuantity {
			return plan, purchaseOrderBadRequest(fmt.Sprintf(
				"%s: quantity cannot be lower than the %d already received",
				current.Name, current.ReceivedQuantity,
			))
		}
		if current.ReceivedQuantity > 0 {
			item.ProductID = current.ProductID
			item.Name = current.Name
			item.Manufacturer = current.Manufacturer
		}
		plan.updates = append(plan.updates, item)
	}

	for _, current := range existing {
		if kept[current.ID] {
			continue
		}
		if current.ReceivedQuantity > 0 {
			return plan, purchaseOrderBadRequest(fmt.Sprintf(
				"%s already has stock received and cannot be removed", current.Name,
			))
		}
		plan.deletes = append(plan.deletes, current.ID)
	}

	// Lines with received stock got their stored product back, so check for duplicates again on
	// what will actually be saved.
	seen := make(map[int]bool)
	for _, item := range append(append([]types.PurchaseOrderItem{}, plan.updates...), plan.inserts...) {
		if item.ProductID == nil {
			continue
		}
		if seen[*item.ProductID] {
			return plan, purchaseOrderBadRequest(
				fmt.Sprintf("%s is listed more than once", item.Name),
			)
		}
		seen[*item.ProductID] = true
	}
	return plan, nil
}

func (s *inventoryService) SavePurchaseOrder(
	ctx context.Context, order types.PurchaseOrderRequest, userID int,
) (int, error) {
	order, err := normalizePurchaseOrderRequest(order)
	if err != nil {
		return 0, err
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return 0, httperror.ServerError("failed to start transaction", err)
	}
	committed := false
	defer func() {
		if !committed {
			s.repo.RollbackTx(tx)
		}
	}()

	var orderID int
	if order.ID == 0 {
		orderID, err = s.createPurchaseOrder(ctx, tx, order, userID)
	} else {
		orderID, err = s.editPurchaseOrder(ctx, tx, order, userID)
	}
	if err != nil {
		return 0, err
	}

	if err = s.repo.CommitTx(tx); err != nil {
		return 0, httperror.ServerError("failed to save purchase order", err)
	}
	committed = true
	return orderID, nil
}

// resolveSupplier finds or creates the supplier for a name. A blank name means no supplier.
func (s *inventoryService) resolveSupplier(
	ctx context.Context, tx *sqlx.Tx, name string,
) (*types.Supplier, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, nil
	}
	if utf8.RuneCountInString(name) > maxPurchaseOrderText {
		return nil, httperror.BadRequest("supplier name is too long", nil)
	}
	supplier, err := s.repo.FindOrCreateSupplierTx(ctx, tx, name)
	if err != nil {
		return nil, httperror.ServerError("failed to resolve supplier", err)
	}
	return &supplier, nil
}

func (s *inventoryService) createPurchaseOrder(
	ctx context.Context, tx *sqlx.Tx, order types.PurchaseOrderRequest, userID int,
) (int, error) {
	id, err := s.insertPurchaseOrder(ctx, tx, order.Supplier, order.Note, nil, userID)
	if err != nil {
		return 0, err
	}
	for _, item := range order.Items {
		item.ID = 0
		if err = s.repo.InsertPurchaseOrderItemTx(ctx, tx, id, item); err != nil {
			return 0, purchaseOrderItemError(err)
		}
	}
	details := map[string]any{
		"supplier": order.Supplier, "note": order.Note, "items": auditItemSummaries(order.Items),
	}
	if err = s.auditPurchaseOrder(
		ctx,
		tx,
		userID,
		id,
		constant.AuditCreated,
		"",
		details,
	); err != nil {
		return 0, err
	}
	return id, nil
}

func (s *inventoryService) insertPurchaseOrder(
	ctx context.Context, tx *sqlx.Tx, supplierName, note string, sourceOrderID *int, userID int,
) (int, error) {
	supplier, err := s.resolveSupplier(ctx, tx, supplierName)
	if err != nil {
		return 0, err
	}
	header := types.NewPurchaseOrder{SourceOrderID: sourceOrderID, Note: note, CreatedByID: userID}
	if supplier != nil {
		header.SupplierID, header.SupplierName = &supplier.ID, supplier.Name
	}
	id, err := s.repo.CreatePurchaseOrderTx(ctx, tx, header)
	if err != nil {
		return 0, httperror.ServerError("failed to create purchase order", err)
	}
	return id, nil
}

func (s *inventoryService) editPurchaseOrder(
	ctx context.Context, tx *sqlx.Tx, order types.PurchaseOrderRequest, userID int,
) (int, error) {
	stored, err := s.repo.GetPurchaseOrderForUpdateTx(ctx, tx, order.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, purchaseOrderNotFound(err)
	}
	if err != nil {
		return 0, httperror.ServerError("failed to fetch purchase order", err)
	}
	// The lock is held, so this is the authoritative check against a stale copy.
	if order.Version != stored.Version {
		return 0, purchaseOrderChangedElsewhere()
	}
	if !purchaseOrderEditable(stored.Status) {
		return 0, httperror.NewHttpError(
			http.StatusConflict,
			fmt.Sprintf("a %s purchase order cannot be edited", stored.Status),
			nil,
		)
	}

	plan, err := planPurchaseOrderEdit(stored.Items, order.Items)
	if err != nil {
		return 0, err
	}

	header := types.NewPurchaseOrder{Note: order.Note}
	if stored.SupplierID != nil && strings.EqualFold(order.Supplier, stored.Supplier) {
		// Same supplier as before, perhaps retyped in another case: keep it, create nothing.
		header.SupplierID, header.SupplierName = stored.SupplierID, stored.Supplier
	} else {
		supplier, err := s.resolveSupplier(ctx, tx, order.Supplier)
		if err != nil {
			return 0, err
		}
		if supplier != nil {
			header.SupplierID, header.SupplierName = &supplier.ID, supplier.Name
		}
	}
	changes := describePurchaseOrderEdit(stored, header, plan)
	if changes == nil {
		return order.ID, nil
	}
	if err = s.repo.UpdatePurchaseOrderHeaderTx(ctx, tx, order.ID, header); err != nil {
		return 0, httperror.ServerError("failed to update purchase order", err)
	}
	for _, itemID := range plan.deletes {
		if err = s.repo.DeletePurchaseOrderItemTx(ctx, tx, order.ID, itemID); err != nil {
			return 0, httperror.ServerError("failed to remove purchase order item", err)
		}
	}
	for _, item := range plan.updates {
		if err = s.repo.UpdatePurchaseOrderItemTx(ctx, tx, order.ID, item); err != nil {
			return 0, purchaseOrderItemError(err)
		}
	}
	for _, item := range plan.inserts {
		if err = s.repo.InsertPurchaseOrderItemTx(ctx, tx, order.ID, item); err != nil {
			return 0, purchaseOrderItemError(err)
		}
	}
	// Changing quantities or adding lines can move an order between open, partial and received.
	if err = s.repo.RefreshPurchaseOrderStatusTx(ctx, tx, order.ID); err != nil {
		return 0, httperror.ServerError("failed to update purchase order status", err)
	}
	err = s.auditPurchaseOrder(ctx, tx, userID, order.ID, constant.AuditEdited, "", changes)
	if err != nil {
		return 0, err
	}
	return order.ID, nil
}

func purchaseOrderItemError(err error) error {
	if errors.Is(err, repository.ErrInvalidReference) {
		return purchaseOrderBadRequest("an item refers to a product that no longer exists")
	}
	return httperror.ServerError("failed to save purchase order item", err)
}

// ListPurchaseOrders returns the most recent "history" (received and closed) orders for the
// "history" view and the "active" (open and partial) orders for anything else.
func (s *inventoryService) ListPurchaseOrders(
	ctx context.Context, view string,
) ([]types.PurchaseOrder, error) {
	// Anything but "history" is the active list, so a missing or mistyped view can never load
	// every order ever created.
	statuses := []string{constant.PurchaseOrderOpen, constant.PurchaseOrderPartial}
	limit := 0
	if view == purchaseOrderViewHistory {
		statuses = []string{constant.PurchaseOrderReceived, constant.PurchaseOrderClosed}
		limit = purchaseOrderHistoryMax
	}
	orders, err := s.repo.ListPurchaseOrders(ctx, statuses, limit)
	if err != nil {
		return nil, httperror.ServerError("failed to fetch purchase orders", err)
	}
	if orders == nil {
		orders = []types.PurchaseOrder{}
	}
	return orders, nil
}

func (s *inventoryService) GetPurchaseOrder(
	ctx context.Context, id int,
) (types.PurchaseOrder, error) {
	order, err := s.repo.GetPurchaseOrder(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return order, purchaseOrderNotFound(err)
	}
	if err != nil {
		return order, httperror.ServerError("failed to fetch purchase order", err)
	}
	return order, nil
}

// ClosePurchaseOrder abandons whatever is still outstanding on an order, for example when
// the supplier will not deliver the rest. Stock already received is unaffected.
func (s *inventoryService) ClosePurchaseOrder(
	ctx context.Context, id, version int, reason string, userID int,
) error {
	reason, err := normalizeAuditReason(reason)
	if err != nil {
		return err
	}
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
	// The user chose to close what they saw, so refuse if it has changed since.
	if order.Version != version {
		return purchaseOrderChangedElsewhere()
	}
	if !purchaseOrderEditable(order.Status) {
		return httperror.NewHttpError(
			http.StatusConflict,
			fmt.Sprintf("this purchase order is already %s", order.Status),
			nil,
		)
	}
	if err = s.repo.ClosePurchaseOrderTx(ctx, tx, id); err != nil {
		return httperror.ServerError("failed to close purchase order", err)
	}
	details := map[string]any{
		"status_before": order.Status,
		"abandoned":     auditItemSummaries(outstandingPurchaseOrderItems(order.Items)),
	}
	if err = s.auditPurchaseOrder(
		ctx,
		tx,
		userID,
		id,
		constant.AuditClosed,
		reason,
		details,
	); err != nil {
		return err
	}
	if err = s.repo.CommitTx(tx); err != nil {
		return httperror.ServerError("failed to close purchase order", err)
	}
	committed = true
	return nil
}

// outstandingPurchaseOrderItems returns fresh, unreceived lines for whatever has not arrived
// yet, ready to go on a new order.
func outstandingPurchaseOrderItems(items []types.PurchaseOrderItem) []types.PurchaseOrderItem {
	var outstanding []types.PurchaseOrderItem
	for _, item := range items {
		if remaining := item.Quantity - item.ReceivedQuantity; remaining > 0 {
			outstanding = append(outstanding, types.PurchaseOrderItem{
				ProductID:    item.ProductID,
				Name:         item.Name,
				Manufacturer: item.Manufacturer,
				Note:         item.Note,
				Quantity:     remaining,
			})
		}
	}
	return outstanding
}

// MovePurchaseOrderOutstanding takes whatever a supplier has not delivered and puts it on a
// new order for a different supplier. The original order is closed, keeping what it did
// receive, and the new order records where its items came from. It returns the new order id.
func (s *inventoryService) MovePurchaseOrderOutstanding(
	ctx context.Context, id, version int, supplierName, reason string, userID int,
) (int, error) {
	if strings.TrimSpace(supplierName) == "" {
		return 0, purchaseOrderBadRequest("choose the supplier for the outstanding items")
	}
	reason, err := normalizeAuditReason(reason)
	if err != nil {
		return 0, err
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return 0, httperror.ServerError("failed to start transaction", err)
	}
	committed := false
	defer func() {
		if !committed {
			s.repo.RollbackTx(tx)
		}
	}()

	order, err := s.repo.GetPurchaseOrderForUpdateTx(ctx, tx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, purchaseOrderNotFound(err)
	}
	if err != nil {
		return 0, httperror.ServerError("failed to fetch purchase order", err)
	}
	if order.Version != version {
		return 0, purchaseOrderChangedElsewhere()
	}
	if !purchaseOrderEditable(order.Status) {
		return 0, httperror.NewHttpError(
			http.StatusConflict,
			fmt.Sprintf("this purchase order is already %s", order.Status),
			nil,
		)
	}
	items := outstandingPurchaseOrderItems(order.Items)
	if len(items) == 0 {
		return 0, httperror.NewHttpError(
			http.StatusConflict, "nothing is outstanding on this purchase order", nil,
		)
	}

	supplier, err := s.resolveSupplier(ctx, tx, supplierName)
	if err != nil {
		return 0, err
	}
	if order.SupplierID != nil && *order.SupplierID == supplier.ID {
		return 0, purchaseOrderBadRequest("choose a different supplier from the one on this order")
	}

	newID, err := s.insertPurchaseOrder(ctx, tx, supplier.Name, "", &order.ID, userID)
	if err != nil {
		return 0, err
	}
	for _, item := range items {
		if err = s.repo.InsertPurchaseOrderItemTx(ctx, tx, newID, item); err != nil {
			return 0, purchaseOrderItemError(err)
		}
	}
	if err = s.repo.ClosePurchaseOrderTx(ctx, tx, order.ID); err != nil {
		return 0, httperror.ServerError("failed to close purchase order", err)
	}
	moved := map[string]any{
		"to_purchase_order_id": newID,
		"supplier":             change{From: order.Supplier, To: supplier.Name},
		"items":                auditItemSummaries(items),
	}
	err = s.auditPurchaseOrder(ctx, tx, userID, order.ID, constant.AuditMoved, reason, moved)
	if err != nil {
		return 0, err
	}
	created := map[string]any{
		"moved_from_purchase_order_id": order.ID,
		"supplier":                     supplier.Name,
		"items":                        auditItemSummaries(items),
	}
	err = s.auditPurchaseOrder(ctx, tx, userID, newID, constant.AuditCreated, reason, created)
	if err != nil {
		return 0, err
	}
	if err = s.repo.CommitTx(tx); err != nil {
		return 0, httperror.ServerError("failed to move outstanding items", err)
	}
	committed = true
	return newID, nil
}

// purchaseOrderReceipt is the quantity of one purchase order line arriving in a receipt.
type purchaseOrderReceipt struct {
	productID int
	quantity  int
}

// groupPurchaseOrderReceipts totals the receipt per purchase order line, so a line split
// over several rows (for example different expiry dates) is checked against its outstanding
// quantity as a whole.
func groupPurchaseOrderReceipts(
	products []types.ReceiveItem,
) (map[int]purchaseOrderReceipt, error) {
	receipts := make(map[int]purchaseOrderReceipt)
	for _, item := range products {
		if item.PurchaseOrderItemID == 0 {
			continue
		}
		receipt := receipts[item.PurchaseOrderItemID]
		if receipt.productID != 0 && receipt.productID != item.ID {
			return nil, purchaseOrderBadRequest(
				"a purchase order item was received as two different products",
			)
		}
		receipt.productID = item.ID
		receipt.quantity += item.Quantity
		receipts[item.PurchaseOrderItemID] = receipt
	}
	return receipts, nil
}

// applyPurchaseOrderReceipt records a receipt against a purchase order inside the receipt's
// own transaction, so stock and order progress commit or roll back together.
func (s *inventoryService) applyPurchaseOrderReceipt(
	ctx context.Context, tx *sqlx.Tx, params types.ReceiveSupplyRequest, receivingBatchID int,
) error {
	receipts, err := groupPurchaseOrderReceipts(params.Products)
	if err != nil {
		return err
	}
	// A receipt with no lines on the order is an ordinary receipt, whatever the order's state.
	if len(receipts) == 0 {
		return nil
	}
	if params.PurchaseOrderID == 0 {
		return purchaseOrderBadRequest("purchase order is required to receive its items")
	}

	order, err := s.repo.GetPurchaseOrderForUpdateTx(ctx, tx, params.PurchaseOrderID)
	if errors.Is(err, sql.ErrNoRows) {
		return purchaseOrderNotFound(err)
	}
	if err != nil {
		return httperror.ServerError("failed to fetch purchase order", err)
	}
	if !purchaseOrderEditable(order.Status) {
		return httperror.NewHttpError(
			http.StatusConflict,
			fmt.Sprintf("this purchase order is already %s", order.Status),
			nil,
		)
	}

	lines := make(map[int]types.PurchaseOrderItem, len(order.Items))
	for _, item := range order.Items {
		lines[item.ID] = item
	}
	itemIDs := make([]int, 0, len(receipts))
	for itemID := range receipts {
		itemIDs = append(itemIDs, itemID)
	}
	sort.Ints(itemIDs)

	received := make([]map[string]any, 0, len(receipts))
	for _, itemID := range itemIDs {
		receipt := receipts[itemID]
		line, ok := lines[itemID]
		if !ok {
			return purchaseOrderBadRequest("an item does not belong to this purchase order")
		}
		if line.ProductID != nil && *line.ProductID != receipt.productID {
			return purchaseOrderBadRequest(
				fmt.Sprintf("%s was received as a different product", line.Name),
			)
		}
		if outstanding := line.Quantity - line.ReceivedQuantity; receipt.quantity > outstanding {
			return purchaseOrderBadRequest(fmt.Sprintf(
				"%s: only %d still outstanding on this purchase order", line.Name, outstanding,
			))
		}
		err = s.repo.ReceivePurchaseOrderItemTx(
			ctx, tx, order.ID, itemID, receipt.productID, receipt.quantity,
		)
		if errors.Is(err, repository.ErrPurchaseOrderItemNotReceivable) {
			return purchaseOrderBadRequest(
				fmt.Sprintf("%s: quantity no longer fits this purchase order", line.Name),
			)
		}
		if err != nil {
			log.Println(err)
			return httperror.ServerError("failed to update purchase order", err)
		}
		received = append(received, map[string]any{
			"purchase_order_item_id": itemID,
			"item":                   line.Name,
			"product_id":             receipt.productID,
			"product_linked":         line.ProductID == nil,
			"quantity":               receipt.quantity,
			"outstanding_after":      line.Quantity - line.ReceivedQuantity - receipt.quantity,
		})
	}
	if err = s.repo.RefreshPurchaseOrderStatusTx(ctx, tx, order.ID); err != nil {
		log.Println(err)
		return httperror.ServerError("failed to update purchase order status", err)
	}
	details := map[string]any{
		"receiving_batch_id": receivingBatchID,
		"supplier":           params.Supplier,
		"lines":              received,
	}
	return s.auditPurchaseOrder(
		ctx, tx, params.UserID, order.ID, constant.AuditReceived, "", details,
	)
}
