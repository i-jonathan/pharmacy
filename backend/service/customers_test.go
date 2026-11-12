package service

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"pharmacy/httperror"
	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"pharmacy/repository"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
)

// fakeCustomerRepo holds customers in memory and records what the service writes.
type fakeCustomerRepo struct {
	repository.PharmacyRepository

	customers map[int]types.Customer
	created   []types.NewCustomer
	updated   []types.Customer
	active    map[int]bool // customers whose active flag was set, and to what
	audit     []types.AuditEntry
	commits   int
	// phoneRace makes an insert or update fail as if another request took the number first.
	phoneRace bool
	listed    types.CustomerFilter
	limit     int
	offset    int
}

func (r *fakeCustomerRepo) BeginTx(context.Context) (*sqlx.Tx, error) { return nil, nil }
func (r *fakeCustomerRepo) RollbackTx(*sqlx.Tx)                       {}
func (r *fakeCustomerRepo) CommitTx(*sqlx.Tx) error {
	r.commits++
	return nil
}

func (r *fakeCustomerRepo) InsertAuditLogTx(
	_ context.Context,
	_ *sqlx.Tx,
	e types.AuditEntry,
) error {
	r.audit = append(r.audit, e)
	return nil
}

func (r *fakeCustomerRepo) CustomerByPhoneTx(
	_ context.Context, _ *sqlx.Tx, phone string,
) (*types.Customer, error) {
	for _, c := range r.customers {
		if c.Phone == phone {
			c := c
			return &c, nil
		}
	}
	return nil, nil
}

func (r *fakeCustomerRepo) CreateCustomerTx(
	_ context.Context, _ *sqlx.Tx, c types.NewCustomer,
) (int, error) {
	if r.phoneRace {
		return 0, repository.ErrCustomerPhoneTaken
	}
	r.created = append(r.created, c)
	return 50, nil
}

func (r *fakeCustomerRepo) GetCustomerForUpdateTx(
	_ context.Context, _ *sqlx.Tx, id int,
) (types.Customer, error) {
	c, ok := r.customers[id]
	if !ok {
		return types.Customer{}, sql.ErrNoRows
	}
	return c, nil
}

func (r *fakeCustomerRepo) UpdateCustomerTx(_ context.Context, _ *sqlx.Tx, c types.Customer) error {
	if r.phoneRace {
		return repository.ErrCustomerPhoneTaken
	}
	r.updated = append(r.updated, c)
	return nil
}

func (r *fakeCustomerRepo) SetCustomerActiveTx(
	_ context.Context, _ *sqlx.Tx, id int, active bool,
) error {
	if r.active == nil {
		r.active = map[int]bool{}
	}
	r.active[id] = active
	return nil
}

func (r *fakeCustomerRepo) GetCustomer(_ context.Context, id int) (types.Customer, error) {
	c, ok := r.customers[id]
	if !ok {
		return types.Customer{}, sql.ErrNoRows
	}
	return c, nil
}

func (r *fakeCustomerRepo) ListCustomers(
	_ context.Context, filter types.CustomerFilter, limit, offset int,
) ([]types.Customer, int, error) {
	r.listed, r.limit, r.offset = filter, limit, offset
	return []types.Customer{}, 0, nil
}

func managerOfCustomers() context.Context {
	return contextWith(constant.ViewCustomersPermissionKey, constant.ManageCustomersPermissionKey)
}

func viewerOfCustomers() context.Context {
	return contextWith(constant.ViewCustomersPermissionKey)
}

// storedAda is a customer already on file.
func storedAda() types.Customer {
	return types.Customer{
		ID: 7, Name: "Ada Obi", Phone: "+2348031234567", Email: "ada@example.com",
		Notes: "prefers liquid syrups", Active: true, Version: 3,
	}
}

func newCustomerService(repo *fakeCustomerRepo) *customerService {
	return NewCustomerService(repo)
}

func requireAudit(t *testing.T, repo *fakeCustomerRepo, action string) types.AuditEntry {
	t.Helper()
	if len(repo.audit) != 1 {
		t.Fatalf("expected one audit entry, got %d", len(repo.audit))
	}
	e := repo.audit[0]
	if e.EntityType != "customer" || e.Action != action {
		t.Fatalf("expected a customer %s audit entry, got %s %s", action, e.EntityType, e.Action)
	}
	return e
}

func TestCreateCustomerSavesTheTidiedCustomerAndAudits(t *testing.T) {
	repo := &fakeCustomerRepo{}
	id, err := newCustomerService(repo).CreateCustomer(managerOfCustomers(), types.CustomerRequest{
		Name: "  Ada Obi ", Phone: "0803 123 4567", Email: " ada@example.com ", Notes: " note ",
		ContactConsent: true,
	}, 4)
	if err != nil || id != 50 {
		t.Fatalf("got id %d, err %v", id, err)
	}
	want := types.NewCustomer{
		Name: "Ada Obi", Phone: "+2348031234567", Email: "ada@example.com", Notes: "note",
		ContactConsent: true, CreatedByID: 4,
	}
	if len(repo.created) != 1 || repo.created[0] != want {
		t.Fatalf("saved %+v, want %+v", repo.created, want)
	}
	e := requireAudit(t, repo, "created")
	if e.ActorID != 4 || e.EntityID != 50 || repo.commits != 1 {
		t.Fatalf("audit %+v, commits %d", e, repo.commits)
	}
	if !strings.Contains(string(e.Details), `"contact_consent":true`) {
		t.Fatalf("the audit entry should record the consent given, got %s", e.Details)
	}
}

func TestCreateCustomerWithoutAPhoneIsAllowed(t *testing.T) {
	repo := &fakeCustomerRepo{}
	_, err := newCustomerService(repo).CreateCustomer(
		managerOfCustomers(), types.CustomerRequest{Name: "Walk-in"}, 4,
	)
	if err != nil || len(repo.created) != 1 || repo.created[0].Phone != "" {
		t.Fatalf("err %v, saved %+v", err, repo.created)
	}
}

func TestCreateCustomerNeedsTheManagePermission(t *testing.T) {
	for name, ctx := range map[string]context.Context{
		"viewer": viewerOfCustomers(), "nobody": contextWith(),
	} {
		repo := &fakeCustomerRepo{}
		_, err := newCustomerService(
			repo,
		).CreateCustomer(ctx, types.CustomerRequest{Name: "Ada"}, 4)
		requireHTTPStatus(t, err, http.StatusForbidden)
		if len(repo.created) != 0 || len(repo.audit) != 0 {
			t.Errorf("%s: nothing should be written", name)
		}
	}
}

func TestCreateCustomerRefusesInputThatCanNeverBeValid(t *testing.T) {
	long := strings.Repeat("x", 121)
	for name, req := range map[string]types.CustomerRequest{
		"blank name":    {Name: "   "},
		"long name":     {Name: long},
		"bad phone":     {Name: "Ada", Phone: "0803 123"},
		"email no at":   {Name: "Ada", Email: "ada.example.com"},
		"email display": {Name: "Ada", Email: "Ada <ada@example.com>"},
		"long notes":    {Name: "Ada", Notes: strings.Repeat("n", 1001)},
	} {
		repo := &fakeCustomerRepo{}
		_, err := newCustomerService(repo).CreateCustomer(managerOfCustomers(), req, 4)
		if err == nil {
			t.Errorf("%s: expected an error", name)
			continue
		}
		requireHTTPStatus(t, err, http.StatusBadRequest)
		if len(repo.created) != 0 {
			t.Errorf("%s: nothing should be written", name)
		}
	}
}

func TestCreateCustomerRefusesAPhoneAnotherCustomerHas(t *testing.T) {
	ada := storedAda()
	ada.Active = false
	repo := &fakeCustomerRepo{customers: map[int]types.Customer{7: ada}}
	_, err := newCustomerService(repo).CreateCustomer(managerOfCustomers(), types.CustomerRequest{
		Name: "Ada's sister", Phone: "08031234567",
	}, 4)
	requireHTTPStatus(t, err, http.StatusConflict)
	var httperr *httperror.HTTPError
	errors.As(err, &httperr)
	if httperr.Reason != "phone_taken" ||
		!strings.Contains(httperr.Message, "Ada Obi") ||
		!strings.Contains(httperr.Message, "deactivated") {
		t.Fatalf("the refusal should name the customer and say they are deactivated: %+v", httperr)
	}
	if len(repo.created) != 0 {
		t.Fatal("nothing should be saved")
	}
}

func TestCreateCustomerLosingARaceForAPhoneIsAConflict(t *testing.T) {
	repo := &fakeCustomerRepo{phoneRace: true}
	_, err := newCustomerService(repo).CreateCustomer(
		managerOfCustomers(), types.CustomerRequest{Name: "Ada", Phone: "08031234567"}, 4,
	)
	requireHTTPStatus(t, err, http.StatusConflict)
	if len(repo.audit) != 0 || repo.commits != 0 {
		t.Fatal("a refused create must not audit or commit")
	}
}

func editAda(change func(*types.CustomerRequest)) types.CustomerRequest {
	req := types.CustomerRequest{
		ID: 7, Version: 3, Name: "Ada Obi", Phone: "+234 803 123 4567", Email: "ada@example.com",
		Notes: "prefers liquid syrups",
	}
	change(&req)
	return req
}

func TestUpdateCustomerSavesAndAuditsOnlyWhatChanged(t *testing.T) {
	repo := &fakeCustomerRepo{customers: map[int]types.Customer{7: storedAda()}}
	err := newCustomerService(repo).UpdateCustomer(managerOfCustomers(),
		editAda(func(r *types.CustomerRequest) { r.Name = "Ada Okafor"; r.Notes = "" }), 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(repo.updated) != 1 {
		t.Fatalf("expected one update, got %d", len(repo.updated))
	}
	got := repo.updated[0]
	if got.Name != "Ada Okafor" || got.Notes != "" || got.Phone != "+2348031234567" ||
		got.ConsentUpdatedAt != nil {
		t.Fatalf("saved %+v", got)
	}
	e := requireAudit(t, repo, "updated")
	d := string(e.Details)
	if !strings.Contains(d, `"name"`) || !strings.Contains(d, `"notes"`) ||
		strings.Contains(d, `"phone"`) || strings.Contains(d, "contact_consent") {
		t.Fatalf("audit details should list name and notes only, got %s", d)
	}
}

func TestUpdateCustomerRecordsWhenConsentChanges(t *testing.T) {
	repo := &fakeCustomerRepo{customers: map[int]types.Customer{7: storedAda()}}
	err := newCustomerService(repo).UpdateCustomer(managerOfCustomers(),
		editAda(func(r *types.CustomerRequest) { r.ContactConsent = true }), 4)
	if err != nil {
		t.Fatal(err)
	}
	if got := repo.updated[0]; !got.ContactConsent {
		t.Fatalf("consent should be saved: %+v", got)
	}
	e := requireAudit(t, repo, "updated")
	if !strings.Contains(string(e.Details), `"contact_consent":{"from":false,"to":true}`) {
		t.Fatalf("audit should record consent going from false to true, got %s", e.Details)
	}
}

func TestUpdateCustomerWithNothingChangedWritesNothing(t *testing.T) {
	repo := &fakeCustomerRepo{customers: map[int]types.Customer{7: storedAda()}}
	// The phone written another way is still the same phone.
	err := newCustomerService(repo).UpdateCustomer(managerOfCustomers(),
		editAda(func(r *types.CustomerRequest) { r.Phone = "0803 123 4567" }), 4)
	if err != nil || len(repo.updated) != 0 || len(repo.audit) != 0 {
		t.Fatalf("err %v, updates %d, audit %d", err, len(repo.updated), len(repo.audit))
	}
}

func TestUpdateCustomerRefusesAStaleVersion(t *testing.T) {
	repo := &fakeCustomerRepo{customers: map[int]types.Customer{7: storedAda()}}
	err := newCustomerService(repo).UpdateCustomer(managerOfCustomers(),
		editAda(func(r *types.CustomerRequest) { r.Version = 2; r.Name = "Someone else" }), 4)
	requireHTTPStatus(t, err, http.StatusConflict)
	if len(repo.updated) != 0 || len(repo.audit) != 0 {
		t.Fatal("nothing should be written")
	}
}

func TestUpdateCustomerThatDoesNotExistIsNotFound(t *testing.T) {
	repo := &fakeCustomerRepo{}
	err := newCustomerService(
		repo,
	).UpdateCustomer(managerOfCustomers(), editAda(func(*types.CustomerRequest) {}), 4)
	requireHTTPStatus(t, err, http.StatusNotFound)
}

func TestUpdateCustomerCannotTakeAnotherCustomersPhone(t *testing.T) {
	bola := types.Customer{ID: 8, Name: "Bola", Phone: "+2348099999999", Active: true, Version: 1}
	repo := &fakeCustomerRepo{customers: map[int]types.Customer{7: storedAda(), 8: bola}}
	err := newCustomerService(repo).UpdateCustomer(managerOfCustomers(),
		editAda(func(r *types.CustomerRequest) { r.Phone = "08099999999" }), 4)
	requireHTTPStatus(t, err, http.StatusConflict)
	if len(repo.updated) != 0 {
		t.Fatal("nothing should be written")
	}
}

func TestUpdateCustomerNeedsTheManagePermission(t *testing.T) {
	repo := &fakeCustomerRepo{customers: map[int]types.Customer{7: storedAda()}}
	err := newCustomerService(repo).UpdateCustomer(viewerOfCustomers(),
		editAda(func(r *types.CustomerRequest) { r.Name = "Changed" }), 4)
	requireHTTPStatus(t, err, http.StatusForbidden)
	if len(repo.updated) != 0 {
		t.Fatal("nothing should be written")
	}
}

func TestDeactivateAndReactivateCustomer(t *testing.T) {
	for _, tc := range []struct {
		name   string
		active bool // the state asked for
		action string
	}{
		{"deactivate", false, "deactivated"},
		{"reactivate", true, "reactivated"},
	} {
		stored := storedAda()
		stored.Active = !tc.active
		repo := &fakeCustomerRepo{customers: map[int]types.Customer{7: stored}}
		err := newCustomerService(repo).SetCustomerActive(managerOfCustomers(), 7, 3, tc.active, 4)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if got, ok := repo.active[7]; !ok || got != tc.active {
			t.Errorf("%s: active flag %v, set %v", tc.name, got, ok)
		}
		requireAudit(t, repo, tc.action)
	}
}

func TestSetCustomerActiveRefusesAStaleVersionAndIgnoresNoChange(t *testing.T) {
	repo := &fakeCustomerRepo{customers: map[int]types.Customer{7: storedAda()}}
	err := newCustomerService(repo).SetCustomerActive(managerOfCustomers(), 7, 1, false, 4)
	requireHTTPStatus(t, err, http.StatusConflict)

	// Already active and asked to be active: nothing to do, nothing to audit.
	err = newCustomerService(repo).SetCustomerActive(managerOfCustomers(), 7, 3, true, 4)
	if err != nil || len(repo.active) != 0 || len(repo.audit) != 0 {
		t.Fatalf("err %v, active %v, audit %d", err, repo.active, len(repo.audit))
	}
}

func TestSetCustomerActiveNeedsTheManagePermission(t *testing.T) {
	repo := &fakeCustomerRepo{customers: map[int]types.Customer{7: storedAda()}}
	err := newCustomerService(repo).SetCustomerActive(viewerOfCustomers(), 7, 3, false, 4)
	requireHTTPStatus(t, err, http.StatusForbidden)
}

func TestReadingCustomersNeedsTheViewPermission(t *testing.T) {
	repo := &fakeCustomerRepo{customers: map[int]types.Customer{7: storedAda()}}
	svc := newCustomerService(repo)
	if _, err := svc.GetCustomer(contextWith(), 7); err == nil {
		t.Error("GetCustomer should be refused")
	} else {
		requireHTTPStatus(t, err, http.StatusForbidden)
	}
	if _, err := svc.ListCustomers(contextWith(), "", false, 1); err == nil {
		t.Error("ListCustomers should be refused")
	} else {
		requireHTTPStatus(t, err, http.StatusForbidden)
	}
	if c, err := svc.GetCustomer(viewerOfCustomers(), 7); err != nil || c.Name != "Ada Obi" {
		t.Errorf("a viewer can read: %v %v", c, err)
	}
	_, err := svc.GetCustomer(viewerOfCustomers(), 99)
	requireHTTPStatus(t, err, http.StatusNotFound)
}

func TestListCustomersSearchesNamesByTextAndPhonesByDigits(t *testing.T) {
	for _, tc := range []struct {
		search string
		want   types.CustomerFilter
	}{
		{"", types.CustomerFilter{}},
		{"  ada  ", types.CustomerFilter{Name: "ada"}},
		{"Ada 2", types.CustomerFilter{Name: "Ada 2"}},
		// A number written the way people write it starts at the start of the stored number.
		{"0803 123", types.CustomerFilter{PhonePrefix: "+234", PhoneDigits: "803123"}},
		{"+234 803", types.CustomerFilter{PhonePrefix: "+234", PhoneDigits: "803"}},
		{"+234 0803", types.CustomerFilter{PhonePrefix: "+234", PhoneDigits: "803"}},
		{"23480", types.CustomerFilter{PhonePrefix: "+234", PhoneDigits: "80"}},
		{"0", types.CustomerFilter{PhonePrefix: "+234"}},
		{"+44 20", types.CustomerFilter{PhonePrefix: "+", PhoneDigits: "4420"}},
		// Digits that do not start a number match anywhere inside it.
		{"803", types.CustomerFilter{PhoneDigits: "803"}},
		{"7038", types.CustomerFilter{PhoneDigits: "7038"}},
	} {
		repo := &fakeCustomerRepo{}
		_, err := newCustomerService(repo).ListCustomers(viewerOfCustomers(), tc.search, false, 1)
		if err != nil {
			t.Fatal(err)
		}
		if repo.listed != tc.want {
			t.Errorf("search %q filtered by %+v, want %+v", tc.search, repo.listed, tc.want)
		}
	}
}

func TestListCustomersPagesAndOnlyShowsDeactivatedWhenAsked(t *testing.T) {
	repo := &fakeCustomerRepo{}
	svc := newCustomerService(repo)

	page, err := svc.ListCustomers(viewerOfCustomers(), "", true, 3)
	if err != nil {
		t.Fatal(err)
	}
	if !repo.listed.IncludeInactive || repo.limit != customersPerPage ||
		repo.offset != 2*customersPerPage || page.Page != 3 || page.PerPage != customersPerPage {
		t.Fatalf(
			"filter %+v limit %d offset %d page %+v",
			repo.listed,
			repo.limit,
			repo.offset,
			page,
		)
	}

	// A page number that is nonsense falls back to the first page instead of an odd offset.
	for _, bad := range []int{0, -4, maxCustomerPage + 1} {
		if page, err = svc.ListCustomers(viewerOfCustomers(), "", false, bad); err != nil ||
			page.Page != 1 || repo.offset != 0 {
			t.Errorf("page %d gave page %d offset %d err %v", bad, page.Page, repo.offset, err)
		}
	}
}
