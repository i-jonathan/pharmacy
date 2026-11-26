package service

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"pharmacy/httperror"
	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"pharmacy/model"
	"pharmacy/repository"
	"testing"

	"github.com/jmoiron/sqlx"
)

// customerSaleRepo records what a sale writes, with customers held in memory.
type customerSaleRepo struct {
	repository.PharmacyRepository

	customers   map[int]types.Customer
	duplicate   bool  // the retry key was used before: the insert adds nothing
	savedFor    *int  // the customer on that earlier save
	insertError error // the insert fails, for example an unknown customer
	sales       []model.Sale
	sawCustomer []int // customers read to check them
	items       int
	begins      int
	commits     int
	rollbacks   int

	// for the stock check: stock by product (nil is plenty of everything)
	stock     map[int]int
	locked    [][]int
	events    []string // "lock" and "price", in the order they happened
	audits    []types.AuditEntry
	saleFlags []saleFlag

	// for the history
	saleRows     []model.Sale
	namesAsked   [][]int
	namesError   error
	namesByID    map[int]string
	namesFetched int
}

func (r *customerSaleRepo) BeginTx(
	context.Context,
) (*sqlx.Tx, error) {
	r.begins++
	return nil, nil
}
func (r *customerSaleRepo) CommitTx(*sqlx.Tx) error { r.commits++; return nil }
func (r *customerSaleRepo) RollbackTx(*sqlx.Tx)     { r.rollbacks++ }

func (r *customerSaleRepo) CreateSaleTx(_ context.Context, _ *sqlx.Tx, s model.Sale) (int, error) {
	if r.insertError != nil {
		return 0, r.insertError
	}
	r.sales = append(r.sales, s)
	if r.duplicate {
		return 0, nil
	}
	return 9, nil
}

func (r *customerSaleRepo) FetchSaleByKeyTx(
	context.Context, *sqlx.Tx, string,
) (types.SavedSale, error) {
	if !r.duplicate {
		return types.SavedSale{}, sql.ErrNoRows // the key was never used
	}
	return types.SavedSale{ReceiptNumber: "2490-7777", CustomerID: r.savedFor}, nil
}

func (r *customerSaleRepo) GetCustomerRefTx(
	_ context.Context, _ *sqlx.Tx, id int,
) (types.CustomerRef, error) {
	r.sawCustomer = append(r.sawCustomer, id)
	c, ok := r.customers[id]
	if !ok {
		return types.CustomerRef{}, sql.ErrNoRows
	}
	return types.CustomerRef{ID: c.ID, Name: c.Name, Active: c.Active}, nil
}

// FetchPriceByID answers price 1 of product 1 at ₦100.00, which is what saleFor sells.
func (r *customerSaleRepo) FetchPriceByID(context.Context, int) (model.ProductPrice, error) {
	return model.ProductPrice{ProductID: 1, QuantityPerUnit: 1, SellingPriceKobo: 10000}, nil
}

func (r *customerSaleRepo) BulkCreateSaleItemsTx(
	context.Context,
	*sqlx.Tx,
	[]model.SaleItem,
) error {
	r.items++
	return nil
}

func (r *customerSaleRepo) BulkCreateSalePaymentsTX(
	context.Context, *sqlx.Tx, []model.SalePayment,
) error {
	return nil
}

func (r *customerSaleRepo) BulkCreateStockMovementTx(
	context.Context, *sqlx.Tx, []model.StockMovement,
) error {
	return nil
}

func (r *customerSaleRepo) FetchSalesTx(
	context.Context, *sqlx.Tx, types.SaleFilter,
) ([]model.Sale, error) {
	return r.saleRows, nil
}

func (r *customerSaleRepo) CountSales(context.Context, types.SaleFilter) (int, error) {
	return len(r.saleRows), nil
}

func (r *customerSaleRepo) BulkFetchSaleItems(
	context.Context, *sqlx.Tx, []int,
) ([]model.SaleItem, error) {
	return nil, nil
}

func (r *customerSaleRepo) BulkFetchSalePayments(
	context.Context, *sqlx.Tx, []int,
) ([]model.SalePayment, error) {
	return nil, nil
}

func (r *customerSaleRepo) BulkFetchReturnItemsBySaleIDs(
	context.Context, []int,
) ([]model.ReturnItemWithSale, error) {
	return nil, nil
}

func (r *customerSaleRepo) BulkFetchProductByIDsTx(
	context.Context, *sqlx.Tx, []int,
) ([]model.Product, error) {
	return nil, nil
}

func (r *customerSaleRepo) BulkFetchUserByIDTx(
	context.Context, *sqlx.Tx, []int,
) ([]model.User, error) {
	return nil, nil
}

func (r *customerSaleRepo) FetchCustomerNames(
	_ context.Context, _ *sqlx.Tx, ids []int,
) (map[int]string, error) {
	r.namesFetched++
	r.namesAsked = append(r.namesAsked, ids)
	return r.namesByID, r.namesError
}

func saleFor(customerID int) types.Sale {
	return types.Sale{
		CashierID:      1,
		IdempotencyKey: "sale-key-1",
		CustomerID:     customerID,
		Subtotal:       100,
		Total:          100,
		Items: []types.SaleItem{
			{ProductID: 1, Quantity: 1, PriceID: 1, UnitPrice: 100, Total: 100},
		},
		Payments: []types.SalePayment{{Amount: 100, PaymentMethod: "Cash"}},
	}
}

func activeCustomers() map[int]types.Customer {
	return map[int]types.Customer{5: {ID: 5, Name: "Ada Obi", Active: true}}
}

func TestSaleWithoutACustomerIsAWalkIn(t *testing.T) {
	repo := &customerSaleRepo{}
	// Nobody needs customer permissions to sell to a walk-in.
	receipt, err := NewSaleService(repo).CreateSale(contextWith(), saleFor(0))
	if err != nil || receipt == "" {
		t.Fatalf("receipt %q, err %v", receipt, err)
	}
	if len(repo.sales) != 1 || repo.sales[0].CustomerID != nil {
		t.Fatalf("a walk-in sale stores no customer: %+v", repo.sales)
	}
	if len(repo.sawCustomer) != 0 || repo.commits != 1 {
		t.Fatalf("customers read %v, commits %d", repo.sawCustomer, repo.commits)
	}
}

func TestSaleStoresTheChosenCustomer(t *testing.T) {
	repo := &customerSaleRepo{customers: activeCustomers()}
	_, err := NewSaleService(repo).CreateSale(
		contextWith(constant.ViewCustomersPermissionKey), saleFor(5),
	)
	if err != nil {
		t.Fatal(err)
	}
	if got := repo.sales[0].CustomerID; got == nil || *got != 5 {
		t.Fatalf("the sale should carry customer 5, got %v", got)
	}
	if repo.commits != 1 || repo.items != 1 {
		t.Fatalf("commits %d, item batches %d", repo.commits, repo.items)
	}
}

func TestSaleWithACustomerNeedsPermissionToSeeCustomers(t *testing.T) {
	repo := &customerSaleRepo{customers: activeCustomers()}
	_, err := NewSaleService(repo).CreateSale(contextWith(), saleFor(5))
	requireHTTPStatus(t, err, http.StatusForbidden)
	// The key is looked up first, so a retry of a saved sale can still be answered; this one is new,
	// so nothing is inserted and nothing is committed.
	if len(repo.sales) != 0 || repo.commits != 0 || repo.rollbacks != 1 {
		t.Fatalf(
			"sales %d, commits %d, rollbacks %d",
			len(repo.sales),
			repo.commits,
			repo.rollbacks,
		)
	}
}

func TestRetryOfASavedSaleIsAnsweredEvenIfTheUserCanNoLongerSeeCustomers(t *testing.T) {
	five := 5
	repo := &customerSaleRepo{
		customers: activeCustomers(), duplicate: true, savedFor: &five,
	}
	// Their right to see customers was removed after the first save; the reply was lost.
	receipt, err := NewSaleService(repo).CreateSale(contextWith(), saleFor(5))
	if err != nil || receipt != "2490-7777" {
		t.Fatalf("a retry should answer with the first save: %q, %v", receipt, err)
	}
	if len(repo.sawCustomer) != 0 || len(repo.sales) != 1 {
		t.Fatalf("customers read %v, inserts tried %d", repo.sawCustomer, len(repo.sales))
	}
}

func TestSaleWithoutAKeyForACustomerTheUserCannotSeeIsRefused(t *testing.T) {
	repo := &customerSaleRepo{customers: activeCustomers(), duplicate: true}
	sale := saleFor(5)
	sale.IdempotencyKey = ""
	_, err := NewSaleService(repo).CreateSale(contextWith(), sale)
	requireHTTPStatus(t, err, http.StatusForbidden)
	if len(repo.sales) != 0 || repo.commits != 0 {
		t.Fatal("nothing should be saved")
	}
}

func TestSaleRefusesANegativeCustomerID(t *testing.T) {
	repo := &customerSaleRepo{}
	_, err := NewSaleService(repo).CreateSale(
		contextWith(constant.ViewCustomersPermissionKey), saleFor(-3),
	)
	requireHTTPStatus(t, err, http.StatusBadRequest)
	if repo.begins != 0 {
		t.Fatal("nothing should be started")
	}
}

func TestSaleForAnUnknownCustomerIsRefusedAndNothingIsSaved(t *testing.T) {
	repo := &customerSaleRepo{insertError: repository.ErrInvalidReference}
	_, err := NewSaleService(repo).CreateSale(
		contextWith(constant.ViewCustomersPermissionKey), saleFor(77),
	)
	requireHTTPStatus(t, err, http.StatusBadRequest)
	if repo.commits != 0 || repo.rollbacks != 1 {
		t.Fatalf("commits %d, rollbacks %d", repo.commits, repo.rollbacks)
	}
}

func TestSaleForADeactivatedCustomerIsRefusedAndRolledBack(t *testing.T) {
	gone := types.Customer{ID: 5, Name: "Ada Obi", Active: false}
	repo := &customerSaleRepo{customers: map[int]types.Customer{5: gone}}
	_, err := NewSaleService(repo).CreateSale(
		contextWith(constant.ViewCustomersPermissionKey), saleFor(5),
	)
	requireHTTPStatus(t, err, http.StatusConflict)
	var httperr *httperror.HTTPError
	errors.As(err, &httperr)
	if httperr.Reason != "customer_inactive" {
		t.Fatalf("the page needs a code to act on, got %q", httperr.Reason)
	}
	if repo.commits != 0 || repo.rollbacks != 1 || repo.items != 0 {
		t.Fatalf("commits %d, rollbacks %d, items %d", repo.commits, repo.rollbacks, repo.items)
	}
}

func TestRetryOfASavedSaleIsNotRefusedBecauseTheCustomerWasDeactivatedSince(t *testing.T) {
	gone := types.Customer{ID: 5, Name: "Ada Obi", Active: false}
	five := 5
	repo := &customerSaleRepo{
		customers: map[int]types.Customer{5: gone}, duplicate: true, savedFor: &five,
	}
	receipt, err := NewSaleService(repo).CreateSale(
		contextWith(constant.ViewCustomersPermissionKey), saleFor(5),
	)
	if err != nil || receipt != "2490-7777" {
		t.Fatalf("a retry should answer with the first save: %q, %v", receipt, err)
	}
	if len(repo.sawCustomer) != 0 {
		t.Fatal("a retry must not check the customer again")
	}
}

func TestRetryUnderADifferentCustomerIsRefusedNotAnsweredWithTheFirstSave(t *testing.T) {
	five := 5
	for name, tc := range map[string]struct {
		saved   *int
		request int
	}{
		"another customer":        {&five, 6},
		"a customer, was walk-in": {nil, 5},
		"walk-in, had a customer": {&five, 0},
	} {
		repo := &customerSaleRepo{duplicate: true, savedFor: tc.saved}
		ctx := contextWith(constant.ViewCustomersPermissionKey)
		_, err := NewSaleService(repo).CreateSale(ctx, saleFor(tc.request))
		if err == nil {
			t.Errorf("%s: the retry must not be answered as if it were the first save", name)
			continue
		}
		requireHTTPStatus(t, err, http.StatusConflict)
		var httperr *httperror.HTTPError
		errors.As(err, &httperr)
		if httperr.Reason != "already_saved" {
			t.Errorf("%s: code %q", name, httperr.Reason)
		}
	}
}

func historyOf(customerID *int) *customerSaleRepo {
	return &customerSaleRepo{
		saleRows: []model.Sale{
			{ReceiptNumber: "A-1", CustomerID: customerID},
			{ReceiptNumber: "A-2"},
		},
		namesByID: map[int]string{5: "Ada Obi"},
	}
}

func TestSalesHistoryNamesTheCustomerForUsersWhoMaySeeCustomers(t *testing.T) {
	five := 5
	repo := historyOf(&five)
	ctx := contextWith(constant.ViewCustomersPermissionKey)
	history, err := NewSaleService(
		repo,
	).FetchSalesHistory(ctx, types.SaleFilter{Page: 1, PerPage: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(history.Data) != 2 {
		t.Fatalf("expected 2 sales, got %d", len(history.Data))
	}
	got := history.Data[0].Customer
	if got == nil || got.ID != 5 || got.Name != "Ada Obi" {
		t.Fatalf("the first sale should name its customer: %+v", got)
	}
	if history.Data[1].Customer != nil {
		t.Fatalf("a walk-in sale has no customer: %+v", history.Data[1].Customer)
	}
	if len(repo.namesAsked) != 1 || len(repo.namesAsked[0]) != 1 || repo.namesAsked[0][0] != 5 {
		t.Fatalf("one lookup for the customers on the page, got %v", repo.namesAsked)
	}
}

func TestSalesHistoryHidesCustomersFromEveryoneElse(t *testing.T) {
	five := 5
	repo := historyOf(&five)
	history, err := NewSaleService(repo).FetchSalesHistory(
		contextWith(), types.SaleFilter{Page: 1, PerPage: 10},
	)
	if err != nil {
		t.Fatal(err)
	}
	if history.Data[0].Customer != nil || repo.namesFetched != 0 {
		t.Fatalf("customers must not reach a user without customer:view: %+v, lookups %d",
			history.Data[0].Customer, repo.namesFetched)
	}
}

func TestSalesHistoryRollsBackWhenTheCustomerLookupFails(t *testing.T) {
	five := 5
	repo := historyOf(&five)
	repo.namesError = errors.New("connection reset")
	ctx := contextWith(constant.ViewCustomersPermissionKey)
	_, err := NewSaleService(repo).FetchSalesHistory(ctx, types.SaleFilter{Page: 1, PerPage: 10})
	requireHTTPStatus(t, err, http.StatusInternalServerError)
	if repo.rollbacks != 1 || repo.commits != 0 {
		t.Fatalf("the open transaction must be given back: rollbacks %d, commits %d",
			repo.rollbacks, repo.commits)
	}
}

func TestSalesHistoryWithNoCustomersLooksNothingUp(t *testing.T) {
	repo := historyOf(nil)
	ctx := contextWith(constant.ViewCustomersPermissionKey)
	_, err := NewSaleService(repo).FetchSalesHistory(ctx, types.SaleFilter{Page: 1, PerPage: 10})
	if err != nil || repo.namesFetched != 0 {
		t.Fatalf("err %v, lookups %d", err, repo.namesFetched)
	}
}

func (r *customerSaleRepo) FetchSaleCustomerChanges(
	context.Context, *sqlx.Tx, []int,
) ([]types.SaleCustomerChangeRow, error) {
	return nil, nil
}
