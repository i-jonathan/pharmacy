package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"pharmacy/httperror"
	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
)

// changeRepo is a saved sale and its customers held in memory.
type changeRepo struct {
	customerSaleRepo

	sale      types.SaleCustomerState
	saleError error
	updated   []*int // the customer each update set
	audits    []types.AuditEntry
}

func (r *changeRepo) FetchSaleCustomerForUpdateTx(
	context.Context, *sqlx.Tx, int,
) (types.SaleCustomerState, error) {
	return r.sale, r.saleError
}

func (r *changeRepo) UpdateSaleCustomerTx(
	_ context.Context, _ *sqlx.Tx, _ int, customerID *int,
) error {
	r.updated = append(r.updated, customerID)
	return nil
}

func (r *changeRepo) InsertAuditLogTx(
	_ context.Context, _ *sqlx.Tx, entry types.AuditEntry,
) error {
	r.audits = append(r.audits, entry)
	return nil
}

// aSaleOf is a recent sale (under 24 hours old) that names the given customer (0 for a walk-in).
func aSaleOf(customerID int) types.SaleCustomerState {
	state := types.SaleCustomerState{ReceiptNumber: "2490-1234", IsRecent: true}
	if customerID > 0 {
		state.CustomerID = intPtr(customerID)
	}
	return state
}

func changeTo(to, expected int) types.ChangeSaleCustomerRequest {
	return types.ChangeSaleCustomerRequest{
		SaleID: 9, CustomerID: to, ExpectedCustomerID: expected, Reason: "picked_by_mistake",
	}
}

// aCashier may change recent sales; aManager may change any.
func aCashier() context.Context {
	return contextWith(constant.ViewCustomersPermissionKey, constant.ManageCustomersPermissionKey)
}

func aManager() context.Context {
	return contextWith(
		constant.ViewCustomersPermissionKey,
		constant.ManageCustomersPermissionKey,
		constant.ChangeOldSaleCustomerPermissionKey,
	)
}

func requireReason(t *testing.T, err error, want string) {
	t.Helper()
	var httperr *httperror.HTTPError
	if !errors.As(err, &httperr) || httperr.Reason != want {
		t.Fatalf("expected code %q, got %v", want, err)
	}
}

func twoCustomers() map[int]types.Customer {
	return map[int]types.Customer{
		5: {ID: 5, Name: "Ada Obi", Active: true},
		6: {ID: 6, Name: "Bola Ade", Active: true},
		7: {ID: 7, Name: "Chidi Eze", Active: false},
	}
}

func TestChangingASaleCustomerNeedsPermissionToManageCustomers(t *testing.T) {
	repo := &changeRepo{sale: aSaleOf(5)}
	ctx := contextWith(constant.ViewCustomersPermissionKey)
	_, err := NewSaleService(repo).ChangeSaleCustomer(ctx, changeTo(6, 5), 1)
	requireHTTPStatus(t, err, http.StatusForbidden)
	if repo.begins != 0 {
		t.Fatal("nothing should be started")
	}
}

func TestChangingASaleCustomerNeedsPermissionToSeeCustomers(t *testing.T) {
	// Without it the customer on a sale is never shown, so neither attaching nor removing one is
	// allowed: a user who cannot see who is on a sale cannot be asked to fix it.
	for name, req := range map[string]types.ChangeSaleCustomerRequest{
		"attach": changeTo(6, 0), "remove": changeTo(0, 5),
	} {
		repo := &changeRepo{sale: aSaleOf(5)}
		ctx := contextWith(constant.ManageCustomersPermissionKey)
		_, err := NewSaleService(repo).ChangeSaleCustomer(ctx, req, 1)
		requireHTTPStatus(t, err, http.StatusForbidden)
		if repo.begins != 0 {
			t.Fatalf("%s: nothing should be started", name)
		}
	}
}

func TestChangingASaleCustomerRefusesBadRequests(t *testing.T) {
	bad := map[string]types.ChangeSaleCustomerRequest{
		"no sale":           {SaleID: 0, CustomerID: 6, Reason: "picked_by_mistake"},
		"negative customer": {SaleID: 9, CustomerID: -1, Reason: "picked_by_mistake"},
		"negative expectation": {
			SaleID:             9,
			CustomerID:         6,
			ExpectedCustomerID: -1,
			Reason:             "picked_by_mistake",
		},
		"no reason":            {SaleID: 9, CustomerID: 6},
		"unknown reason":       {SaleID: 9, CustomerID: 6, Reason: "because"},
		"other without a note": {SaleID: 9, CustomerID: 6, Reason: "other", Note: "  "},
		"a long note": {
			SaleID: 9, CustomerID: 6, Reason: "other", Note: strings.Repeat("x", 251),
		},
	}
	for name, req := range bad {
		repo := &changeRepo{
			sale:             aSaleOf(5),
			customerSaleRepo: customerSaleRepo{customers: twoCustomers()},
		}
		_, err := NewSaleService(repo).ChangeSaleCustomer(aManager(), req, 1)
		if err == nil {
			t.Errorf("%s: expected a refusal", name)
			continue
		}
		requireHTTPStatus(t, err, http.StatusBadRequest)
		if repo.begins != 0 || len(repo.updated) != 0 {
			t.Errorf("%s: nothing should be started or saved", name)
		}
	}
}

func TestChangingTheCustomerSavesItAndWritesTheAuditEntry(t *testing.T) {
	repo := &changeRepo{
		sale: aSaleOf(5), customerSaleRepo: customerSaleRepo{customers: twoCustomers()},
	}
	req := changeTo(6, 5)
	req.Reason = "other"
	req.Note = "  Phone number was read out wrongly "
	got, err := NewSaleService(repo).ChangeSaleCustomer(aCashier(), req, 42)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.ID != 6 || got.Name != "Bola Ade" {
		t.Fatalf("the reply names the new customer, got %+v", got)
	}
	if len(repo.updated) != 1 || repo.updated[0] == nil || *repo.updated[0] != 6 {
		t.Fatalf("customer 6 should be saved, got %v", repo.updated)
	}
	if repo.commits != 1 || len(repo.audits) != 1 {
		t.Fatalf("commits %d, audits %d", repo.commits, len(repo.audits))
	}
	audit := repo.audits[0]
	if audit.ActorID != 42 || audit.EntityType != constant.AuditEntitySale || audit.EntityID != 9 ||
		audit.Action != constant.AuditActionSaleCustomerChanged ||
		audit.Reason != "other: Phone number was read out wrongly" {
		t.Fatalf("wrong audit entry: %+v", audit)
	}
	var details map[string]any
	if err := json.Unmarshal(audit.Details, &details); err != nil {
		t.Fatal(err)
	}
	if details["receipt_number"] != "2490-1234" || details["from_customer_id"] != float64(5) ||
		details["to_customer_id"] != float64(6) || len(details) != 3 {
		t.Fatalf("the audit holds the receipt and the two ids only: %v", details)
	}
}

func TestRemovingTheCustomerMakesTheSaleAWalkInEvenIfTheyAreDeactivated(t *testing.T) {
	repo := &changeRepo{
		sale: aSaleOf(7), customerSaleRepo: customerSaleRepo{customers: twoCustomers()},
	}
	got, err := NewSaleService(repo).ChangeSaleCustomer(aCashier(), changeTo(0, 7), 1)
	if err != nil || got != nil {
		t.Fatalf("a walk-in has no customer in the reply: %+v, %v", got, err)
	}
	if len(repo.updated) != 1 || repo.updated[0] != nil {
		t.Fatalf("the customer should be cleared, got %v", repo.updated)
	}
	if len(repo.sawCustomer) != 0 {
		t.Fatal("nobody needs reading to remove a customer")
	}
	var details map[string]any
	_ = json.Unmarshal(repo.audits[0].Details, &details)
	if details["to_customer_id"] != float64(0) || details["from_customer_id"] != float64(7) {
		t.Fatalf("a walk-in is id 0 in the audit: %v", details)
	}
}

func TestAddingACustomerToAWalkInSale(t *testing.T) {
	repo := &changeRepo{
		sale: aSaleOf(0), customerSaleRepo: customerSaleRepo{customers: twoCustomers()},
	}
	if _, err := NewSaleService(
		repo,
	).ChangeSaleCustomer(aCashier(), changeTo(5, 0), 1); err != nil {
		t.Fatal(err)
	}
	if len(repo.updated) != 1 || *repo.updated[0] != 5 {
		t.Fatalf("customer 5 should be saved, got %v", repo.updated)
	}
}

func TestChangeIsRefusedWhenSomeoneElseChangedTheCustomerFirst(t *testing.T) {
	repo := &changeRepo{
		sale: aSaleOf(6), customerSaleRepo: customerSaleRepo{customers: twoCustomers()},
	}
	_, err := NewSaleService(repo).ChangeSaleCustomer(aCashier(), changeTo(5, 3), 1)
	requireHTTPStatus(t, err, http.StatusConflict)
	requireReason(t, err, "customer_changed")
	if len(repo.updated) != 0 || len(repo.audits) != 0 || repo.commits != 0 || repo.rollbacks != 1 {
		t.Fatal("a stale change must leave nothing behind")
	}
}

func TestChangeToTheCustomerAlreadyOnTheSaleWritesNothing(t *testing.T) {
	repo := &changeRepo{
		sale: aSaleOf(5), customerSaleRepo: customerSaleRepo{customers: twoCustomers()},
	}
	got, err := NewSaleService(repo).ChangeSaleCustomer(aCashier(), changeTo(5, 5), 1)
	if err != nil || got == nil || got.ID != 5 {
		t.Fatalf("a repeat is answered as saved: %+v, %v", got, err)
	}
	if len(repo.updated) != 0 || len(repo.audits) != 0 {
		t.Fatal("a repeat must not write or audit anything")
	}
}

func TestRetryOfAChangeThatAlreadyCommittedIsAnsweredAsSaved(t *testing.T) {
	// The first request moved the sale from customer 5 to 6 and the reply was lost: the page
	// still says 5 and asks again. That is the same change, not someone else's.
	repo := &changeRepo{
		sale: aSaleOf(6), customerSaleRepo: customerSaleRepo{customers: twoCustomers()},
	}
	got, err := NewSaleService(repo).ChangeSaleCustomer(aCashier(), changeTo(6, 5), 1)
	if err != nil || got == nil || got.ID != 6 {
		t.Fatalf("a retry is answered as saved: %+v, %v", got, err)
	}
	if len(repo.updated) != 0 || len(repo.audits) != 0 {
		t.Fatal("a retry must not write or audit anything")
	}
	// Even if the customer was deactivated since, the retry is still answered as saved.
	repo.customers[6] = types.Customer{ID: 6, Name: "Bola Ade", Active: false}
	if _, err = NewSaleService(repo).ChangeSaleCustomer(aCashier(), changeTo(6, 5), 1); err != nil {
		t.Fatalf("a retry of a saved change is not refused for a later deactivation: %v", err)
	}
	// A retry of a removal that already committed.
	walkIn := &changeRepo{sale: aSaleOf(0)}
	got, err = NewSaleService(walkIn).ChangeSaleCustomer(aCashier(), changeTo(0, 5), 1)
	if err != nil || got != nil || len(walkIn.updated) != 0 {
		t.Fatalf("a retried removal is answered as saved: %+v, %v", got, err)
	}
}

func TestRemovingFromAWalkInSaleWritesNothing(t *testing.T) {
	repo := &changeRepo{sale: aSaleOf(0)}
	got, err := NewSaleService(repo).ChangeSaleCustomer(aCashier(), changeTo(0, 0), 1)
	if err != nil || got != nil || len(repo.updated) != 0 || len(repo.audits) != 0 {
		t.Fatalf("expected a quiet no-op: %+v, %v", got, err)
	}
}

func TestAttachingADeactivatedOrUnknownCustomerIsRefused(t *testing.T) {
	repo := &changeRepo{
		sale: aSaleOf(5), customerSaleRepo: customerSaleRepo{customers: twoCustomers()},
	}
	_, err := NewSaleService(repo).ChangeSaleCustomer(aCashier(), changeTo(7, 5), 1)
	requireHTTPStatus(t, err, http.StatusConflict)
	requireReason(t, err, "customer_inactive")

	_, err = NewSaleService(repo).ChangeSaleCustomer(aCashier(), changeTo(99, 5), 1)
	requireHTTPStatus(t, err, http.StatusBadRequest)
	if len(repo.updated) != 0 || len(repo.audits) != 0 || repo.commits != 0 {
		t.Fatal("a refused change must leave nothing behind")
	}
}

func TestOnlyAManagerChangesTheCustomerOfAnOlderSale(t *testing.T) {
	older := aSaleOf(5)
	older.IsRecent = false
	repo := &changeRepo{sale: older, customerSaleRepo: customerSaleRepo{customers: twoCustomers()}}

	// Even a request that is also stale gets the age rule first: a cashier learns nothing more
	// about an older sale than History already shows.
	_, err := NewSaleService(repo).ChangeSaleCustomer(aCashier(), changeTo(6, 99), 1)
	requireHTTPStatus(t, err, http.StatusForbidden)
	requireReason(t, err, "sale_too_old")
	if len(repo.updated) != 0 || repo.commits != 0 {
		t.Fatal("a refused change must leave nothing behind")
	}

	if _, err = NewSaleService(repo).ChangeSaleCustomer(aManager(), changeTo(6, 5), 1); err != nil {
		t.Fatal(err)
	}
	if len(repo.updated) != 1 || *repo.updated[0] != 6 {
		t.Fatalf("the manager's change should be saved, got %v", repo.updated)
	}
}

func TestRetryOfAChangeOnAnOldSaleIsAnsweredAsSavedForACashier(t *testing.T) {
	// The first request was made inside the window and committed; the reply was lost and the retry
	// comes after the sale has aged out. It is the same change, so it is answered as saved and
	// writes nothing, whoever asks.
	older := aSaleOf(6)
	older.IsRecent = false
	repo := &changeRepo{sale: older, customerSaleRepo: customerSaleRepo{customers: twoCustomers()}}
	got, err := NewSaleService(repo).ChangeSaleCustomer(aCashier(), changeTo(6, 5), 1)
	if err != nil || got == nil || got.ID != 6 {
		t.Fatalf("a retry on an old sale is answered as saved: %+v, %v", got, err)
	}
	if len(repo.updated) != 0 || len(repo.audits) != 0 {
		t.Fatal("a retry must not write or audit anything")
	}

	// A retried removal on an old sale that is already a walk-in.
	oldWalkIn := aSaleOf(0)
	oldWalkIn.IsRecent = false
	walkIn := &changeRepo{sale: oldWalkIn}
	got, err = NewSaleService(walkIn).ChangeSaleCustomer(aCashier(), changeTo(0, 5), 1)
	if err != nil || got != nil || len(walkIn.updated) != 0 || len(walkIn.audits) != 0 {
		t.Fatalf("a retried removal on an old sale is answered as saved: %+v, %v", got, err)
	}
}

func TestACashierStillCannotMakeARealChangeOnAnOldSale(t *testing.T) {
	older := aSaleOf(5)
	older.IsRecent = false
	repo := &changeRepo{sale: older, customerSaleRepo: customerSaleRepo{customers: twoCustomers()}}
	// Not a repeat (the sale is on 5, the request asks for 6), and not stale either.
	_, err := NewSaleService(repo).ChangeSaleCustomer(aCashier(), changeTo(6, 5), 1)
	requireHTTPStatus(t, err, http.StatusForbidden)
	requireReason(t, err, "sale_too_old")
	if len(repo.updated) != 0 || len(repo.audits) != 0 || repo.commits != 0 {
		t.Fatal("a refused change must leave nothing behind")
	}
}

func TestChangingTheCustomerOfAMissingSaleIsNotFound(t *testing.T) {
	repo := &changeRepo{saleError: sql.ErrNoRows}
	_, err := NewSaleService(repo).ChangeSaleCustomer(aCashier(), changeTo(6, 5), 1)
	requireHTTPStatus(t, err, http.StatusNotFound)
}
