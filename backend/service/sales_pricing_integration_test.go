package service_test

import (
	"errors"
	"net/http"
	"testing"

	"pharmacy/httperror"
	"pharmacy/internal/types"
	"pharmacy/repository"
	"pharmacy/service"
)

func TestTheServerPricesTheSaleAgainstTheDatabase(t *testing.T) {
	f := newFixture(t)
	sales := service.NewSaleService(repository.NewStore(f.db))
	ctx := asUser()

	pid := f.product(t, "Panadol") // sells at ₦150.00
	var priceID int
	err := f.db.Get(&priceID, `SELECT id FROM product_price WHERE product_id = $1`, pid)
	if err != nil {
		t.Fatal(err)
	}
	counts := func() int {
		return f.count(t, `
			SELECT (SELECT COUNT(*) FROM sales) + (SELECT COUNT(*) FROM sales_item)
			+ (SELECT COUNT(*) FROM sales_payment)
			+ (SELECT COUNT(*) FROM stock_movement WHERE movement_type = 'OUT_SALE')`)
	}
	before := counts()

	// A till that sends a ₦1 price for a ₦150 product is refused and nothing is written.
	cheat := types.Sale{
		CashierID: f.uid, IdempotencyKey: "cheat", Subtotal: 2, Total: 2,
		Items: []types.SaleItem{
			{ProductID: pid, Quantity: 2, PriceID: priceID, UnitPrice: 1, Total: 2},
		},
		Payments: []types.SalePayment{{Amount: 2, PaymentMethod: "Cash"}},
	}
	_, err = sales.CreateSale(ctx, cheat)
	var refusal *httperror.HTTPError
	if !errors.As(err, &refusal) || refusal.Code != http.StatusConflict ||
		refusal.Reason != "price_changed" {
		t.Fatalf("a cheaper price should be refused as price_changed: %v", err)
	}
	if counts() != before {
		t.Fatal("a refused sale wrote rows")
	}

	// An honest sale stores the price list's numbers: 2 × ₦150.00 less ₦30.00.
	honest := types.Sale{
		CashierID: f.uid, IdempotencyKey: "honest", Subtotal: 300, Discount: 30, Total: 270,
		Items: []types.SaleItem{
			{
				ProductID: pid,
				Quantity:  2,
				PriceID:   priceID,
				UnitPrice: 150,
				Discount:  30,
				Total:     270,
			},
		},
		Payments: []types.SalePayment{{Amount: 270, PaymentMethod: "Cash"}},
	}
	if _, err := sales.CreateSale(ctx, honest); err != nil {
		t.Fatal(err)
	}
	var sub, disc, total, unit, line int
	if err := f.db.QueryRow(`SELECT subtotal, discount, total FROM sales`).
		Scan(&sub, &disc, &total); err != nil {
		t.Fatal(err)
	}
	if err := f.db.QueryRow(`SELECT unit_price, total_price FROM sales_item`).
		Scan(&unit, &line); err != nil {
		t.Fatal(err)
	}
	if sub != 30000 || disc != 3000 || total != 27000 || unit != 15000 || line != 27000 {
		t.Fatalf("stored kobo: sale %d/%d/%d line %d/%d", sub, disc, total, unit, line)
	}

	// A return refunds from the stored line, so half of ₦270.00.
	history, err := sales.FetchSalesHistory(ctx, types.SaleFilter{Page: 1, PerPage: 20})
	if err != nil || len(history.Data) != 1 {
		t.Fatalf("history: %v %v", history.Data, err)
	}
	err = sales.ReturnItems(ctx, types.ReturnSale{
		SaleID: history.Data[0].ID, CashierID: f.uid,
		ReturnItems: []types.ReturnSaleItem{{SaleItemID: history.Data[0].Items[0].ID, Quantity: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	refunded := f.count(t, `SELECT total_refunded FROM returns`)
	if refunded != 13500 {
		t.Fatalf("refunded %d kobo, want 13500", refunded)
	}

	// A price edited after the item went into the cart refuses the old cart.
	f.db.MustExec(`UPDATE product_price SET selling_price = 20000 WHERE id = $1`, priceID)
	honest.IdempotencyKey = "stale"
	_, err = sales.CreateSale(ctx, honest)
	if !errors.As(err, &refusal) || refusal.Reason != "price_changed" {
		t.Fatalf("a stale cart price should be refused: %v", err)
	}
}
