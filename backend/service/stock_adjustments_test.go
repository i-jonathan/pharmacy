package service

import (
	"context"
	"errors"
	"net/http"
	"pharmacy/httperror"
	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"pharmacy/model"
	"pharmacy/repository"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
)

// fakeStockRepo holds products' stock and records what an adjustment writes.
type fakeStockRepo struct {
	repository.PharmacyRepository

	products  map[int]types.ProductStock
	flags     map[int]types.StockFlag
	existing  int // id returned for an idempotency key already used
	flagErr   error
	items     []types.NewStockAdjustmentItem
	movements []model.StockMovement
	expiries  map[int]time.Time
	audit     []types.AuditEntry
	resolved  []int
	dismissed int
	adjusted  []types.NewStockAdjustment
	commits   int
	offset    int
	counting  []int // products counted in an in-progress stock taking
}

func (r *fakeStockRepo) ProductsInActiveStockTakingTx(
	_ context.Context, _ *sqlx.Tx, ids []int,
) ([]int, error) {
	var found []int
	for _, id := range ids {
		for _, c := range r.counting {
			if c == id {
				found = append(found, id)
			}
		}
	}
	return found, nil
}

func (r *fakeStockRepo) BeginTx(context.Context) (*sqlx.Tx, error) { return nil, nil }
func (r *fakeStockRepo) RollbackTx(*sqlx.Tx)                       {}
func (r *fakeStockRepo) CommitTx(*sqlx.Tx) error {
	r.commits++
	return nil
}

func (r *fakeStockRepo) LockProductStockTx(
	_ context.Context, _ *sqlx.Tx, ids []int,
) ([]types.ProductStock, error) {
	var out []types.ProductStock
	for _, id := range ids {
		if p, ok := r.products[id]; ok {
			out = append(out, p)
		}
	}
	return out, nil
}

func (r *fakeStockRepo) StockAdjustmentIDByKeyTx(
	_ context.Context, _ *sqlx.Tx, _ int, _ string,
) (int, error) {
	return r.existing, nil
}

func (r *fakeStockRepo) CreateStockAdjustmentTx(
	_ context.Context, _ *sqlx.Tx, adj types.NewStockAdjustment,
) (int, bool, error) {
	r.adjusted = append(r.adjusted, adj)
	return 70, true, nil
}

func (r *fakeStockRepo) InsertStockAdjustmentItemTx(
	_ context.Context, _ *sqlx.Tx, _ int, item types.NewStockAdjustmentItem,
) error {
	r.items = append(r.items, item)
	return nil
}

func (r *fakeStockRepo) BulkCreateStockMovementTx(
	_ context.Context, _ *sqlx.Tx, m []model.StockMovement,
) error {
	r.movements = append(r.movements, m...)
	return nil
}

func (r *fakeStockRepo) SetProductCurrentExpiryTx(
	_ context.Context, _ *sqlx.Tx, productID int, expiry time.Time,
) error {
	if r.expiries == nil {
		r.expiries = map[int]time.Time{}
	}
	r.expiries[productID] = expiry
	return nil
}

func (r *fakeStockRepo) InsertAuditLogTx(_ context.Context, _ *sqlx.Tx, e types.AuditEntry) error {
	r.audit = append(r.audit, e)
	return nil
}

func (r *fakeStockRepo) CreateStockFlagTx(
	_ context.Context, _ *sqlx.Tx, _ types.StockFlagRequest, _ int,
) (int, error) {
	if r.flagErr != nil {
		return 0, r.flagErr
	}
	return 90, nil
}

func (r *fakeStockRepo) GetStockFlagsForUpdateTx(
	_ context.Context, _ *sqlx.Tx, ids []int,
) ([]types.StockFlag, error) {
	var out []types.StockFlag
	for _, id := range ids {
		if f, ok := r.flags[id]; ok {
			out = append(out, f)
		}
	}
	return out, nil
}

func (r *fakeStockRepo) ResolveStockFlagsTx(
	_ context.Context, _ *sqlx.Tx, ids []int, _, _ int,
) error {
	r.resolved = append(r.resolved, ids...)
	return nil
}

func (r *fakeStockRepo) DismissStockFlagTx(
	_ context.Context, _ *sqlx.Tx, id, _ int, _ string,
) error {
	r.dismissed = id
	return nil
}

func (r *fakeStockRepo) ListStockAdjustments(
	_ context.Context, _ int, offset int,
) ([]types.StockAdjustment, int, error) {
	r.offset = offset
	return []types.StockAdjustment{}, 0, nil
}

func (r *fakeStockRepo) ListOpenStockFlags(context.Context) ([]types.StockFlag, error) {
	return []types.StockFlag{}, nil
}

func (r *fakeStockRepo) CountOpenStockFlags(context.Context) (int, error) { return 3, nil }

const amoxil, paracetamol = 1, 2

func newStockRepo() *fakeStockRepo {
	return &fakeStockRepo{
		products: map[int]types.ProductStock{
			amoxil:      {ID: amoxil, Name: "Amoxil", CostPriceKobo: 45000, Stock: 12},
			paracetamol: {ID: paracetamol, Name: "Paracetamol", CostPriceKobo: 2000, Stock: 2},
		},
		flags: map[int]types.StockFlag{
			5: {ID: 5, ProductID: amoxil, Status: "open"},
			6: {ID: 6, ProductID: paracetamol, Status: "open"},
			7: {ID: 7, ProductID: amoxil, Status: "resolved"},
		},
	}
}

func managerContext() context.Context {
	return contextWith(constant.AdjustStockPermissionKey, constant.FlagStockPermissionKey)
}

func cashierContext() context.Context { return contextWith(constant.FlagStockPermissionKey) }

func adjust(items ...types.StockAdjustmentItemRequest) types.StockAdjustmentRequest {
	return types.StockAdjustmentRequest{Reason: "damaged", Items: items}
}

func item(productID int, mode string, quantity int) types.StockAdjustmentItemRequest {
	return types.StockAdjustmentItemRequest{ProductID: productID, Mode: mode, Quantity: quantity}
}

func strPtr(s string) *string { return &s }

func TestStockAdjustmentNeedsTheAdjustPermission(t *testing.T) {
	repo := newStockRepo()
	svc := NewInventoryService(repo)

	_, err := svc.CreateStockAdjustment(cashierContext(), adjust(item(amoxil, "set", 10)), 4)

	requireHTTPStatus(t, err, http.StatusForbidden)
	if repo.commits != 0 {
		t.Fatal("nothing should be saved")
	}
}

func TestStockAdjustmentRejectsInvalidRequests(t *testing.T) {
	many := make([]types.StockAdjustmentItemRequest, 201)
	for i := range many {
		many[i] = item(i+1, "add", 1)
	}
	cases := map[string]types.StockAdjustmentRequest{
		"unknown reason": {Reason: "borrowed", Items: adjust(item(amoxil, "add", 1)).Items},
		"other, no note": {Reason: "other", Items: adjust(item(amoxil, "add", 1)).Items},
		"no items":       {Reason: "damaged"},
		"too many items": {Reason: "damaged", Items: many},
		"same product":   adjust(item(amoxil, "add", 1), item(amoxil, "remove", 1)),
		"unknown mode":   adjust(item(amoxil, "double", 1)),
		"negative count": adjust(item(amoxil, "set", -1)),
		"add nothing":    adjust(item(amoxil, "add", 0)),
		"remove nothing": adjust(item(amoxil, "remove", 0)),
		"huge add":       adjust(item(amoxil, "add", 1_000_001)),
		"huge count":     adjust(item(amoxil, "set", 1_000_001)),
		"bad expiry": adjust(
			types.StockAdjustmentItemRequest{
				ProductID: amoxil,
				Mode:      "add",
				Quantity:  1,
				Expiry:    strPtr("31/12/2027"),
			},
		),
		"ancient expiry": adjust(types.StockAdjustmentItemRequest{
			ProductID: amoxil, Mode: "add", Quantity: 1, Expiry: strPtr("0001-01-01"),
		}),
		"far-off expiry": adjust(types.StockAdjustmentItemRequest{
			ProductID: amoxil, Mode: "add", Quantity: 1, Expiry: strPtr("9999-12-31"),
		}),
		"unknown product":   adjust(item(99, "add", 1)),
		"below zero":        adjust(item(paracetamol, "remove", 3)),
		"nothing to change": adjust(item(amoxil, "set", 12)),
	}
	for name, req := range cases {
		t.Run(name, func(t *testing.T) {
			repo := newStockRepo()

			_, err := NewInventoryService(repo).CreateStockAdjustment(managerContext(), req, 4)

			requireHTTPStatus(t, err, http.StatusBadRequest)
			if repo.commits != 0 || len(repo.movements) != 0 {
				t.Fatal("nothing should be saved")
			}
		})
	}
}

func TestStockAdjustmentBelowZeroNamesTheProduct(t *testing.T) {
	_, err := NewInventoryService(newStockRepo()).CreateStockAdjustment(
		managerContext(), adjust(item(paracetamol, "remove", 3)), 4,
	)
	if err == nil || !strings.Contains(err.Error(), "Paracetamol: only 2 in stock") {
		t.Fatalf("expected the product and its stock in the error, got %v", err)
	}
}

func TestStockAdjustmentRecordsEachModeAsMovements(t *testing.T) {
	repo := newStockRepo()
	req := adjust(item(amoxil, "set", 10), item(paracetamol, "add", 5))
	req.Note = "  crushed in delivery "

	id, err := NewInventoryService(repo).CreateStockAdjustment(managerContext(), req, 4)
	if err != nil {
		t.Fatal(err)
	}

	if id != 70 || repo.commits != 1 {
		t.Fatalf("expected adjustment 70 committed once, got %d, %d commits", id, repo.commits)
	}
	if got := repo.adjusted[0]; got.Reason != "damaged" || got.Note != "crushed in delivery" ||
		got.CreatedByID != 4 {
		t.Fatalf("wrong header: %+v", got)
	}
	wantItems := []types.NewStockAdjustmentItem{
		{ProductID: amoxil, QuantityBefore: 12, QuantityChange: -2, UnitCostKobo: 45000},
		{ProductID: paracetamol, QuantityBefore: 2, QuantityChange: 5, UnitCostKobo: 2000},
	}
	for i, want := range wantItems {
		got := repo.items[i]
		if got.ProductID != want.ProductID || got.QuantityBefore != want.QuantityBefore ||
			got.QuantityChange != want.QuantityChange || got.UnitCostKobo != want.UnitCostKobo {
			t.Fatalf("item %d: want %+v, got %+v", i, want, got)
		}
	}
	wantMoves := []model.StockMovement{
		{ProductID: amoxil, Quantity: 2, ReferenceID: 70,
			MovementType: model.MovementTypeOutManualAdjustment},
		{ProductID: paracetamol, Quantity: 5, ReferenceID: 70,
			MovementType: model.MovementTypeInManualAdjustment},
	}
	for i, want := range wantMoves {
		got := repo.movements[i]
		if got.ProductID != want.ProductID || got.Quantity != want.Quantity ||
			got.ReferenceID != want.ReferenceID || got.MovementType != want.MovementType {
			t.Fatalf("movement %d: want %+v, got %+v", i, want, got)
		}
	}
	if len(repo.audit) != 1 || repo.audit[0].EntityType != constant.AuditEntityStockAdjustment ||
		repo.audit[0].EntityID != 70 || repo.audit[0].ActorID != 4 ||
		!strings.Contains(string(repo.audit[0].Details), "Amoxil") {
		t.Fatalf("expected one audit entry for the adjustment, got %+v", repo.audit)
	}
}

func TestStockAdjustmentCanCorrectOnlyTheExpiry(t *testing.T) {
	repo := newStockRepo()
	req := adjust(types.StockAdjustmentItemRequest{
		ProductID: amoxil, Mode: "set", Quantity: 12, Expiry: strPtr("2027-03-31"),
	})
	req.Reason = "count_correction"

	if _, err := NewInventoryService(
		repo,
	).CreateStockAdjustment(managerContext(), req, 4); err != nil {
		t.Fatal(err)
	}

	if len(repo.movements) != 0 {
		t.Fatalf("an expiry-only change moves no stock, got %+v", repo.movements)
	}
	want := time.Date(2027, 3, 31, 0, 0, 0, 0, time.UTC)
	if got := repo.expiries[amoxil]; !got.Equal(want) {
		t.Fatalf("expected expiry %v, got %v", want, got)
	}
	if repo.items[0].ExpiryAfter == nil || !repo.items[0].ExpiryAfter.Equal(want) {
		t.Fatalf("the item should record the new expiry, got %+v", repo.items[0])
	}
}

func TestStockAdjustmentRetryReturnsTheFirstAdjustment(t *testing.T) {
	repo := newStockRepo()
	repo.existing = 55
	// Stock already reflects the first save, so re-checking the items would refuse them.
	req := adjust(item(paracetamol, "remove", 3))
	req.IdempotencyKey = "key-1"

	id, err := NewInventoryService(repo).CreateStockAdjustment(managerContext(), req, 4)

	if err != nil || id != 55 {
		t.Fatalf("expected the existing adjustment 55, got %d, %v", id, err)
	}
	if len(repo.items) != 0 || len(repo.movements) != 0 || len(repo.audit) != 0 {
		t.Fatal("a retry must not write anything again")
	}
}

func TestStockAdjustmentResolvesTheFlagsItAnswers(t *testing.T) {
	repo := newStockRepo()
	req := adjust(item(amoxil, "set", 10))
	req.FlagIDs = []int{5}

	if _, err := NewInventoryService(
		repo,
	).CreateStockAdjustment(managerContext(), req, 4); err != nil {
		t.Fatal(err)
	}
	if len(repo.resolved) != 1 || repo.resolved[0] != 5 {
		t.Fatalf("flag 5 should be resolved, got %v", repo.resolved)
	}
	var flagAudits int
	for _, e := range repo.audit {
		if e.EntityType == constant.AuditEntityStockFlag && e.Action == constant.AuditResolved {
			flagAudits++
		}
	}
	if flagAudits != 1 {
		t.Fatalf("expected one resolved-flag audit entry, got %d", flagAudits)
	}
}

func TestStockAdjustmentRefusesFlagsItDoesNotAnswer(t *testing.T) {
	cases := map[string]struct {
		flags []int
		code  int
	}{
		"flag for another product": {[]int{6}, http.StatusBadRequest},
		"flag already handled":     {[]int{7}, http.StatusConflict},
		"flag that doesn't exist":  {[]int{99}, http.StatusNotFound},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			repo := newStockRepo()
			req := adjust(item(amoxil, "set", 10))
			req.FlagIDs = c.flags

			_, err := NewInventoryService(repo).CreateStockAdjustment(managerContext(), req, 4)

			requireHTTPStatus(t, err, c.code)
			if repo.commits != 0 {
				t.Fatal("nothing should be saved")
			}
		})
	}
}

func TestFlagStockNeedsAFlagOrAdjustPermission(t *testing.T) {
	_, err := NewInventoryService(newStockRepo()).FlagStock(
		bareContext(), types.StockFlagRequest{ProductID: amoxil, Note: "looks low"}, 4,
	)
	requireHTTPStatus(t, err, http.StatusForbidden)
}

func TestFlagStockValidatesAndAudits(t *testing.T) {
	negative, huge := -1, 1_000_001
	bad := map[string]types.StockFlagRequest{
		"no note":        {ProductID: amoxil, Note: "  "},
		"negative count": {ProductID: amoxil, Note: "x", CountedQuantity: &negative},
		"huge count":     {ProductID: amoxil, Note: "x", CountedQuantity: &huge},
	}
	for name, req := range bad {
		_, err := NewInventoryService(newStockRepo()).FlagStock(cashierContext(), req, 4)
		if status := errorStatus(err); status != http.StatusBadRequest {
			t.Errorf("%s: expected 400, got %v", name, err)
		}
	}

	repo := newStockRepo()
	repo.flagErr = repository.ErrInvalidReference
	_, err := NewInventoryService(repo).FlagStock(
		cashierContext(), types.StockFlagRequest{ProductID: 99, Note: "missing"}, 4,
	)
	requireHTTPStatus(t, err, http.StatusBadRequest)

	repo = newStockRepo()
	repo.flagErr = repository.ErrAlreadyFlagged
	_, err = NewInventoryService(repo).FlagStock(
		cashierContext(), types.StockFlagRequest{ProductID: amoxil, Note: "again"}, 4,
	)
	requireHTTPStatus(t, err, http.StatusConflict)

	repo = newStockRepo()
	id, err := NewInventoryService(repo).FlagStock(
		cashierContext(), types.StockFlagRequest{ProductID: amoxil, Note: " two packs crushed "}, 4,
	)
	if err != nil || id != 90 || repo.commits != 1 {
		t.Fatalf("expected flag 90 committed, got %d, %v", id, err)
	}
	if len(repo.audit) != 1 || repo.audit[0].EntityType != constant.AuditEntityStockFlag ||
		repo.audit[0].Action != constant.AuditCreated {
		t.Fatalf("expected a created-flag audit entry, got %+v", repo.audit)
	}
}

func TestDismissStockFlag(t *testing.T) {
	svc := NewInventoryService(newStockRepo())
	requireHTTPStatus(t, svc.DismissStockFlag(cashierContext(), 5, "fine", 4), http.StatusForbidden)
	requireHTTPStatus(t, svc.DismissStockFlag(managerContext(), 5, " ", 4), http.StatusBadRequest)
	requireHTTPStatus(t, svc.DismissStockFlag(managerContext(), 7, "fine", 4), http.StatusConflict)
	requireHTTPStatus(t, svc.DismissStockFlag(managerContext(), 99, "fine", 4), http.StatusNotFound)

	repo := newStockRepo()
	if err := NewInventoryService(
		repo,
	).DismissStockFlag(managerContext(), 5, "counted, all fine", 4); err != nil {
		t.Fatal(err)
	}
	if repo.dismissed != 5 || repo.commits != 1 || len(repo.audit) != 1 ||
		repo.audit[0].Action != constant.AuditDismissed {
		t.Fatalf("flag 5 should be dismissed and audited, got %+v", repo.audit)
	}
}

func TestStockAdjustmentWithTheSameExpirySaysSo(t *testing.T) {
	repo := newStockRepo()
	current := time.Date(2027, 3, 31, 0, 0, 0, 0, time.UTC)
	p := repo.products[amoxil]
	p.CurrentExpiry = &current
	repo.products[amoxil] = p
	req := adjust(types.StockAdjustmentItemRequest{
		ProductID: amoxil, Mode: "set", Quantity: 12, Expiry: strPtr("2027-03-31"),
	})

	_, err := NewInventoryService(repo).CreateStockAdjustment(managerContext(), req, 4)

	if err == nil || !strings.Contains(err.Error(), "the expiry is already 2027-03-31") {
		t.Fatalf("expected the error to say the expiry is unchanged, got %v", err)
	}
}

func TestStockAdjustmentRefusesProductsBeingCounted(t *testing.T) {
	repo := newStockRepo()
	repo.counting = []int{paracetamol}

	_, err := NewInventoryService(repo).CreateStockAdjustment(
		managerContext(), adjust(item(amoxil, "add", 1), item(paracetamol, "remove", 1)), 4,
	)

	requireHTTPStatus(t, err, http.StatusConflict)
	if !strings.Contains(err.Error(), "Paracetamol") || repo.commits != 0 {
		t.Fatalf("expected a refusal naming Paracetamol and nothing saved, got %v", err)
	}
}

func TestStockAdjustmentAuditOnlyRecordsAnExpiryThatChanged(t *testing.T) {
	repo := newStockRepo()
	current := time.Date(2027, 3, 1, 0, 0, 0, 0, time.UTC)
	p := repo.products[amoxil]
	p.CurrentExpiry = &current
	repo.products[amoxil] = p

	_, err := NewInventoryService(repo).CreateStockAdjustment(
		managerContext(), adjust(item(amoxil, "add", 1)), 4,
	)

	if err != nil || len(repo.audit) != 1 {
		t.Fatalf("expected one audit entry, got %v, %v", repo.audit, err)
	}
	if strings.Contains(string(repo.audit[0].Details), "expiry") {
		t.Fatalf("a quantity-only change must not record an expiry change: %s",
			repo.audit[0].Details)
	}
}

func TestStockAdjustmentHistoryCapsThePage(t *testing.T) {
	repo := newStockRepo()

	_, err := NewInventoryService(repo).ListStockAdjustments(managerContext(), 1<<62)

	if err != nil || repo.offset < 0 {
		t.Fatalf("a huge page must give a valid offset, got %d, %v", repo.offset, err)
	}
}

func TestStockAdjustmentListsNeedTheAdjustPermission(t *testing.T) {
	svc := NewInventoryService(newStockRepo())
	_, err := svc.ListStockAdjustments(cashierContext(), 1)
	requireHTTPStatus(t, err, http.StatusForbidden)
	_, err = svc.ListStockFlags(cashierContext())
	requireHTTPStatus(t, err, http.StatusForbidden)
	_, err = svc.CountStockFlags(cashierContext())
	requireHTTPStatus(t, err, http.StatusForbidden)
	if n, err := svc.CountStockFlags(managerContext()); err != nil || n != 3 {
		t.Fatalf("expected 3 open flags, got %d, %v", n, err)
	}
}

func errorStatus(err error) int {
	var httperr *httperror.HTTPError
	if errors.As(err, &httperr) {
		return httperr.Code
	}
	return 0
}
