package service_test

import (
	"net/http"
	"testing"

	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"pharmacy/repository"
	"pharmacy/service"
)

// These run the real service and SQL against a throwaway database; see internal/testdb.

func TestSaleWithACustomerAgainstTheDatabase(t *testing.T) {
	f := newFixture(t)
	store := repository.NewStore(f.db)
	customers := service.NewCustomerService(store)
	sales := service.NewSaleService(store)
	product := f.stocked(t, "Amoxil", 100)
	var priceID int
	if err := f.db.Get(
		&priceID,
		`SELECT id FROM product_price WHERE product_id = $1`,
		product,
	); err != nil {
		t.Fatal(err)
	}
	seeCustomers := asUser(
		constant.ViewCustomersPermissionKey,
		constant.ManageCustomersPermissionKey,
	)

	ada, err := customers.CreateCustomer(seeCustomers, types.CustomerRequest{
		Name: "Ada Obi", Phone: "0803 555 0101",
	}, f.uid)
	if err != nil {
		t.Fatal(err)
	}
	sale := func(key string, customer int) types.Sale {
		return types.Sale{
			CashierID: f.uid, IdempotencyKey: key, CustomerID: customer, Subtotal: 150, Total: 150,
			Items: []types.SaleItem{{
				ProductID: product, Quantity: 1, PriceID: priceID, UnitPrice: 150, Total: 150,
			}},
			Payments: []types.SalePayment{{Amount: 150, PaymentMethod: "Cash"}},
		}
	}

	withCustomer, err := sales.CreateSale(seeCustomers, sale("with-customer", ada))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = sales.CreateSale(asUser(), sale("walk-in", 0)); err != nil {
		t.Fatal(err)
	}
	// A user who may not see customers can't attach one, and nothing is saved.
	_, err = sales.CreateSale(asUser(), sale("no-permission", ada))
	requireStatus(t, err, http.StatusForbidden)
	// A foreign key that is not the customer's (a cashier who no longer exists) is not blamed on
	// the customer.
	badCashier := sale("bad-cashier", ada)
	badCashier.CashierID = f.uid + 999
	_, err = sales.CreateSale(seeCustomers, badCashier)
	requireStatus(t, err, http.StatusInternalServerError)
	// An unknown customer is refused cleanly and leaves no sale behind.
	_, err = sales.CreateSale(seeCustomers, sale("unknown", ada+500))
	requireStatus(t, err, http.StatusBadRequest)
	if n := f.count(t, `SELECT count(*) FROM sales`); n != 2 {
		t.Fatalf("only the two good sales should exist, found %d", n)
	}

	// History names the customer for a user who may see customers, and only for them.
	filter := types.SaleFilter{Page: 1, PerPage: 10}
	seen, err := sales.FetchSalesHistory(seeCustomers, filter)
	if err != nil {
		t.Fatal(err)
	}
	named := 0
	for _, s := range seen.Data {
		if s.Customer != nil {
			named++
			if s.ReceiptNumber != withCustomer || s.Customer.ID != ada ||
				s.Customer.Name != "Ada Obi" {
				t.Fatalf("wrong customer on %s: %+v", s.ReceiptNumber, s.Customer)
			}
		}
	}
	if len(seen.Data) != 2 || named != 1 {
		t.Fatalf("expected 2 sales with 1 named, got %d with %d", len(seen.Data), named)
	}
	hidden, err := sales.FetchSalesHistory(asUser(), filter)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range hidden.Data {
		if s.Customer != nil {
			t.Fatalf("a user without customer:view must not see customers: %+v", s.Customer)
		}
	}

	// Deactivating the customer refuses a new sale for them, but not a retry of the saved one,
	// and their earlier sale still names them.
	cust, err := customers.GetCustomer(seeCustomers, ada)
	if err != nil {
		t.Fatal(err)
	}
	if err = customers.SetCustomerActive(
		seeCustomers,
		ada,
		cust.Version,
		false,
		f.uid,
	); err != nil {
		t.Fatal(err)
	}
	_, err = sales.CreateSale(seeCustomers, sale("after-deactivation", ada))
	requireStatus(t, err, http.StatusConflict)
	if n := f.count(t, `SELECT count(*) FROM sales`); n != 2 {
		t.Fatalf("a refused sale must leave nothing behind, found %d sales", n)
	}
	retry, err := sales.CreateSale(seeCustomers, sale("with-customer", ada))
	if err != nil || retry != withCustomer {
		t.Fatalf("a retry should return the first save %q, got %q, %v", withCustomer, retry, err)
	}
	// A retry that names someone else is not the saved sale: refused, not answered with the first save.
	_, err = sales.CreateSale(seeCustomers, sale("with-customer", 0))
	requireStatus(t, err, http.StatusConflict)
	after, _ := sales.FetchSalesHistory(seeCustomers, filter)
	found := false
	for _, s := range after.Data {
		found = found || (s.Customer != nil && s.Customer.Name == "Ada Obi")
	}
	if !found {
		t.Fatal("a deactivated customer is still named on their earlier sale")
	}
}

func TestChangingTheCustomerOnASavedSaleAgainstTheDatabase(t *testing.T) {
	f := newFixture(t)
	store := repository.NewStore(f.db)
	customers := service.NewCustomerService(store)
	sales := service.NewSaleService(store)
	product := f.stocked(t, "Amoxil", 100)
	var priceID int
	if err := f.db.Get(
		&priceID, `SELECT id FROM product_price WHERE product_id = $1`, product,
	); err != nil {
		t.Fatal(err)
	}
	cashier := asUser(constant.ViewCustomersPermissionKey, constant.ManageCustomersPermissionKey)
	manager := asUser(
		constant.ViewCustomersPermissionKey,
		constant.ManageCustomersPermissionKey,
		constant.ChangeOldSaleCustomerPermissionKey,
	)
	newCustomer := func(name, phone string) int {
		id, err := customers.CreateCustomer(
			cashier, types.CustomerRequest{Name: name, Phone: phone}, f.uid,
		)
		if err != nil {
			t.Fatal(err)
		}
		return id
	}
	ada, bola, chidi := newCustomer("Ada Obi", "0803 555 0101"),
		newCustomer("Bola Ade", "0803 555 0102"), newCustomer("Chidi Eze", "0803 555 0103")
	sell := func(key string, customer int) (receipt string, id int) {
		receipt, err := sales.CreateSale(cashier, types.Sale{
			CashierID: f.uid, IdempotencyKey: key, CustomerID: customer, Subtotal: 150, Total: 150,
			Items: []types.SaleItem{{
				ProductID: product, Quantity: 1, PriceID: priceID, UnitPrice: 150, Total: 150,
			}},
			Payments: []types.SalePayment{{Amount: 150, PaymentMethod: "Cash"}},
		})
		if err != nil {
			t.Fatal(err)
		}
		if err = f.db.Get(
			&id, `SELECT id FROM sales WHERE receipt_number = $1`, receipt,
		); err != nil {
			t.Fatal(err)
		}
		return receipt, id
	}
	customerOf := func(saleID int) (id int) {
		if err := f.db.Get(
			&id, `SELECT COALESCE(customer_id, 0) FROM sales WHERE id = $1`, saleID,
		); err != nil {
			t.Fatal(err)
		}
		return id
	}
	req := func(sale, to, expected int, reason string) types.ChangeSaleCustomerRequest {
		return types.ChangeSaleCustomerRequest{
			SaleID: sale, CustomerID: to, ExpectedCustomerID: expected, Reason: reason,
		}
	}

	// The permission is seeded by the migration to managers and admins, not to cashiers.
	var holders []string
	if err := f.db.Select(&holders, `
		SELECT r.name FROM role_permissions rp
		JOIN roles r ON r.id = rp.role_id
		JOIN permissions p ON p.id = rp.permission_id
		WHERE p.resource = 'Sale' AND p.action = 'ChangeOldCustomer' ORDER BY r.name`,
	); err != nil {
		t.Fatal(err)
	}
	if len(holders) != 2 || holders[0] != "Admin" || holders[1] != "Manager" {
		t.Fatalf("only Admin and Manager get the old-sales permission, got %v", holders)
	}

	receipt, saleID := sell("one", ada)

	// Change to another customer: saved, audited with ids and the reason, and History shows it.
	got, err := sales.ChangeSaleCustomer(
		cashier, req(saleID, bola, ada, "picked_by_mistake"), f.uid,
	)
	if err != nil || got == nil || got.ID != bola || got.Name != "Bola Ade" {
		t.Fatalf("change: %+v, %v", got, err)
	}
	if customerOf(saleID) != bola {
		t.Fatal("the sale should name Bola")
	}
	if n := f.count(t, `SELECT count(*) FROM audit_log WHERE entity_type = 'sale'
		AND entity_id = $1 AND action = 'customer_changed' AND reason = 'picked_by_mistake'
		AND details->>'receipt_number' = $2
		AND (details->>'from_customer_id')::int = $3 AND (details->>'to_customer_id')::int = $4`,
		saleID, receipt, ada, bola); n != 1 {
		t.Fatalf("expected one audit entry, found %d", n)
	}
	history, err := sales.FetchSalesHistory(cashier, types.SaleFilter{Page: 1, PerPage: 10})
	if err != nil || len(history.Data) != 1 || history.Data[0].Customer == nil ||
		history.Data[0].Customer.Name != "Bola Ade" {
		t.Fatalf("History should show Bola: %+v, %v", history.Data, err)
	}
	// The server says which sales are less than 24 hours old, on the same test as the change rule,
	// so the page never has to guess with the browser's clock.
	if !history.Data[0].IsRecent {
		t.Fatal("a sale just made is within the 24 hours")
	}

	// The page that still thinks it is Ada is refused and nothing moves.
	_, err = sales.ChangeSaleCustomer(cashier, req(saleID, chidi, ada, "picked_by_mistake"), f.uid)
	requireStatus(t, err, http.StatusConflict)
	// A repeat of the change that was just made writes nothing more.
	if _, err = sales.ChangeSaleCustomer(
		cashier, req(saleID, bola, bola, "picked_by_mistake"), f.uid,
	); err != nil {
		t.Fatal(err)
	}
	if n := f.count(t, `SELECT count(*) FROM audit_log WHERE action = 'customer_changed'`); n != 1 {
		t.Fatalf("only the one change is audited, found %d", n)
	}

	// A deactivated customer cannot be attached, but can be removed.
	cust, err := customers.GetCustomer(cashier, chidi)
	if err != nil {
		t.Fatal(err)
	}
	if err = customers.SetCustomerActive(cashier, chidi, cust.Version, false, f.uid); err != nil {
		t.Fatal(err)
	}
	_, err = sales.ChangeSaleCustomer(cashier, req(saleID, chidi, bola, "customer_asked"), f.uid)
	requireStatus(t, err, http.StatusConflict)
	_, err = sales.ChangeSaleCustomer(cashier, req(saleID, bola+900, bola, "customer_asked"), f.uid)
	requireStatus(t, err, http.StatusBadRequest)

	// Remove the customer: the sale becomes a walk-in.
	got, err = sales.ChangeSaleCustomer(cashier, req(saleID, 0, bola, "customer_asked"), f.uid)
	if err != nil || got != nil || customerOf(saleID) != 0 {
		t.Fatalf("removal: %+v, %v, customer %d", got, err, customerOf(saleID))
	}

	// The rule is a rolling 24 hours, not the calendar day: a sale 23 hours old is still recent,
	// even if that was yesterday, so a cashier can fix it.
	f.db.MustExec(
		`UPDATE sales SET created_at = now() - INTERVAL '23 hours' WHERE id = $1`, saleID,
	)
	recent, err := sales.FetchSalesHistory(cashier, types.SaleFilter{Page: 1, PerPage: 10})
	if err != nil || len(recent.Data) != 1 || !recent.Data[0].IsRecent {
		t.Fatalf("a sale 23 hours old is still recent: %+v, %v", recent.Data, err)
	}
	if _, err = sales.ChangeSaleCustomer(
		cashier, req(saleID, ada, 0, "picked_by_mistake"), f.uid,
	); err != nil || customerOf(saleID) != ada {
		t.Fatalf("a cashier may change a 23 hour old sale: %v, now %d", err, customerOf(saleID))
	}
	if _, err = sales.ChangeSaleCustomer(
		cashier, req(saleID, 0, ada, "picked_by_mistake"), f.uid,
	); err != nil || customerOf(saleID) != 0 {
		t.Fatalf("and take it off again: %v, customer %d", err, customerOf(saleID))
	}

	// A sale more than 24 hours old: a cashier is refused, a manager is not.
	f.db.MustExec(
		`UPDATE sales SET created_at = now() - INTERVAL '25 hours' WHERE id = $1`, saleID,
	)
	earlier, err := sales.FetchSalesHistory(cashier, types.SaleFilter{Page: 1, PerPage: 10})
	if err != nil || len(earlier.Data) != 1 || earlier.Data[0].IsRecent {
		t.Fatalf("a sale 25 hours old is not recent: %+v, %v", earlier.Data, err)
	}
	_, err = sales.ChangeSaleCustomer(cashier, req(saleID, ada, 0, "picked_by_mistake"), f.uid)
	requireStatus(t, err, http.StatusForbidden)
	if customerOf(saleID) != 0 {
		t.Fatal("a refused change must not move the customer")
	}
	if _, err = sales.ChangeSaleCustomer(
		manager, req(saleID, ada, 0, "picked_by_mistake"), f.uid,
	); err != nil || customerOf(saleID) != ada {
		t.Fatalf("a manager may change an earlier sale: %v, customer %d", err, customerOf(saleID))
	}

	// A retry of that change after the sale has aged out is answered as saved for a cashier too.
	if got, err = sales.ChangeSaleCustomer(
		cashier, req(saleID, ada, 0, "picked_by_mistake"), f.uid,
	); err != nil || got == nil || got.ID != ada {
		t.Fatalf("a retry on an old sale is answered as saved: %+v, %v", got, err)
	}

	// Missing sale.
	_, err = sales.ChangeSaleCustomer(manager, req(saleID+500, ada, 0, "picked_by_mistake"), f.uid)
	requireStatus(t, err, http.StatusNotFound)
}
