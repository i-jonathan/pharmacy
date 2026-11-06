package service_test

import (
	"net/http"
	"testing"

	"pharmacy/httperror"
	"pharmacy/internal/types"
	"pharmacy/repository"
	"pharmacy/service"

	"errors"
)

func TestReceiptNumberAndReprintAgainstTheDatabase(t *testing.T) {
	f := newFixture(t)
	sales := service.NewSaleService(repository.NewStore(f.db))
	ctx := asUser()

	// A product with no manufacturer must not break the history that lists its sales.
	pid := f.product(t, "Panadol")
	f.db.MustExec(`UPDATE product SET manufacturer = NULL WHERE id = $1`, pid)
	var priceID int
	err := f.db.Get(&priceID, `SELECT id FROM product_price WHERE product_id = $1`, pid)
	if err != nil {
		t.Fatal(err)
	}

	sale := types.Sale{
		CashierID: f.uid, IdempotencyKey: "till-1", Subtotal: 150, Total: 150,
		Items: []types.SaleItem{
			{ProductID: pid, Quantity: 1, PriceID: priceID, UnitPrice: 150, Total: 150},
		},
		Payments: []types.SalePayment{{Amount: 150, PaymentMethod: "Cash"}},
	}
	first, err := sales.CreateSale(ctx, sale)
	if err != nil || first == "" {
		t.Fatalf("expected a receipt number, got %q, %v", first, err)
	}
	again, err := sales.CreateSale(ctx, sale) // the till retries after a lost reply
	if err != nil || again != first {
		t.Fatalf("a repeat should return %q, got %q, %v", first, again, err)
	}
	var count int
	f.db.Get(&count, `SELECT COUNT(*) FROM sales`)
	if count != 1 {
		t.Fatalf("the repeat must not save a second sale, found %d", count)
	}

	history, err := sales.FetchSalesHistory(ctx, types.SaleFilter{Page: 1, PerPage: 20})
	if err != nil {
		t.Fatalf("history failed for a product without a manufacturer: %v", err)
	}
	if len(history.Data) != 1 || history.Data[0].ReceiptNumber != first {
		t.Fatalf("history should list the sale under %q: %#v", first, history.Data)
	}
	saleID := history.Data[0].ID

	reprint, err := sales.ReprintReceipt(ctx, saleID, f.uid)
	if err != nil {
		t.Fatal(err)
	}
	if reprint.ReceiptNumber != first || reprint.ReprintedAt == "" {
		t.Fatalf("unexpected reprint reply: %#v", reprint)
	}
	var action, receipt string
	err = f.db.QueryRow(`
		SELECT action, details->>'receipt_number' FROM audit_log
		WHERE entity_type = 'sale' AND entity_id = $1 AND actor_id = $2`, saleID, f.uid,
	).Scan(&action, &receipt)
	if err != nil || action != "receipt_reprinted" || receipt != first {
		t.Fatalf("expected a receipt_reprinted audit row for %q: %q %q %v", first, action, receipt, err)
	}
	// Each reprint is its own entry.
	if _, err := sales.ReprintReceipt(ctx, saleID, f.uid); err != nil {
		t.Fatal(err)
	}
	f.db.Get(&count, `SELECT COUNT(*) FROM audit_log WHERE entity_type = 'sale'`)
	if count != 2 {
		t.Fatalf("expected two audit rows after two reprints, got %d", count)
	}

	_, err = sales.ReprintReceipt(ctx, saleID+1000, f.uid)
	var httpErr *httperror.HTTPError
	if !errors.As(err, &httpErr) || httpErr.Code != http.StatusNotFound {
		t.Fatalf("an unknown sale should be a 404, got %v", err)
	}
}
