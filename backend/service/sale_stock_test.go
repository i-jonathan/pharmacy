package service

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"pharmacy/internal/constant"
	"pharmacy/internal/types"

	"github.com/jmoiron/sqlx"
)

// saleFlag is a stock flag a sale opened.
type saleFlag struct {
	productID, userID int
	note              string
}

func (r *customerSaleRepo) LockProductStockTx(
	_ context.Context, _ *sqlx.Tx, ids []int,
) ([]types.ProductStock, error) {
	r.locked = append(r.locked, ids)
	r.events = append(r.events, "lock")
	out := make([]types.ProductStock, len(ids))
	for i, id := range ids {
		stock := 1_000_000
		if r.stock != nil {
			stock = r.stock[id]
		}
		out[i] = types.ProductStock{ID: id, Name: "Paracetamol 500mg", Stock: stock}
	}
	return out, nil
}

func (r *customerSaleRepo) InsertAuditLogTx(
	_ context.Context, _ *sqlx.Tx, e types.AuditEntry,
) error {
	r.audits = append(r.audits, e)
	return nil
}

func (r *customerSaleRepo) OpenSaleStockFlagTx(
	_ context.Context, _ *sqlx.Tx, productID, userID int, note string, _ int,
) error {
	r.saleFlags = append(r.saleFlags, saleFlag{productID, userID, note})
	return nil
}

// packAndSingles sells product 1 twice: 1 pack of 10 (₦900) and 3 singles (₦300), 13 in stock
// units in all.
func packAndSingles() types.Sale {
	return types.Sale{
		CashierID:      4,
		IdempotencyKey: "stock-key-1",
		Subtotal:       1200,
		Total:          1200,
		Items: []types.SaleItem{
			{ProductID: 1, PriceID: 2, Quantity: 1, UnitPrice: 900, Total: 900},
			{ProductID: 1, PriceID: 1, Quantity: 3, UnitPrice: 100, Total: 300},
		},
		Payments: []types.SalePayment{{Amount: 1200, PaymentMethod: "Cash"}},
	}
}

func TestASaleWithinStockLocksItsProductsAndSaves(t *testing.T) {
	repo := newPricingRepo()
	repo.stock = map[int]int{1: 13}
	if err := createSale(t, repo, packAndSingles()); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(repo.locked, [][]int{{1}}) {
		t.Fatalf("want product 1 locked once, got %v", repo.locked)
	}
	if repo.commits != 1 || len(repo.audits) != 0 || len(repo.saleFlags) != 0 {
		t.Fatalf("commits %d, audits %v, flags %v", repo.commits, repo.audits, repo.saleFlags)
	}
}

func TestASaleBeyondStockIsRefused(t *testing.T) {
	repo := newPricingRepo()
	repo.stock = map[int]int{1: 12} // the two lines take 13
	err := createSale(t, repo, packAndSingles())
	refusal := wantRefusal(t, err, http.StatusConflict, "insufficient_stock")
	want := "Paracetamol 500mg (12 units in stock, selling 13 units)"
	if !strings.Contains(refusal.Message, want) {
		t.Fatalf("message should name the product and the numbers: %q", refusal.Message)
	}
	wantNothingSaved(t, repo)
}

func TestTheProductsOfASaleAreLockedInIDOrder(t *testing.T) {
	repo := newPricingRepo()
	sale := honestSale()
	sale.Items = []types.SaleItem{
		{ProductID: 2, PriceID: 3, Quantity: 1, UnitPrice: 19.99, Total: 19.99},
		{ProductID: 1, PriceID: 1, Quantity: 1, UnitPrice: 100, Total: 100},
	}
	sale.Subtotal, sale.Discount, sale.Total = 119.99, 0, 119.99
	if err := createSale(t, repo, sale); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(repo.locked, [][]int{{1, 2}}) {
		t.Fatalf("want products 1 and 2 locked in order, got %v", repo.locked)
	}
}

func TestOversellingNeedsThePermission(t *testing.T) {
	repo := newPricingRepo()
	repo.stock = map[int]int{1: 0}
	sale := packAndSingles()
	sale.AllowOversell = true
	_, err := NewSaleService(repo).CreateSale(contextWith(), sale)
	wantRefusal(t, err, http.StatusForbidden, "")
	wantNothingSaved(t, repo)
}

func TestAPermittedOversellSavesWithAnAuditEntryAndAFlag(t *testing.T) {
	repo := newPricingRepo()
	repo.stock = map[int]int{1: 5}
	sale := packAndSingles()
	sale.AllowOversell = true
	_, err := NewSaleService(repo).CreateSale(contextWith(constant.OversellPermissionKey), sale)
	if err != nil {
		t.Fatal(err)
	}
	if repo.commits != 1 {
		t.Fatalf("want the sale committed, got %d commits", repo.commits)
	}
	if len(repo.audits) != 1 {
		t.Fatalf("want one audit entry, got %v", repo.audits)
	}
	a := repo.audits[0]
	if a.ActorID != 4 || a.EntityType != "sale" || a.EntityID != 9 || a.Action != "oversold" {
		t.Fatalf("audit entry: %+v", a)
	}
	var details struct {
		Products []struct {
			ProductID   int `json:"product_id"`
			StockBefore int `json:"stock_before"`
			Sold        int `json:"sold"`
		} `json:"products"`
	}
	if err := json.Unmarshal(a.Details, &details); err != nil {
		t.Fatal(err)
	}
	if len(details.Products) != 1 || details.Products[0].ProductID != 1 ||
		details.Products[0].StockBefore != 5 || details.Products[0].Sold != 13 {
		t.Fatalf("audit details: %s", a.Details)
	}
	if len(repo.saleFlags) != 1 {
		t.Fatalf("want one stock flag, got %v", repo.saleFlags)
	}
	f := repo.saleFlags[0]
	if f.productID != 1 || f.userID != 4 ||
		!strings.Contains(f.note, "Sold 13 units with 5 units in stock") ||
		!strings.Contains(f.note, repo.sales[0].ReceiptNumber) {
		t.Fatalf("stock flag: %+v (receipt %s)", f, repo.sales[0].ReceiptNumber)
	}
}

func TestAllowOversellWithinStockWritesNothingExtra(t *testing.T) {
	repo := newPricingRepo()
	repo.stock = map[int]int{1: 13}
	sale := packAndSingles()
	sale.AllowOversell = true
	_, err := NewSaleService(repo).CreateSale(contextWith(constant.OversellPermissionKey), sale)
	if err != nil {
		t.Fatal(err)
	}
	if len(repo.audits) != 0 || len(repo.saleFlags) != 0 {
		t.Fatalf("audits %v, flags %v", repo.audits, repo.saleFlags)
	}
}

func TestARetryOfASavedSaleIsNotCheckedForStock(t *testing.T) {
	repo := newPricingRepo()
	repo.duplicate = true
	repo.stock = map[int]int{1: 0}
	if err := createSale(t, repo, packAndSingles()); err != nil {
		t.Fatalf("a retry should be answered as saved, got %v", err)
	}
	if len(repo.locked) != 0 {
		t.Fatalf("a retry locked %v", repo.locked)
	}
}

// A deactivated customer is reported before a stock shortage, so a manager isn't asked to confirm
// an oversell only to be refused for the customer next.
func TestADeactivatedCustomerIsReportedBeforeAStockShortage(t *testing.T) {
	gone := types.Customer{ID: 5, Name: "Ada Obi", Active: false}
	repo := newPricingRepo()
	repo.customers = map[int]types.Customer{5: gone}
	repo.stock = map[int]int{1: 0}
	sale := packAndSingles()
	sale.CustomerID = 5
	_, err := NewSaleService(repo).CreateSale(
		contextWith(constant.ViewCustomersPermissionKey, constant.OversellPermissionKey), sale,
	)
	// The products are locked first (prices are read under the lock), but the customer is refused
	// before the stock is compared.
	wantRefusal(t, err, http.StatusConflict, "customer_inactive")
	wantNothingSaved(t, repo)
}

// Prices are read under the product lock, so a product edit holding it (which may change or delete
// a price) commits before this sale is priced, not between pricing and saving.
func TestASaleIsPricedUnderTheProductLock(t *testing.T) {
	repo := newPricingRepo()
	if err := createSale(t, repo, packAndSingles()); err != nil {
		t.Fatal(err)
	}
	if len(repo.events) == 0 || repo.events[0] != "lock" {
		t.Fatalf("want the lock before any price is read, got %v", repo.events)
	}
	if !reflect.DeepEqual(repo.locked, [][]int{{1}}) {
		t.Fatalf("want product 1 locked once, got %v", repo.locked)
	}
}

func TestOneUnitIsNotCalledUnits(t *testing.T) {
	repo := newPricingRepo()
	repo.stock = map[int]int{1: 1}
	sale := packAndSingles()
	sale.Items = sale.Items[:1] // one pack of 10
	sale.Subtotal, sale.Total = 900, 900
	sale.Payments[0].Amount = 900
	refusal := wantRefusal(t, createSale(t, repo, sale), http.StatusConflict, "insufficient_stock")
	if !strings.Contains(refusal.Message, "(1 unit in stock, selling 10 units)") {
		t.Fatalf("message: %q", refusal.Message)
	}
}
