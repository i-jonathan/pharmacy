package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
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
	// The quantity columns are 32-bit integers, so this is the only ceiling. Large orders are
	// the user's call; the editor asks them to confirm.
	maxPurchaseOrderQuantity = math.MaxInt32
	maxPurchaseOrderText     = 200
	maxPurchaseOrderNote     = 1000
	maxPurchaseOrderItemNote = 500

	purchaseOrderViewActive  = "active"
	purchaseOrderViewDrafts  = "drafts"
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

// purchaseOrderActive reports whether an approved order is still waiting for stock. Only those
// can be received against, closed or moved.
func purchaseOrderActive(status string) bool {
	return status == constant.PurchaseOrderOpen || status == constant.PurchaseOrderPartial
}

// purchaseOrderAwaitingApproval reports whether an order is a draft or submitted for approval.
func purchaseOrderAwaitingApproval(status string) bool {
	return status == constant.PurchaseOrderDraft || status == constant.PurchaseOrderPending
}

// purchaseOrderStateConflict is the 409 for an action the order's status does not allow.
func purchaseOrderStateConflict(status string) error {
	message := fmt.Sprintf("this purchase order is already %s", status)
	if purchaseOrderAwaitingApproval(status) {
		message = "this purchase order has not been approved yet"
	}
	return httperror.NewHttpError(http.StatusConflict, message, nil)
}

func canApprovePurchaseOrders(ctx context.Context) bool {
	return HasPermission(ctx, constant.ApprovePurchaseOrderPermissionKey)
}

func canDraftPurchaseOrders(ctx context.Context) bool {
	return HasPermission(ctx, constant.DraftPurchaseOrderPermissionKey)
}

func requireApprover(ctx context.Context) error {
	if !canApprovePurchaseOrders(ctx) {
		return httperror.Forbidden("only a manager can do this", nil)
	}
	return nil
}

// checkPurchaseOrderEditable decides whether this user may change the order. Managers can change
// anything that is not finished; a drafter can change only their own order while it is a draft.
func checkPurchaseOrderEditable(ctx context.Context, order types.PurchaseOrder, userID int) error {
	if !purchaseOrderAwaitingApproval(order.Status) && !purchaseOrderActive(order.Status) {
		return httperror.NewHttpError(
			http.StatusConflict,
			fmt.Sprintf("a %s purchase order cannot be edited", order.Status),
			nil,
		)
	}
	if canApprovePurchaseOrders(ctx) {
		return nil
	}
	own := order.CreatedByID != nil && *order.CreatedByID == userID
	if order.Status == constant.PurchaseOrderDraft && own && canDraftPurchaseOrders(ctx) {
		return nil
	}
	if order.Status == constant.PurchaseOrderPending && own {
		return httperror.NewHttpError(
			http.StatusConflict,
			"this purchase order has been submitted; a manager will review it",
			nil,
		)
	}
	return httperror.Forbidden("you cannot change this purchase order", nil)
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
	seenNew := make(map[string]bool)
	for _, item := range append(append([]types.PurchaseOrderItem{}, plan.updates...), plan.inserts...) {
		if item.ProductID == nil {
			// A new line typed in by name can repeat one the order already has. Kept lines were
			// already checked against each other when the request was normalized.
			key := strings.ToLower(item.Name + "\x00" + item.Manufacturer)
			if seenNew[key] && item.ID == 0 {
				return plan, purchaseOrderBadRequest(
					fmt.Sprintf("%s is listed more than once", item.Name),
				)
			}
			seenNew[key] = true
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
	// A manager's order is approved by definition; anyone else writes a draft.
	status := constant.PurchaseOrderDraft
	switch {
	case canApprovePurchaseOrders(ctx):
		status = constant.PurchaseOrderOpen
	case !canDraftPurchaseOrders(ctx):
		return 0, httperror.Forbidden("you cannot create purchase orders", nil)
	}
	id, err := s.insertPurchaseOrder(ctx, tx, order.Supplier, order.Note, status, nil, userID)
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
		"supplier": order.Supplier, "note": order.Note, "status": status,
		"items": auditItemSummaries(order.Items),
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
	ctx context.Context,
	tx *sqlx.Tx,
	supplierName, note, status string,
	sourceOrderID *int,
	userID int,
) (int, error) {
	supplier, err := s.resolveSupplier(ctx, tx, supplierName)
	if err != nil {
		return 0, err
	}
	header := types.NewPurchaseOrder{
		Status: status, SourceOrderID: sourceOrderID, Note: note, CreatedByID: userID,
	}
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
	if err = checkPurchaseOrderEditable(ctx, stored, userID); err != nil {
		return 0, err
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
	// Changing quantities or adding lines can move an approved order between open, partial and
	// received. Drafts have no such progress.
	if purchaseOrderActive(stored.Status) {
		if err = s.repo.RefreshPurchaseOrderStatusTx(ctx, tx, order.ID); err != nil {
			return 0, httperror.ServerError("failed to update purchase order status", err)
		}
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

// ListPurchaseOrders returns one tab of orders: "drafts" (being written or waiting for approval),
// "history" (received and closed, the newest 100) or, for anything else, "active" (open and
// partial). Managers see everyone's orders; everyone else sees only their own.
func (s *inventoryService) ListPurchaseOrders(
	ctx context.Context, view string, userID int,
) ([]types.PurchaseOrder, error) {
	statuses := []string{constant.PurchaseOrderOpen, constant.PurchaseOrderPartial}
	limit := 0
	switch view {
	case purchaseOrderViewDrafts:
		statuses = []string{constant.PurchaseOrderDraft, constant.PurchaseOrderPending}
	case purchaseOrderViewHistory:
		statuses = []string{constant.PurchaseOrderReceived, constant.PurchaseOrderClosed}
		limit = purchaseOrderHistoryMax
	}
	var createdBy *int
	if !canApprovePurchaseOrders(ctx) {
		createdBy = &userID
	}
	orders, err := s.repo.ListPurchaseOrders(ctx, statuses, limit, createdBy)
	if err != nil {
		return nil, httperror.ServerError("failed to fetch purchase orders", err)
	}
	if orders == nil {
		orders = []types.PurchaseOrder{}
	}
	return orders, nil
}

// visiblePurchaseOrder hides drafts and submitted orders from people who neither wrote them nor
// can approve them. Approved orders are visible to anyone, because anyone can receive against
// one.
func visiblePurchaseOrder(ctx context.Context, order types.PurchaseOrder, userID int) bool {
	if !purchaseOrderAwaitingApproval(order.Status) || canApprovePurchaseOrders(ctx) {
		return true
	}
	return order.CreatedByID != nil && *order.CreatedByID == userID
}

func (s *inventoryService) GetPurchaseOrder(
	ctx context.Context, id, userID int,
) (types.PurchaseOrder, error) {
	order, err := s.repo.GetPurchaseOrder(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return order, purchaseOrderNotFound(err)
	}
	if err != nil {
		return order, httperror.ServerError("failed to fetch purchase order", err)
	}
	if !visiblePurchaseOrder(ctx, order, userID) {
		return types.PurchaseOrder{}, purchaseOrderNotFound(sql.ErrNoRows)
	}
	return order, nil
}

// ClosePurchaseOrder abandons whatever is still outstanding on an order, for example when
// the supplier will not deliver the rest. Stock already received is unaffected.
func (s *inventoryService) ClosePurchaseOrder(
	ctx context.Context, id, version int, reason string, userID int,
) error {
	if err := requireApprover(ctx); err != nil {
		return err
	}
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
	if !purchaseOrderActive(order.Status) {
		return purchaseOrderStateConflict(order.Status)
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
	if err := requireApprover(ctx); err != nil {
		return 0, err
	}
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
	if !purchaseOrderActive(order.Status) {
		return 0, purchaseOrderStateConflict(order.Status)
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

	newID, err := s.insertPurchaseOrder(
		ctx, tx, supplier.Name, "", constant.PurchaseOrderOpen, &order.ID, userID,
	)
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
	if !purchaseOrderActive(order.Status) {
		return purchaseOrderStateConflict(order.Status)
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
