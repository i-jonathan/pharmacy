package service_test

import (
	"testing"

	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"pharmacy/repository"
	"pharmacy/service"
)

// Runs the real service and SQL against a throwaway database; see internal/testdb.

func TestSalesHistoryListsCustomerChangesAgainstTheDatabase(t *testing.T) {
	f := newFixture(t)
	store := repository.NewStore(f.db)
	customers := service.NewCustomerService(store)
	sales := service.NewSaleService(store)
	product := f.product(t, "Amoxil")
	var priceID int
	if err := f.db.Get(
		&priceID, `SELECT id FROM product_price WHERE product_id = $1`, product,
	); err != nil {
		t.Fatal(err)
	}
	seeAndManage := asUser(
		constant.ViewCustomersPermissionKey,
		constant.ManageCustomersPermissionKey,
	)
	seeOnly := asUser(constant.ViewCustomersPermissionKey)

	newCustomer := func(name, phone string) int {
		id, err := customers.CreateCustomer(
			seeAndManage, types.CustomerRequest{Name: name, Phone: phone}, f.uid,
		)
		if err != nil {
			t.Fatal(err)
		}
		return id
	}
	ada, bola := newCustomer("Ada Obi", "0803 555 0101"), newCustomer("Bola Ade", "0803 555 0102")
	sell := func(key string, customer int) int {
		receipt, err := sales.CreateSale(seeAndManage, types.Sale{
			CashierID: f.uid, IdempotencyKey: key, CustomerID: customer, Subtotal: 150, Total: 150,
			Items: []types.SaleItem{{
				ProductID: product, Quantity: 1, PriceID: priceID, UnitPrice: 150, Total: 150,
			}},
			Payments: []types.SalePayment{{Amount: 150, PaymentMethod: "Cash"}},
		})
		if err != nil {
			t.Fatal(err)
		}
		var id int
		if err = f.db.Get(
			&id,
			`SELECT id FROM sales WHERE receipt_number = $1`,
			receipt,
		); err != nil {
			t.Fatal(err)
		}
		return id
	}
	change := func(sale, to, expected int, reason, note string) {
		_, err := sales.ChangeSaleCustomer(seeAndManage, types.ChangeSaleCustomerRequest{
			SaleID: sale, CustomerID: to, ExpectedCustomerID: expected, Reason: reason, Note: note,
		}, f.uid)
		if err != nil {
			t.Fatal(err)
		}
	}

	changed, other, untouched := sell("one", ada), sell("two", bola), sell("three", 0)
	change(changed, bola, ada, "picked_by_mistake", "")
	change(changed, 0, bola, "other", "she wanted no customer on it")
	change(other, ada, bola, "customer_asked", "")
	// A reprint is also audited against the sale; it is not a customer change.
	if _, err := sales.ReprintReceipt(seeOnly, changed, f.uid); err != nil {
		t.Fatal(err)
	}

	history, err := sales.FetchSalesHistory(seeOnly, types.SaleFilter{Page: 1, PerPage: 10})
	if err != nil {
		t.Fatal(err)
	}
	bySale := map[int]types.SaleResponse{}
	for _, s := range history.Data {
		bySale[s.ID] = s
	}

	got := bySale[changed].CustomerChanges
	if len(got) != 2 {
		t.Fatalf("the sale was changed twice and reprinted once, listed: %+v", got)
	}
	// Newest first: the removal, then the first change.
	removal, first := got[0], got[1]
	if removal.To != nil || removal.From == nil || removal.From.Name != "Bola Ade" ||
		removal.Reason != "other" || removal.Note != "she wanted no customer on it" {
		t.Fatalf("newest should be the removal: %+v", removal)
	}
	if first.From == nil || first.From.Name != "Ada Obi" || first.To == nil ||
		first.To.ID != bola || first.To.Name != "Bola Ade" ||
		first.Reason != "picked_by_mistake" || first.Note != "" {
		t.Fatalf("oldest should be Ada -> Bola: %+v", first)
	}
	var username string
	if err = f.db.Get(&username, `SELECT username FROM users WHERE id = $1`, f.uid); err != nil {
		t.Fatal(err)
	}
	if removal.By != username || first.By != username || removal.At == "" {
		t.Fatalf("who and when must be filled in: %+v", got)
	}

	if n := len(bySale[other].CustomerChanges); n != 1 {
		t.Fatalf("the other sale was changed once, listed %d", n)
	}
	if c := bySale[untouched].CustomerChanges; c == nil || len(c) != 0 {
		t.Fatalf("an unchanged sale lists nothing: %+v", c)
	}

	// Someone who may not see customers gets no history of them.
	hidden, err := sales.FetchSalesHistory(asUser(), types.SaleFilter{Page: 1, PerPage: 10})
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range hidden.Data {
		if len(s.CustomerChanges) != 0 {
			t.Fatalf("customer changes leaked to a user without customer:view: %+v", s)
		}
	}
}
