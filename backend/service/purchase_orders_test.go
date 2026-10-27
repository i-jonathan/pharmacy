package service

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"pharmacy/httperror"
	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"pharmacy/repository"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
)

func intPtr(v int) *int { return &v }

func requireHTTPStatus(t *testing.T, err error, code int) {
	t.Helper()
	var httperr *httperror.HTTPError
	if !errors.As(err, &httperr) {
		t.Fatalf("expected an HTTPError with status %d, got %v", code, err)
	}
	if httperr.Code != code {
		t.Fatalf("expected status %d, got %d (%s)", code, httperr.Code, httperr.Message)
	}
}

func TestNormalizePurchaseOrderRequest(t *testing.T) {
	valid := types.PurchaseOrderItem{Name: "  Amoxil  ", Manufacturer: " GSK ", Quantity: 5}

	t.Run("trims text", func(t *testing.T) {
		got, err := normalizePurchaseOrderRequest(types.PurchaseOrderRequest{
			Supplier: "  Emzor ", Items: []types.PurchaseOrderItem{valid},
		})
		if err != nil {
			t.Fatal(err)
		}
		if got.Supplier != "Emzor" || got.Items[0].Name != "Amoxil" ||
			got.Items[0].Manufacturer != "GSK" {
			t.Fatalf("text was not trimmed: %+v", got)
		}
	})

	cases := map[string][]types.PurchaseOrderItem{
		"no items":      nil,
		"blank name":    {{Name: "  ", Quantity: 1}},
		"zero quantity": {{Name: "A", Quantity: 0}},
		"huge quantity": {{Name: "A", Quantity: maxPurchaseOrderQuantity + 1}},
		"duplicate product": {
			{ProductID: intPtr(1), Name: "A", Quantity: 1},
			{ProductID: intPtr(1), Name: "A", Quantity: 2},
		},
		"duplicate new item": {
			{Name: "Brand New", Manufacturer: "X", Quantity: 1},
			{Name: "brand new", Manufacturer: "x", Quantity: 1},
		},
	}
	for name, items := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := normalizePurchaseOrderRequest(types.PurchaseOrderRequest{Items: items})
			requireHTTPStatus(t, err, http.StatusBadRequest)
		})
	}

	t.Run("same name for different manufacturers is allowed", func(t *testing.T) {
		_, err := normalizePurchaseOrderRequest(
			types.PurchaseOrderRequest{Items: []types.PurchaseOrderItem{
				{Name: "Paracetamol", Manufacturer: "A", Quantity: 1},
				{Name: "Paracetamol", Manufacturer: "B", Quantity: 1},
			}},
		)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestPlanPurchaseOrderEdit(t *testing.T) {
	existing := []types.PurchaseOrderItem{
		{ID: 1, ProductID: intPtr(10), Name: "Received", Quantity: 10, ReceivedQuantity: 4},
		{ID: 2, ProductID: intPtr(11), Name: "Untouched", Quantity: 5},
		{ID: 3, Name: "New item", Quantity: 3},
	}

	t.Run("adds, updates and removes untouched lines", func(t *testing.T) {
		plan, err := planPurchaseOrderEdit(existing, []types.PurchaseOrderItem{
			{ID: 1, ProductID: intPtr(99), Name: "Renamed", Quantity: 12},
			{ID: 2, ProductID: intPtr(11), Name: "Untouched", Quantity: 8},
			{Name: "Brand new line", Quantity: 1},
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(plan.inserts) != 1 || len(plan.updates) != 2 || len(plan.deletes) != 1 ||
			plan.deletes[0] != 3 {
			t.Fatalf("unexpected plan: %+v", plan)
		}
		received := plan.updates[0]
		if received.Name != "Received" || *received.ProductID != 10 || received.Quantity != 12 {
			t.Fatalf("a line with received stock should keep its identity: %+v", received)
		}
	})

	t.Run("quantity can drop to exactly what was received", func(t *testing.T) {
		_, err := planPurchaseOrderEdit(existing, []types.PurchaseOrderItem{
			{ID: 1, Quantity: 4}, {ID: 2, Quantity: 5}, {ID: 3, Quantity: 3},
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("quantity below received is rejected", func(t *testing.T) {
		_, err := planPurchaseOrderEdit(existing, []types.PurchaseOrderItem{
			{ID: 1, Quantity: 3}, {ID: 2, Quantity: 5}, {ID: 3, Quantity: 3},
		})
		requireHTTPStatus(t, err, http.StatusBadRequest)
	})

	t.Run("removing a line with received stock is rejected", func(t *testing.T) {
		_, err := planPurchaseOrderEdit(existing, []types.PurchaseOrderItem{
			{ID: 2, Quantity: 5}, {ID: 3, Quantity: 3},
		})
		requireHTTPStatus(t, err, http.StatusBadRequest)
	})

	t.Run("lines from another order or repeated lines are rejected", func(t *testing.T) {
		_, err := planPurchaseOrderEdit(existing, []types.PurchaseOrderItem{
			{ID: 1, Quantity: 5}, {ID: 2, Quantity: 5}, {ID: 3, Quantity: 3}, {ID: 77, Quantity: 1},
		})
		requireHTTPStatus(t, err, http.StatusBadRequest)
		_, err = planPurchaseOrderEdit(existing, []types.PurchaseOrderItem{
			{ID: 1, Quantity: 5}, {ID: 1, Quantity: 6}, {ID: 2, Quantity: 5}, {ID: 3, Quantity: 3},
		})
		requireHTTPStatus(t, err, http.StatusBadRequest)
	})
}

func TestGroupPurchaseOrderReceipts(t *testing.T) {
	t.Run("sums rows of one line and ignores plain rows", func(t *testing.T) {
		got, err := groupPurchaseOrderReceipts([]types.ReceiveItem{
			{ID: 10, PurchaseOrderItemID: 1, Quantity: 3},
			{ID: 10, PurchaseOrderItemID: 1, Quantity: 4},
			{ID: 11, Quantity: 9},
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 || got[1].quantity != 7 || got[1].productID != 10 {
			t.Fatalf("unexpected receipts: %+v", got)
		}
	})

	t.Run("one line received as two products is rejected", func(t *testing.T) {
		_, err := groupPurchaseOrderReceipts([]types.ReceiveItem{
			{ID: 10, PurchaseOrderItemID: 1, Quantity: 1},
			{ID: 11, PurchaseOrderItemID: 1, Quantity: 1},
		})
		requireHTTPStatus(t, err, http.StatusBadRequest)
	})
}

type fakePurchaseOrderRepo struct {
	repository.PharmacyRepository

	order       types.PurchaseOrder
	lookupErr   error
	received    map[int]int
	refreshed   int
	closed      int
	rollbacks   int
	commits     int
	insertCalls int

	suppliers      map[string]types.Supplier
	created        []types.NewPurchaseOrder
	insertedLines  []types.PurchaseOrderItem
	supplierSaved  *int
	supplierSavedN string
	noteSaved      string
	audit          []types.AuditEntry
}

func (r *fakePurchaseOrderRepo) FindOrCreateSupplierTx(
	_ context.Context, _ *sqlx.Tx, name string,
) (types.Supplier, error) {
	key := strings.ToLower(strings.TrimSpace(name))
	if supplier, ok := r.suppliers[key]; ok {
		return supplier, nil
	}
	if r.suppliers == nil {
		r.suppliers = map[string]types.Supplier{}
	}
	supplier := types.Supplier{ID: len(r.suppliers) + 1, Name: strings.TrimSpace(name)}
	r.suppliers[key] = supplier
	return supplier, nil
}

func (r *fakePurchaseOrderRepo) CreatePurchaseOrderTx(
	_ context.Context, _ *sqlx.Tx, order types.NewPurchaseOrder,
) (int, error) {
	r.created = append(r.created, order)
	return 100 + len(r.created), nil
}

func (r *fakePurchaseOrderRepo) UpdatePurchaseOrderHeaderTx(
	_ context.Context, _ *sqlx.Tx, _ int, header types.NewPurchaseOrder,
) error {
	r.supplierSaved, r.supplierSavedN, r.noteSaved = header.SupplierID, header.SupplierName, header.Note
	return nil
}

func (r *fakePurchaseOrderRepo) BeginTx(context.Context) (*sqlx.Tx, error) { return nil, nil }
func (r *fakePurchaseOrderRepo) RollbackTx(*sqlx.Tx)                       { r.rollbacks++ }

func (r *fakePurchaseOrderRepo) CommitTx(
	*sqlx.Tx,
) error {
	r.commits++
	return nil
}

func (r *fakePurchaseOrderRepo) GetPurchaseOrderForUpdateTx(
	context.Context, *sqlx.Tx, int,
) (types.PurchaseOrder, error) {
	return r.order, r.lookupErr
}

func (r *fakePurchaseOrderRepo) ReceivePurchaseOrderItemTx(
	_ context.Context, _ *sqlx.Tx, _, itemID, _, quantity int,
) error {
	if r.received == nil {
		r.received = map[int]int{}
	}
	r.received[itemID] += quantity
	return nil
}

func (r *fakePurchaseOrderRepo) RefreshPurchaseOrderStatusTx(context.Context, *sqlx.Tx, int) error {
	r.refreshed++
	return nil
}

func (r *fakePurchaseOrderRepo) UpdatePurchaseOrderItemTx(
	context.Context, *sqlx.Tx, int, types.PurchaseOrderItem,
) error {
	return nil
}

func (r *fakePurchaseOrderRepo) DeletePurchaseOrderItemTx(
	context.Context,
	*sqlx.Tx,
	int,
	int,
) error {
	return nil
}

func (r *fakePurchaseOrderRepo) InsertAuditLogTx(
	_ context.Context, _ *sqlx.Tx, entry types.AuditEntry,
) error {
	r.audit = append(r.audit, entry)
	return nil
}

func (r *fakePurchaseOrderRepo) ClosePurchaseOrderTx(context.Context, *sqlx.Tx, int) error {
	r.closed++
	return nil
}

func (r *fakePurchaseOrderRepo) InsertPurchaseOrderItemTx(
	_ context.Context, _ *sqlx.Tx, _ int, item types.PurchaseOrderItem,
) error {
	r.insertCalls++
	r.insertedLines = append(r.insertedLines, item)
	return nil
}

func newOrder(status string) types.PurchaseOrder {
	return types.PurchaseOrder{ID: 1, Status: status, Version: 1, Items: []types.PurchaseOrderItem{
		{ID: 1, ProductID: intPtr(10), Name: "Known", Quantity: 10, ReceivedQuantity: 6},
		{ID: 2, Name: "Brand new", Quantity: 5},
	}}
}

func TestApplyPurchaseOrderReceipt(t *testing.T) {
	apply := func(repo *fakePurchaseOrderRepo, params types.ReceiveSupplyRequest) error {
		return NewInventoryService(
			repo,
		).applyPurchaseOrderReceipt(context.Background(), nil, params, 9)
	}

	t.Run("partial delivery updates the line and refreshes the order", func(t *testing.T) {
		repo := &fakePurchaseOrderRepo{order: newOrder(constant.PurchaseOrderPartial)}
		err := apply(
			repo,
			types.ReceiveSupplyRequest{PurchaseOrderID: 1, Products: []types.ReceiveItem{
				{ID: 10, PurchaseOrderItemID: 1, Quantity: 4},
				{ID: 55, PurchaseOrderItemID: 2, Quantity: 2},
				{ID: 77, Quantity: 3},
			}},
		)
		if err != nil {
			t.Fatal(err)
		}
		if repo.received[1] != 4 || repo.received[2] != 2 || len(repo.received) != 2 ||
			repo.refreshed != 1 {
			t.Fatalf("unexpected updates: %+v refreshed=%d", repo.received, repo.refreshed)
		}
	})

	t.Run("more than outstanding is rejected, even when split over rows", func(t *testing.T) {
		repo := &fakePurchaseOrderRepo{order: newOrder(constant.PurchaseOrderPartial)}
		err := apply(
			repo,
			types.ReceiveSupplyRequest{PurchaseOrderID: 1, Products: []types.ReceiveItem{
				{ID: 10, PurchaseOrderItemID: 1, Quantity: 2},
				{ID: 10, PurchaseOrderItemID: 1, Quantity: 3},
			}},
		)
		requireHTTPStatus(t, err, http.StatusBadRequest)
		if len(repo.received) != 0 {
			t.Fatal("nothing should be recorded when validation fails")
		}
	})

	t.Run("a known line received as another product is rejected", func(t *testing.T) {
		repo := &fakePurchaseOrderRepo{order: newOrder(constant.PurchaseOrderOpen)}
		err := apply(
			repo,
			types.ReceiveSupplyRequest{PurchaseOrderID: 1, Products: []types.ReceiveItem{
				{ID: 99, PurchaseOrderItemID: 1, Quantity: 1},
			}},
		)
		requireHTTPStatus(t, err, http.StatusBadRequest)
	})

	t.Run("closed and received orders conflict", func(t *testing.T) {
		for _, status := range []string{constant.PurchaseOrderClosed, constant.PurchaseOrderReceived} {
			repo := &fakePurchaseOrderRepo{order: newOrder(status)}
			err := apply(
				repo,
				types.ReceiveSupplyRequest{PurchaseOrderID: 1, Products: []types.ReceiveItem{
					{ID: 10, PurchaseOrderItemID: 1, Quantity: 1},
				}},
			)
			requireHTTPStatus(t, err, http.StatusConflict)
		}
	})

	t.Run("unknown order is not found", func(t *testing.T) {
		repo := &fakePurchaseOrderRepo{lookupErr: sql.ErrNoRows}
		err := apply(
			repo,
			types.ReceiveSupplyRequest{PurchaseOrderID: 5, Products: []types.ReceiveItem{
				{ID: 10, PurchaseOrderItemID: 1, Quantity: 1},
			}},
		)
		requireHTTPStatus(t, err, http.StatusNotFound)
	})

	t.Run("line ids without an order id are rejected", func(t *testing.T) {
		repo := &fakePurchaseOrderRepo{}
		err := apply(repo, types.ReceiveSupplyRequest{Products: []types.ReceiveItem{
			{ID: 10, PurchaseOrderItemID: 1, Quantity: 1},
		}})
		requireHTTPStatus(t, err, http.StatusBadRequest)
	})

	t.Run("a receipt without purchase order data does nothing", func(t *testing.T) {
		repo := &fakePurchaseOrderRepo{}
		if err := apply(
			repo,
			types.ReceiveSupplyRequest{Products: []types.ReceiveItem{{ID: 10, Quantity: 1}}},
		); err != nil {
			t.Fatal(err)
		}
		if repo.refreshed != 0 {
			t.Fatal("order should not be touched")
		}
	})
}

func TestSaveAndClosePurchaseOrder(t *testing.T) {
	item := types.PurchaseOrderItem{Name: "Thing", Quantity: 1}

	t.Run("editing a received order conflicts and rolls back", func(t *testing.T) {
		repo := &fakePurchaseOrderRepo{order: newOrder(constant.PurchaseOrderReceived)}
		_, err := NewInventoryService(repo).SavePurchaseOrder(
			context.Background(),
			types.PurchaseOrderRequest{
				ID:      1,
				Version: 1,
				Items:   []types.PurchaseOrderItem{item},
			}, 1,
		)
		requireHTTPStatus(t, err, http.StatusConflict)
		if repo.rollbacks != 1 || repo.commits != 0 {
			t.Fatalf("expected one rollback and no commit, got %d/%d", repo.rollbacks, repo.commits)
		}
	})

	t.Run("editing an unknown order is not found", func(t *testing.T) {
		repo := &fakePurchaseOrderRepo{lookupErr: sql.ErrNoRows}
		_, err := NewInventoryService(repo).SavePurchaseOrder(
			context.Background(),
			types.PurchaseOrderRequest{
				ID:      9,
				Version: 1,
				Items:   []types.PurchaseOrderItem{item},
			}, 1,
		)
		requireHTTPStatus(t, err, http.StatusNotFound)
	})

	t.Run("invalid requests never open a transaction", func(t *testing.T) {
		repo := &fakePurchaseOrderRepo{}
		_, err := NewInventoryService(repo).SavePurchaseOrder(
			context.Background(), types.PurchaseOrderRequest{}, 1,
		)
		requireHTTPStatus(t, err, http.StatusBadRequest)
		if repo.rollbacks != 0 || repo.commits != 0 {
			t.Fatal("no transaction expected")
		}
	})

	t.Run("closing", func(t *testing.T) {
		repo := &fakePurchaseOrderRepo{order: newOrder(constant.PurchaseOrderPartial)}
		if err := NewInventoryService(
			repo,
		).ClosePurchaseOrder(context.Background(), 1, 1, "supplier cannot deliver", 1); err != nil {
			t.Fatal(err)
		}
		if repo.closed != 1 || repo.commits != 1 {
			t.Fatalf("expected close and commit, got %d/%d", repo.closed, repo.commits)
		}

		repo = &fakePurchaseOrderRepo{order: newOrder(constant.PurchaseOrderReceived)}
		err := NewInventoryService(
			repo,
		).ClosePurchaseOrder(context.Background(), 1, 1, "supplier cannot deliver", 1)
		requireHTTPStatus(t, err, http.StatusConflict)
		if !strings.Contains(err.Error(), "received") {
			t.Fatalf("error should name the status: %v", err)
		}
	})
}

func TestOutstandingPurchaseOrderItems(t *testing.T) {
	got := outstandingPurchaseOrderItems([]types.PurchaseOrderItem{
		{
			ID:               1,
			ProductID:        intPtr(10),
			Name:             "Partly",
			Manufacturer:     "M",
			Quantity:         10,
			ReceivedQuantity: 4,
		},
		{ID: 2, Name: "Done", Quantity: 5, ReceivedQuantity: 5},
		{ID: 3, Name: "Untouched", Quantity: 7},
	})
	if len(got) != 2 {
		t.Fatalf("expected 2 outstanding lines, got %+v", got)
	}
	if got[0].ID != 0 || got[0].Quantity != 6 || got[0].ReceivedQuantity != 0 ||
		*got[0].ProductID != 10 ||
		got[0].Manufacturer != "M" {
		t.Fatalf("unexpected first line: %+v", got[0])
	}
	if got[1].Name != "Untouched" || got[1].Quantity != 7 {
		t.Fatalf("unexpected second line: %+v", got[1])
	}
}

func TestMovePurchaseOrderOutstanding(t *testing.T) {
	ctx := context.Background()

	t.Run("moves unfilled items to a new order and closes the original", func(t *testing.T) {
		order := newOrder(constant.PurchaseOrderPartial)
		order.SupplierID = intPtr(1)
		repo := &fakePurchaseOrderRepo{order: order}
		repo.suppliers = map[string]types.Supplier{"emzor": {ID: 1, Name: "Emzor"}}

		newID, err := NewInventoryService(
			repo,
		).MovePurchaseOrderOutstanding(ctx, 1, 1, "  Fidson ", "supplier out of stock", 7)
		if err != nil {
			t.Fatal(err)
		}
		if newID != 101 || len(repo.created) != 1 {
			t.Fatalf("expected one new order 101, got %d / %+v", newID, repo.created)
		}
		header := repo.created[0]
		if header.SupplierName != "Fidson" || header.SupplierID == nil || *header.SupplierID != 2 ||
			header.SourceOrderID == nil || *header.SourceOrderID != 1 || header.CreatedByID != 7 {
			t.Fatalf("unexpected header: %+v", header)
		}
		if len(repo.insertedLines) != 2 || repo.insertedLines[0].Quantity != 4 ||
			repo.insertedLines[1].Quantity != 5 {
			t.Fatalf("unexpected lines: %+v", repo.insertedLines)
		}
		if repo.closed != 1 || repo.commits != 1 {
			t.Fatalf("expected close and commit, got %d/%d", repo.closed, repo.commits)
		}
	})

	t.Run("the same supplier is rejected, ignoring case", func(t *testing.T) {
		order := newOrder(constant.PurchaseOrderPartial)
		order.SupplierID = intPtr(1)
		repo := &fakePurchaseOrderRepo{
			order:     order,
			suppliers: map[string]types.Supplier{"emzor": {ID: 1, Name: "Emzor"}},
		}
		_, err := NewInventoryService(
			repo,
		).MovePurchaseOrderOutstanding(ctx, 1, 1, "EMZOR", "supplier out of stock", 1)
		requireHTTPStatus(t, err, http.StatusBadRequest)
		if len(repo.created) != 0 || repo.commits != 0 {
			t.Fatal("nothing should be written")
		}
	})

	t.Run("a supplier is required", func(t *testing.T) {
		repo := &fakePurchaseOrderRepo{order: newOrder(constant.PurchaseOrderPartial)}
		_, err := NewInventoryService(
			repo,
		).MovePurchaseOrderOutstanding(ctx, 1, 1, "  ", "supplier out of stock", 1)
		requireHTTPStatus(t, err, http.StatusBadRequest)
	})

	t.Run("finished, closed and missing orders", func(t *testing.T) {
		for _, status := range []string{constant.PurchaseOrderReceived, constant.PurchaseOrderClosed} {
			repo := &fakePurchaseOrderRepo{order: newOrder(status)}
			_, err := NewInventoryService(
				repo,
			).MovePurchaseOrderOutstanding(ctx, 1, 1, "Fidson", "supplier out of stock", 1)
			requireHTTPStatus(t, err, http.StatusConflict)
		}
		repo := &fakePurchaseOrderRepo{lookupErr: sql.ErrNoRows}
		_, err := NewInventoryService(
			repo,
		).MovePurchaseOrderOutstanding(ctx, 1, 1, "Fidson", "supplier out of stock", 1)
		requireHTTPStatus(t, err, http.StatusNotFound)
	})

	t.Run("an order with nothing outstanding conflicts", func(t *testing.T) {
		order := newOrder(constant.PurchaseOrderPartial)
		for i := range order.Items {
			order.Items[i].ReceivedQuantity = order.Items[i].Quantity
		}
		repo := &fakePurchaseOrderRepo{order: order}
		_, err := NewInventoryService(
			repo,
		).MovePurchaseOrderOutstanding(ctx, 1, 1, "Fidson", "supplier out of stock", 1)
		requireHTTPStatus(t, err, http.StatusConflict)
	})
}

func TestSavePurchaseOrderResolvesSupplier(t *testing.T) {
	ctx := context.Background()
	items := []types.PurchaseOrderItem{{Name: "Thing", Quantity: 1}}

	t.Run("a typed name becomes a supplier on create", func(t *testing.T) {
		repo := &fakePurchaseOrderRepo{}
		if _, err := NewInventoryService(repo).SavePurchaseOrder(
			ctx, types.PurchaseOrderRequest{Supplier: " emzor ", Items: items}, 3,
		); err != nil {
			t.Fatal(err)
		}
		header := repo.created[0]
		if header.SupplierID == nil || header.SupplierName != "emzor" || header.CreatedByID != 3 {
			t.Fatalf("unexpected header: %+v", header)
		}
	})

	t.Run("no supplier is allowed", func(t *testing.T) {
		repo := &fakePurchaseOrderRepo{}
		if _, err := NewInventoryService(repo).SavePurchaseOrder(
			ctx, types.PurchaseOrderRequest{Items: items}, 3,
		); err != nil {
			t.Fatal(err)
		}
		if repo.created[0].SupplierID != nil || repo.created[0].SupplierName != "" {
			t.Fatalf("expected no supplier: %+v", repo.created[0])
		}
	})

	t.Run("editing can clear the supplier", func(t *testing.T) {
		order := newOrder(constant.PurchaseOrderOpen)
		repo := &fakePurchaseOrderRepo{order: order}
		_, err := NewInventoryService(repo).SavePurchaseOrder(ctx, types.PurchaseOrderRequest{
			ID:      1,
			Version: 1,
			Items: []types.PurchaseOrderItem{
				{ID: 1, Name: "Known", Quantity: 10},
				{ID: 2, Name: "Brand new", Quantity: 5},
			},
		}, 3)
		if err != nil {
			t.Fatal(err)
		}
		if repo.supplierSaved != nil || repo.supplierSavedN != "" {
			t.Fatalf("supplier should be cleared: %v %q", repo.supplierSaved, repo.supplierSavedN)
		}
	})
}

func TestPurchaseOrderNote(t *testing.T) {
	ctx := context.Background()
	items := []types.PurchaseOrderItem{{Name: "Thing", Quantity: 1}}

	t.Run("a long note is rejected", func(t *testing.T) {
		_, err := normalizePurchaseOrderRequest(types.PurchaseOrderRequest{
			Note: strings.Repeat("x", maxPurchaseOrderNote+1), Items: items,
		})
		requireHTTPStatus(t, err, http.StatusBadRequest)
	})

	t.Run("the note is trimmed and stored on create", func(t *testing.T) {
		repo := &fakePurchaseOrderRepo{}
		_, err := NewInventoryService(repo).SavePurchaseOrder(
			ctx, types.PurchaseOrderRequest{Note: "  deliver to the back door ", Items: items}, 1,
		)
		if err != nil {
			t.Fatal(err)
		}
		if repo.created[0].Note != "deliver to the back door" {
			t.Fatalf("unexpected note %q", repo.created[0].Note)
		}
	})

	t.Run("the note is updated on edit", func(t *testing.T) {
		repo := &fakePurchaseOrderRepo{order: newOrder(constant.PurchaseOrderOpen)}
		_, err := NewInventoryService(repo).SavePurchaseOrder(ctx, types.PurchaseOrderRequest{
			ID: 1, Version: 1, Note: "call before delivery",
			Items: []types.PurchaseOrderItem{
				{ID: 1, Name: "Known", Quantity: 10}, {ID: 2, Name: "Brand new", Quantity: 5},
			},
		}, 1)
		if err != nil {
			t.Fatal(err)
		}
		if repo.noteSaved != "call before delivery" {
			t.Fatalf("unexpected note %q", repo.noteSaved)
		}
	})
}

func TestPurchaseOrderItemNotes(t *testing.T) {
	t.Run("item notes are trimmed and limited", func(t *testing.T) {
		got, err := normalizePurchaseOrderRequest(types.PurchaseOrderRequest{
			Items: []types.PurchaseOrderItem{{Name: "A", Quantity: 1, Note: "  keep cold "}},
		})
		if err != nil || got.Items[0].Note != "keep cold" {
			t.Fatalf("note not trimmed: %+v %v", got.Items, err)
		}
		_, err = normalizePurchaseOrderRequest(types.PurchaseOrderRequest{
			Items: []types.PurchaseOrderItem{
				{Name: "A", Quantity: 1, Note: strings.Repeat("x", maxPurchaseOrderItemNote+1)},
			},
		})
		requireHTTPStatus(t, err, http.StatusBadRequest)
	})

	t.Run(
		"a line with received stock keeps its identity but its note can change",
		func(t *testing.T) {
			existing := []types.PurchaseOrderItem{
				{
					ID:               1,
					ProductID:        intPtr(10),
					Name:             "Received",
					Quantity:         10,
					ReceivedQuantity: 4,
					Note:             "old",
				},
			}
			plan, err := planPurchaseOrderEdit(existing, []types.PurchaseOrderItem{
				{ID: 1, Name: "Renamed", Quantity: 10, Note: "new"},
			})
			if err != nil {
				t.Fatal(err)
			}
			if plan.updates[0].Name != "Received" || plan.updates[0].Note != "new" {
				t.Fatalf("unexpected update: %+v", plan.updates[0])
			}
		},
	)

	t.Run("moved items keep their notes", func(t *testing.T) {
		got := outstandingPurchaseOrderItems([]types.PurchaseOrderItem{
			{ID: 1, Name: "A", Quantity: 5, ReceivedQuantity: 1, Note: "blue box"},
		})
		if got[0].Note != "blue box" {
			t.Fatalf("note lost: %+v", got[0])
		}
	})
}

func auditActions(entries []types.AuditEntry) []string {
	actions := make([]string, len(entries))
	for i, entry := range entries {
		actions[i] = entry.Action
	}
	return actions
}

func TestPurchaseOrderStaleEdits(t *testing.T) {
	ctx := context.Background()
	items := []types.PurchaseOrderItem{
		{ID: 1, Name: "Known", Quantity: 10}, {ID: 2, Name: "Brand new", Quantity: 5},
	}

	t.Run("an edit made from an old version is refused", func(t *testing.T) {
		order := newOrder(constant.PurchaseOrderOpen)
		order.Version = 4
		repo := &fakePurchaseOrderRepo{order: order}
		_, err := NewInventoryService(repo).SavePurchaseOrder(
			ctx, types.PurchaseOrderRequest{ID: 1, Version: 3, Items: items}, 1,
		)
		requireHTTPStatus(t, err, http.StatusConflict)
		if repo.commits != 0 || len(repo.audit) != 0 {
			t.Fatal("a refused edit must write nothing")
		}
	})

	t.Run("a missing version is refused", func(t *testing.T) {
		repo := &fakePurchaseOrderRepo{order: newOrder(constant.PurchaseOrderOpen)}
		_, err := NewInventoryService(repo).SavePurchaseOrder(
			ctx, types.PurchaseOrderRequest{ID: 1, Items: items}, 1,
		)
		requireHTTPStatus(t, err, http.StatusConflict)
	})

	t.Run("the current version is accepted", func(t *testing.T) {
		repo := &fakePurchaseOrderRepo{order: newOrder(constant.PurchaseOrderOpen)}
		_, err := NewInventoryService(repo).SavePurchaseOrder(
			ctx, types.PurchaseOrderRequest{ID: 1, Version: 1, Items: items}, 1,
		)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestPurchaseOrderAudit(t *testing.T) {
	ctx := context.Background()

	t.Run("create, edit, close and move are recorded with the actor", func(t *testing.T) {
		repo := &fakePurchaseOrderRepo{}
		svc := NewInventoryService(repo)
		if _, err := svc.SavePurchaseOrder(ctx, types.PurchaseOrderRequest{
			Items: []types.PurchaseOrderItem{{Name: "A", Quantity: 1}},
		}, 7); err != nil {
			t.Fatal(err)
		}

		repo.order = newOrder(constant.PurchaseOrderPartial)
		if _, err := svc.SavePurchaseOrder(ctx, types.PurchaseOrderRequest{
			ID: 1, Version: 1, Note: "changed",
			Items: []types.PurchaseOrderItem{
				{ID: 1, Name: "Known", Quantity: 12}, {ID: 2, Name: "Brand new", Quantity: 5},
			},
		}, 8); err != nil {
			t.Fatal(err)
		}
		if err := svc.ClosePurchaseOrder(ctx, 1, 1, "supplier cannot deliver", 9); err != nil {
			t.Fatal(err)
		}
		if _, err := svc.MovePurchaseOrderOutstanding(
			ctx,
			1,
			1,
			"Fidson",
			"out of stock",
			10,
		); err != nil {
			t.Fatal(err)
		}

		got := auditActions(repo.audit)
		want := []string{"created", "edited", "closed", "moved", "created"}
		if strings.Join(got, ",") != strings.Join(want, ",") {
			t.Fatalf("expected %v, got %v", want, got)
		}
		actors := []int{7, 8, 9, 10, 10}
		for i, entry := range repo.audit {
			if entry.ActorID != actors[i] || entry.EntityType != constant.AuditEntityPurchaseOrder {
				t.Fatalf("entry %d has the wrong actor or entity: %+v", i, entry)
			}
		}
		if repo.audit[2].Reason != "supplier cannot deliver" ||
			repo.audit[3].Reason != "out of stock" {
			t.Fatalf("reasons were not stored: %+v", repo.audit)
		}
		if !strings.Contains(string(repo.audit[1].Details), `"quantity":{"from":10,"to":12}`) {
			t.Fatalf("the edit should record before and after: %s", repo.audit[1].Details)
		}
	})

	t.Run("an edit that changes nothing writes no audit entry", func(t *testing.T) {
		repo := &fakePurchaseOrderRepo{order: newOrder(constant.PurchaseOrderOpen)}
		_, err := NewInventoryService(repo).SavePurchaseOrder(ctx, types.PurchaseOrderRequest{
			ID: 1, Version: 1,
			Items: []types.PurchaseOrderItem{
				{ID: 1, Name: "Known", Quantity: 10}, {ID: 2, Name: "Brand new", Quantity: 5},
			},
		}, 1)
		if err != nil {
			t.Fatal(err)
		}
		if len(repo.audit) != 0 {
			t.Fatalf("unexpected audit entries: %+v", repo.audit)
		}
	})

	t.Run("closing and moving need a reason", func(t *testing.T) {
		for _, reason := range []string{"", "  ", "no"} {
			repo := &fakePurchaseOrderRepo{order: newOrder(constant.PurchaseOrderPartial)}
			err := NewInventoryService(repo).ClosePurchaseOrder(ctx, 1, 1, reason, 1)
			requireHTTPStatus(t, err, http.StatusBadRequest)
			_, err = NewInventoryService(
				repo,
			).MovePurchaseOrderOutstanding(ctx, 1, 1, "Fidson", reason, 1)
			requireHTTPStatus(t, err, http.StatusBadRequest)
			if repo.commits != 0 || len(repo.audit) != 0 {
				t.Fatal("nothing should be written without a reason")
			}
		}
	})

	t.Run(
		"receiving records the lines, the batch and whether a product was linked",
		func(t *testing.T) {
			repo := &fakePurchaseOrderRepo{order: newOrder(constant.PurchaseOrderPartial)}
			err := NewInventoryService(
				repo,
			).applyPurchaseOrderReceipt(ctx, nil, types.ReceiveSupplyRequest{
				PurchaseOrderID: 1, UserID: 5, Supplier: "Emzor",
				Products: []types.ReceiveItem{
					{ID: 10, PurchaseOrderItemID: 1, Quantity: 4},
					{ID: 55, PurchaseOrderItemID: 2, Quantity: 2},
				},
			}, 42)
			if err != nil {
				t.Fatal(err)
			}
			if len(repo.audit) != 1 || repo.audit[0].Action != constant.AuditReceived ||
				repo.audit[0].ActorID != 5 {
				t.Fatalf("unexpected audit: %+v", repo.audit)
			}
			details := string(repo.audit[0].Details)
			for _, want := range []string{`"receiving_batch_id":42`, `"product_linked":true`, `"quantity":4`, `"outstanding_after":3`} {
				if !strings.Contains(details, want) {
					t.Fatalf("details missing %s: %s", want, details)
				}
			}
		},
	)
}

func TestReceiptWithoutOrderLinesIgnoresTheOrder(t *testing.T) {
	for _, status := range []string{constant.PurchaseOrderClosed, constant.PurchaseOrderReceived} {
		repo := &fakePurchaseOrderRepo{order: newOrder(status)}
		err := NewInventoryService(repo).applyPurchaseOrderReceipt(
			context.Background(), nil,
			types.ReceiveSupplyRequest{
				PurchaseOrderID: 1, Products: []types.ReceiveItem{{ID: 77, Quantity: 3}},
			},
			1,
		)
		if err != nil {
			t.Fatalf("a plain receipt must not fail because the order is %s: %v", status, err)
		}
		if len(repo.audit) != 0 || repo.refreshed != 0 {
			t.Fatal("the order should not be touched")
		}
	}
}

func TestSavingAnUnchangedOrderWritesNothing(t *testing.T) {
	repo := &fakePurchaseOrderRepo{order: newOrder(constant.PurchaseOrderOpen)}
	_, err := NewInventoryService(
		repo,
	).SavePurchaseOrder(context.Background(), types.PurchaseOrderRequest{
		ID: 1, Version: 1,
		Items: []types.PurchaseOrderItem{
			{ID: 1, Name: "Known", Quantity: 10}, {ID: 2, Name: "Brand new", Quantity: 5},
		},
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if repo.refreshed != 0 || repo.supplierSavedN != "" || len(repo.audit) != 0 {
		t.Fatalf("an unchanged save must not bump the version: refreshed=%d", repo.refreshed)
	}
}

func TestCloseAndMoveRefuseAStaleVersion(t *testing.T) {
	ctx := context.Background()
	order := newOrder(constant.PurchaseOrderPartial)
	order.Version = 5

	repo := &fakePurchaseOrderRepo{order: order}
	err := NewInventoryService(repo).ClosePurchaseOrder(ctx, 1, 4, "supplier cannot deliver", 1)
	requireHTTPStatus(t, err, http.StatusConflict)
	_, err = NewInventoryService(
		repo,
	).MovePurchaseOrderOutstanding(ctx, 1, 4, "Fidson", "out of stock", 1)
	requireHTTPStatus(t, err, http.StatusConflict)
	if repo.closed != 0 || len(repo.created) != 0 || repo.commits != 0 {
		t.Fatal("a stale close or move must write nothing")
	}

	if err = NewInventoryService(
		repo,
	).ClosePurchaseOrder(ctx, 1, 5, "supplier cannot deliver", 1); err != nil {
		t.Fatal(err)
	}
}

func TestEditKeepsProductsUniqueAfterIdentityIsRestored(t *testing.T) {
	existing := []types.PurchaseOrderItem{
		{ID: 1, ProductID: intPtr(5), Name: "Received", Quantity: 10, ReceivedQuantity: 4},
	}
	// The client points the received line at another product and adds a line for product 5.
	_, err := planPurchaseOrderEdit(existing, []types.PurchaseOrderItem{
		{ID: 1, ProductID: intPtr(9), Name: "Received", Quantity: 10},
		{ProductID: intPtr(5), Name: "Again", Quantity: 1},
	})
	requireHTTPStatus(t, err, http.StatusBadRequest)
}

func TestUnchangedSupplierIsNotRecreated(t *testing.T) {
	order := newOrder(constant.PurchaseOrderOpen)
	order.Supplier, order.SupplierID = "Emzor", intPtr(1)
	repo := &fakePurchaseOrderRepo{order: order}
	_, err := NewInventoryService(repo).SavePurchaseOrder(
		context.Background(),
		types.PurchaseOrderRequest{
			ID: 1, Version: 1, Supplier: "EMZOR",
			Items: []types.PurchaseOrderItem{
				{ID: 1, Name: "Known", Quantity: 10}, {ID: 2, Name: "Brand new", Quantity: 5},
			},
		},
		1,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(repo.suppliers) != 0 {
		t.Fatalf("no supplier should be looked up or created: %+v", repo.suppliers)
	}
}

func TestPurchaseOrderLimitsCountCharactersAndKeysDoNotCollide(t *testing.T) {
	// 200 multi-byte characters are within the limit even though they exceed 200 bytes.
	accented := strings.Repeat("é", maxPurchaseOrderText)
	if _, err := normalizePurchaseOrderRequest(types.PurchaseOrderRequest{
		Items: []types.PurchaseOrderItem{{Name: accented, Quantity: 1}},
	}); err != nil {
		t.Fatalf("200 characters should be accepted: %v", err)
	}

	// ("A|B", "") and ("A", "B") are different items.
	if _, err := normalizePurchaseOrderRequest(types.PurchaseOrderRequest{
		Items: []types.PurchaseOrderItem{
			{Name: "A|B", Quantity: 1}, {Name: "A", Manufacturer: "B", Quantity: 1},
		},
	}); err != nil {
		t.Fatalf("distinct items were treated as duplicates: %v", err)
	}
}
