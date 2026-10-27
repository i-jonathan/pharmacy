package service_test

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"

	"pharmacy/httperror"
	"pharmacy/internal/constant"
	"pharmacy/internal/testdb"
	"pharmacy/internal/types"
	"pharmacy/repository"
	"pharmacy/service"
)

// These run the real service and SQL against a throwaway database; see internal/testdb.

type fixture struct {
	db  *sqlx.DB
	svc service.InventoryService
	uid int
	ctx context.Context
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	db := testdb.New(t, 0)
	db.MustExec(`INSERT INTO users (username, password) VALUES ('tester', 'x')`)
	db.MustExec(`INSERT INTO category (name) VALUES ('General')`)
	f := &fixture{
		db:  db,
		svc: service.NewInventoryService(repository.NewStore(db)),
		ctx: context.Background(),
	}
	if err := db.Get(&f.uid, `SELECT id FROM users LIMIT 1`); err != nil {
		t.Fatal(err)
	}
	return f
}

func (f *fixture) product(t *testing.T, name string) int {
	t.Helper()
	var category int
	f.db.Get(&category, `SELECT id FROM category LIMIT 1`)
	res, err := f.svc.CreateProduct(f.ctx, types.CreateProductRequest{
		Name: name, Barcode: name, Manufacturer: "M", CategoryID: category,
		ReorderLevel: 5, CostPrice: 100, SellingPrice: 150,
	})
	if err != nil {
		t.Fatal(err)
	}
	return res.ID
}

func (f *fixture) order(
	t *testing.T,
	supplier string,
	items ...types.PurchaseOrderItem,
) types.PurchaseOrder {
	t.Helper()
	id, err := f.svc.SavePurchaseOrder(
		f.ctx,
		types.PurchaseOrderRequest{Supplier: supplier, Items: items},
		f.uid,
	)
	if err != nil {
		t.Fatal(err)
	}
	return f.get(t, id)
}

func (f *fixture) get(t *testing.T, id int) types.PurchaseOrder {
	t.Helper()
	order, err := f.svc.GetPurchaseOrder(f.ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	return order
}

func (f *fixture) receive(key, supplier string, orderID int, rows ...types.ReceiveItem) error {
	for i := range rows {
		rows[i].CostPrice, rows[i].SellingPrice = 100, 150
		rows[i].Expiry = time.Now().AddDate(1, 0, 0)
	}
	return f.svc.ReceiveProductSupply(f.ctx, types.ReceiveSupplyRequest{
		Supplier:        supplier,
		UserID:          f.uid,
		PurchaseOrderID: orderID,
		IdempotencyKey:  key,
		Products:        rows,
	})
}

func (f *fixture) stock(t *testing.T, productID int) int {
	t.Helper()
	var stock int
	if err := f.db.Get(
		&stock,
		`SELECT stock FROM inventory_view WHERE id = $1`,
		productID,
	); err != nil {
		t.Fatal(err)
	}
	return stock
}

func status(err error) int {
	var httperr *httperror.HTTPError
	if errors.As(err, &httperr) {
		return httperr.Code
	}
	return 0
}

func TestPartialSupplyAndTransferInvariant(t *testing.T) {
	f := newFixture(t)
	alpha := f.product(t, "Alpha")
	po := f.order(t, "emzor",
		types.PurchaseOrderItem{ProductID: &alpha, Name: "Alpha", Quantity: 10, Note: "blister"},
		types.PurchaseOrderItem{Name: "Never Seen", Quantity: 4},
	)
	if po.Supplier != "emzor" || po.SupplierID == nil || po.Version != 1 || po.Status != "open" {
		t.Fatalf("unexpected new order: %+v", po)
	}

	if err := f.receive(
		"k1",
		"Emzor",
		po.ID,
		types.ReceiveItem{ID: alpha, PurchaseOrderItemID: po.Items[0].ID, Quantity: 4},
	); err != nil {
		t.Fatal(err)
	}
	po = f.get(t, po.ID)
	if po.Status != "partial" || po.Items[0].ReceivedQuantity != 4 || po.Version != 2 {
		t.Fatalf("unexpected order after a partial receipt: %+v", po)
	}
	// Both spellings resolve to the supplier created by the order.
	var suppliers int
	f.db.Get(&suppliers, `SELECT count(*) FROM supplier`)
	if suppliers != 1 {
		t.Fatalf("one supplier expected, got %d", suppliers)
	}

	// The batch records which order line it was received against.
	var linked int
	f.db.Get(
		&linked,
		`SELECT count(*) FROM product_batch WHERE purchase_order_item_id = $1 AND quantity = 4`,
		po.Items[0].ID,
	)
	if linked != 1 {
		t.Fatalf("the stock batch should reference the order line, got %d", linked)
	}

	// Ordered = received + outstanding: moving the rest carries 6, not 10, and 4 for the new item.
	moved, err := f.svc.MovePurchaseOrderOutstanding(
		f.ctx,
		po.ID,
		f.get(t, po.ID).Version,
		"Fidson",
		"supplier out of stock",
		f.uid,
	)
	if err != nil {
		t.Fatal(err)
	}
	next, original := f.get(t, moved), f.get(t, po.ID)
	if original.Status != "closed" || original.MovedToID == nil || *original.MovedToID != moved {
		t.Fatalf("original should be closed and point at the new order: %+v", original)
	}
	if next.Status != "open" || next.SourceOrderID == nil || *next.SourceOrderID != po.ID ||
		next.Supplier != "Fidson" {
		t.Fatalf("unexpected new order: %+v", next)
	}
	if len(next.Items) != 2 || next.Items[0].Quantity != 6 || next.Items[0].Note != "blister" ||
		next.Items[1].Quantity != 4 {
		t.Fatalf("outstanding quantities not carried over: %+v", next.Items)
	}
	if original.Items[0].Quantity != 10 || original.Items[0].ReceivedQuantity != 4 {
		t.Fatalf("the original keeps what it ordered and received: %+v", original.Items)
	}

	// The remainder cannot be taken by a second order, and the closed order cannot be received.
	if _, err = f.svc.MovePurchaseOrderOutstanding(
		f.ctx,
		po.ID,
		f.get(t, po.ID).Version,
		"Other",
		"again",
		f.uid,
	); status(
		err,
	) != http.StatusConflict {
		t.Fatalf("a second move should conflict, got %v", err)
	}
	err = f.receive(
		"k2",
		"Emzor",
		po.ID,
		types.ReceiveItem{ID: alpha, PurchaseOrderItemID: po.Items[0].ID, Quantity: 1},
	)
	if status(err) != http.StatusConflict {
		t.Fatalf("receiving on a closed order should conflict, got %v", err)
	}

	// The new supplier delivers against the new order.
	if err = f.receive(
		"k3",
		"fidson",
		moved,
		types.ReceiveItem{ID: alpha, PurchaseOrderItemID: next.Items[0].ID, Quantity: 6},
	); err != nil {
		t.Fatal(err)
	}
	if got := f.get(t, moved); got.Items[0].ReceivedQuantity != 6 {
		t.Fatalf("unexpected receipt on the new order: %+v", got.Items)
	}
	if stock := f.stock(t, alpha); stock != 10 {
		t.Fatalf("stock should be 4 + 6, got %d", stock)
	}
}

func TestFailedReceiptLeavesNothingBehind(t *testing.T) {
	f := newFixture(t)
	alpha := f.product(t, "Alpha")
	po := f.order(
		t,
		"Emzor",
		types.PurchaseOrderItem{ProductID: &alpha, Name: "Alpha", Quantity: 5},
	)

	err := f.receive(
		"k1",
		"Emzor",
		po.ID,
		types.ReceiveItem{ID: alpha, PurchaseOrderItemID: po.Items[0].ID, Quantity: 6},
	)
	if status(err) != http.StatusBadRequest {
		t.Fatalf("over-receiving should be rejected, got %v", err)
	}
	if stock := f.stock(t, alpha); stock != 0 {
		t.Fatalf("a rejected receipt must not change stock, got %d", stock)
	}
	if got := f.get(t, po.ID); got.Items[0].ReceivedQuantity != 0 || got.Version != 1 {
		t.Fatalf("a rejected receipt must not touch the order: %+v", got)
	}
	var batches, audits int
	f.db.Get(&batches, `SELECT count(*) FROM receiving_batch`)
	f.db.Get(&audits, `SELECT count(*) FROM audit_log WHERE action = 'received'`)
	if batches != 0 || audits != 0 {
		t.Fatalf("nothing should be recorded, got %d batches and %d audit entries", batches, audits)
	}
}

func TestRetriedReceiptIsAppliedOnce(t *testing.T) {
	f := newFixture(t)
	alpha := f.product(t, "Alpha")
	po := f.order(
		t,
		"Emzor",
		types.PurchaseOrderItem{ProductID: &alpha, Name: "Alpha", Quantity: 10},
	)
	row := types.ReceiveItem{ID: alpha, PurchaseOrderItemID: po.Items[0].ID, Quantity: 4}

	for attempt := 0; attempt < 3; attempt++ {
		if err := f.receive("same-key", "Emzor", po.ID, row); err != nil {
			t.Fatalf("attempt %d: %v", attempt, err)
		}
	}
	if got := f.get(t, po.ID); got.Items[0].ReceivedQuantity != 4 {
		t.Fatalf("the retried receipt must count once, got %d", got.Items[0].ReceivedQuantity)
	}
	if stock := f.stock(t, alpha); stock != 4 {
		t.Fatalf("stock must be received once, got %d", stock)
	}
}

func TestConcurrentReceiptsCannotOverReceive(t *testing.T) {
	f := newFixture(t)
	alpha := f.product(t, "Alpha")
	po := f.order(
		t,
		"Emzor",
		types.PurchaseOrderItem{ProductID: &alpha, Name: "Alpha", Quantity: 10},
	)

	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for _, key := range []string{"a", "b"} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- f.receive(key, "Emzor", po.ID, types.ReceiveItem{ID: alpha, PurchaseOrderItemID: po.Items[0].ID, Quantity: 6})
		}()
	}
	wg.Wait()
	close(errs)
	succeeded := 0
	for err := range errs {
		if err == nil {
			succeeded++
		}
	}
	if succeeded != 1 {
		t.Fatalf("exactly one of two 6-unit receipts fits in 10, got %d", succeeded)
	}
	if got := f.get(t, po.ID); got.Items[0].ReceivedQuantity != 6 || f.stock(t, alpha) != 6 {
		t.Fatalf("order and stock must agree on 6: %+v stock %d", got.Items, f.stock(t, alpha))
	}
}

func TestConcurrentTransfersCreateOneOrder(t *testing.T) {
	f := newFixture(t)
	alpha := f.product(t, "Alpha")
	po := f.order(
		t,
		"Emzor",
		types.PurchaseOrderItem{ProductID: &alpha, Name: "Alpha", Quantity: 10},
	)

	results := make(chan error, 2)
	version := f.get(t, po.ID).Version
	for _, supplier := range []string{"Fidson", "Neimeth"} {
		go func() {
			_, err := f.svc.MovePurchaseOrderOutstanding(
				f.ctx,
				po.ID,
				version,
				supplier,
				"out of stock",
				f.uid,
			)
			results <- err
		}()
	}
	var ok, conflict int
	for i := 0; i < 2; i++ {
		switch err := <-results; {
		case err == nil:
			ok++
		case status(err) == http.StatusConflict:
			conflict++
		default:
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if ok != 1 || conflict != 1 {
		t.Fatalf("one transfer should win and one conflict, got %d and %d", ok, conflict)
	}
	var orders int
	f.db.Get(
		&orders,
		`SELECT count(*) FROM purchase_order WHERE source_purchase_order_id = $1`,
		po.ID,
	)
	if orders != 1 {
		t.Fatalf("the outstanding items must exist on exactly one new order, got %d", orders)
	}
}

func TestStaleEditsAreRefused(t *testing.T) {
	f := newFixture(t)
	alpha := f.product(t, "Alpha")
	po := f.order(
		t,
		"Emzor",
		types.PurchaseOrderItem{ProductID: &alpha, Name: "Alpha", Quantity: 10},
	)
	line := po.Items[0]

	// Person B loads the order, then person A saves an extra line.
	staleVersion := po.Version
	_, err := f.svc.SavePurchaseOrder(f.ctx, types.PurchaseOrderRequest{
		ID: po.ID, Version: po.Version, Supplier: "Emzor",
		Items: []types.PurchaseOrderItem{
			{ID: line.ID, ProductID: &alpha, Name: "Alpha", Quantity: 10},
			{Name: "Added by A", Quantity: 2},
		},
	}, f.uid)
	if err != nil {
		t.Fatal(err)
	}

	// B saves from the old copy: without the version check this would delete A's line.
	_, err = f.svc.SavePurchaseOrder(f.ctx, types.PurchaseOrderRequest{
		ID:       po.ID,
		Version:  staleVersion,
		Supplier: "Emzor",
		Items: []types.PurchaseOrderItem{
			{ID: line.ID, ProductID: &alpha, Name: "Alpha", Quantity: 12},
		},
	}, f.uid)
	if status(err) != http.StatusConflict {
		t.Fatalf("a stale edit should conflict, got %v", err)
	}
	got := f.get(t, po.ID)
	if len(got.Items) != 2 || got.Items[0].Quantity != 10 {
		t.Fatalf("the stale edit must not change anything: %+v", got.Items)
	}

	// A receipt also moves the version, so an editor holding an older copy cannot save over it.
	if err = f.receive(
		"k",
		"Emzor",
		po.ID,
		types.ReceiveItem{ID: alpha, PurchaseOrderItemID: line.ID, Quantity: 3},
	); err != nil {
		t.Fatal(err)
	}
	_, err = f.svc.SavePurchaseOrder(f.ctx, types.PurchaseOrderRequest{
		ID:       po.ID,
		Version:  got.Version,
		Supplier: "Emzor",
		Items: []types.PurchaseOrderItem{
			{ID: line.ID, ProductID: &alpha, Name: "Alpha", Quantity: 10},
		},
	}, f.uid)
	if status(err) != http.StatusConflict {
		t.Fatalf("an edit made before a receipt should conflict, got %v", err)
	}

	// A closed order cannot be reopened by an edit that has the current version.
	current := f.get(t, po.ID)
	if err = f.svc.ClosePurchaseOrder(
		f.ctx,
		po.ID,
		f.get(t, po.ID).Version,
		"supplier cannot deliver",
		f.uid,
	); err != nil {
		t.Fatal(err)
	}
	closed := f.get(t, po.ID)
	_, err = f.svc.SavePurchaseOrder(f.ctx, types.PurchaseOrderRequest{
		ID:      po.ID,
		Version: closed.Version,
		Items:   []types.PurchaseOrderItem{{ID: line.ID, Name: "Alpha", Quantity: 10}},
	}, f.uid)
	if status(err) != http.StatusConflict ||
		f.get(t, po.ID).Status != constant.PurchaseOrderClosed {
		t.Fatalf("a closed order must stay closed, got %v", err)
	}
	if closed.Version <= current.Version {
		t.Fatal("closing should change the version")
	}
}

func TestAuditTrailOfAnOrder(t *testing.T) {
	f := newFixture(t)
	alpha := f.product(t, "Alpha")
	po := f.order(
		t,
		"Emzor",
		types.PurchaseOrderItem{ProductID: &alpha, Name: "Alpha", Quantity: 10},
	)

	_, err := f.svc.SavePurchaseOrder(f.ctx, types.PurchaseOrderRequest{
		ID:       po.ID,
		Version:  po.Version,
		Supplier: "Emzor",
		Note:     "call first",
		Items: []types.PurchaseOrderItem{
			{ID: po.Items[0].ID, ProductID: &alpha, Name: "Alpha", Quantity: 12},
		},
	}, f.uid)
	if err != nil {
		t.Fatal(err)
	}
	if err = f.receive(
		"k",
		"Emzor",
		po.ID,
		types.ReceiveItem{ID: alpha, PurchaseOrderItemID: po.Items[0].ID, Quantity: 5},
	); err != nil {
		t.Fatal(err)
	}
	if err = f.svc.ClosePurchaseOrder(
		f.ctx,
		po.ID,
		f.get(t, po.ID).Version,
		"supplier cannot deliver",
		f.uid,
	); err != nil {
		t.Fatal(err)
	}

	history, err := f.svc.GetPurchaseOrderHistory(f.ctx, po.ID)
	if err != nil {
		t.Fatal(err)
	}
	var actions []string
	for _, record := range history {
		actions = append(actions, record.Action)
		if record.Actor != "tester" {
			t.Fatalf("the actor should be recorded: %+v", record)
		}
	}
	// Newest first.
	want := []string{"closed", "received", "edited", "created"}
	if len(actions) != len(want) {
		t.Fatalf("expected %v, got %v", want, actions)
	}
	for i := range want {
		if actions[i] != want[i] {
			t.Fatalf("expected %v, got %v", want, actions)
		}
	}
	if history[0].Reason != "supplier cannot deliver" {
		t.Fatalf("the reason should be stored: %+v", history[0])
	}

	if _, err = f.svc.GetPurchaseOrderHistory(f.ctx, 9999); status(err) != http.StatusNotFound {
		t.Fatalf("history of an unknown order should be not found, got %v", err)
	}
}

func TestTransferAndAuditAreAtomic(t *testing.T) {
	f := newFixture(t)
	alpha := f.product(t, "Alpha")
	po := f.order(
		t,
		"Emzor",
		types.PurchaseOrderItem{ProductID: &alpha, Name: "Alpha", Quantity: 10},
	)

	// Break the audit write for the second entry of the transfer. The whole transfer must roll
	// back: no new order, the original still open.
	f.db.MustExec(`
		CREATE FUNCTION fail_second_audit() RETURNS trigger AS $$
		BEGIN
			IF NEW.action = 'created' AND NEW.details ? 'moved_from_purchase_order_id' THEN
				RAISE EXCEPTION 'audit write failed';
			END IF;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql`)
	f.db.MustExec(
		`CREATE TRIGGER fail_audit BEFORE INSERT ON audit_log FOR EACH ROW EXECUTE FUNCTION fail_second_audit()`,
	)

	if _, err := f.svc.MovePurchaseOrderOutstanding(
		f.ctx,
		po.ID,
		po.Version,
		"Fidson",
		"out of stock",
		f.uid,
	); err == nil {
		t.Fatal("the transfer should fail when its audit record cannot be written")
	}
	got := f.get(t, po.ID)
	var orders int
	f.db.Get(&orders, `SELECT count(*) FROM purchase_order`)
	if got.Status != "open" || got.MovedToID != nil || orders != 1 {
		t.Fatalf("a failed transfer must leave nothing behind: %+v, %d orders", got, orders)
	}
}
