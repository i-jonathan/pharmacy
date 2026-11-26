package service_test

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"pharmacy/httperror"
	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"pharmacy/repository"
	"pharmacy/service"
)

func TestASaleCannotTakeStockBelowZeroUnlessAManagerOversells(t *testing.T) {
	f := newFixture(t)
	sales := service.NewSaleService(repository.NewStore(f.db))

	pid := f.stocked(t, "Panadol", 2) // sells at ₦150.00, 1 stock unit each
	var priceID int
	err := f.db.Get(&priceID, `SELECT id FROM product_price WHERE product_id = $1`, pid)
	if err != nil {
		t.Fatal(err)
	}
	stock := func() int {
		return f.count(t, `SELECT COALESCE(SUM(CASE WHEN movement_type LIKE 'IN%' THEN quantity
			ELSE -quantity END), 0) FROM stock_movement WHERE product_id = $1`, pid)
	}
	sellThree := func(key string, allow bool) types.Sale {
		return types.Sale{
			CashierID: f.uid, IdempotencyKey: key, Subtotal: 450, Total: 450, AllowOversell: allow,
			Items: []types.SaleItem{
				{ProductID: pid, Quantity: 3, PriceID: priceID, UnitPrice: 150, Total: 450},
			},
			Payments: []types.SalePayment{{Amount: 450, PaymentMethod: "Cash"}},
		}
	}

	// 3 with 2 in stock is refused, and nothing is written.
	_, err = sales.CreateSale(asUser(), sellThree("short", false))
	var refusal *httperror.HTTPError
	if !errors.As(err, &refusal) || refusal.Code != http.StatusConflict ||
		refusal.Reason != "insufficient_stock" {
		t.Fatalf("want insufficient_stock, got %v", err)
	}
	if f.count(t, `SELECT COUNT(*) FROM sales`) != 0 || stock() != 2 {
		t.Fatal("a refused sale wrote rows")
	}

	// A cashier can't oversell even by asking.
	_, err = sales.CreateSale(asUser(), sellThree("cashier", true))
	if !errors.As(err, &refusal) || refusal.Code != http.StatusForbidden {
		t.Fatalf("want 403 for a cashier asking to oversell, got %v", err)
	}

	// A manager can: stock goes to -1, with an audit entry and an open flag showing -1.
	manager := asUser(constant.OversellPermissionKey)
	receipt, err := sales.CreateSale(manager, sellThree("oversell", true))
	if err != nil {
		t.Fatal(err)
	}
	if stock() != -1 {
		t.Fatalf("stock %d, want -1", stock())
	}
	if n := f.count(t, `SELECT COUNT(*) FROM audit_log
		WHERE entity_type = 'sale' AND action = 'oversold' AND actor_id = $1`, f.uid); n != 1 {
		t.Fatalf("want one oversold audit entry, got %d", n)
	}
	var atFlag int
	var note string
	if err := f.db.QueryRow(`SELECT stock_at_flag, note FROM stock_flag
		WHERE product_id = $1 AND status = 'open' AND flagged_by_id = $2`, pid, f.uid).
		Scan(&atFlag, &note); err != nil {
		t.Fatal(err)
	}
	if atFlag != -1 || note != "Sold 3 units with 2 units in stock (receipt "+receipt+
		"). Count it and correct the stock." {
		t.Fatalf("flag stock %d, note %q", atFlag, note)
	}

	// A second oversell by the same manager saves, and adds its note to the one open flag rather
	// than being lost; the flag's stock at flagging stays as it was.
	again, err := sales.CreateSale(manager, sellThree("again", true))
	if err != nil {
		t.Fatalf("a second oversell failed on the open flag: %v", err)
	}
	if n := f.count(t, `SELECT COUNT(*) FROM stock_flag WHERE product_id = $1`, pid); n != 1 {
		t.Fatalf("want one flag, got %d", n)
	}
	if err := f.db.QueryRow(`SELECT stock_at_flag, note FROM stock_flag WHERE product_id = $1`,
		pid).Scan(&atFlag, &note); err != nil {
		t.Fatal(err)
	}
	if atFlag != -1 || !strings.Contains(note, "receipt "+receipt+")") ||
		!strings.Contains(note, "Sold 3 units with -1 units in stock (receipt "+again+")") {
		t.Fatalf("flag stock %d, note %q", atFlag, note)
	}

	// A retry of a saved sale is answered as saved, whatever the stock now.
	if got, err := sales.CreateSale(asUser(), sellThree("oversell", false)); err != nil ||
		got != receipt {
		t.Fatalf("a retry should return %s, got %q %v", receipt, got, err)
	}
}

func TestTheOversellPermissionIsSeededToManagersAndAdmins(t *testing.T) {
	f := newFixture(t)
	var roles []string
	err := f.db.Select(&roles, `
		SELECT r.name FROM roles r
		JOIN role_permissions rp ON rp.role_id = r.id
		JOIN permissions p ON p.id = rp.permission_id
		WHERE p.resource = 'Sale' AND p.action = 'Oversell' ORDER BY r.name`)
	if err != nil {
		t.Fatal(err)
	}
	if len(roles) != 2 || roles[0] != "Admin" || roles[1] != "Manager" {
		t.Fatalf("roles with Sale/Oversell: %v", roles)
	}
}
