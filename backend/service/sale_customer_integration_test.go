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
	product := f.product(t, "Amoxil")
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
