package service_test

import (
	"errors"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/jmoiron/sqlx"

	"pharmacy/httperror"
	"pharmacy/internal/constant"
	"pharmacy/internal/testdb"
	"pharmacy/internal/types"
	"pharmacy/repository"
	"pharmacy/service"
)

// These run the real service and SQL against a throwaway database; see internal/testdb.

func newCustomerFixture(t *testing.T) (service.CustomerService, *fixtureDB) {
	t.Helper()
	db := testdb.New(t, 0)
	db.MustExec(`INSERT INTO users (username, password) VALUES ('tester', 'x')`)
	f := &fixtureDB{db: db}
	if err := db.Get(&f.uid, `SELECT id FROM users LIMIT 1`); err != nil {
		t.Fatal(err)
	}
	return service.NewCustomerService(repository.NewStore(db)), f
}

type fixtureDB struct {
	db  *sqlx.DB
	uid int
}

var customerCtx = asUser(constant.ViewCustomersPermissionKey, constant.ManageCustomersPermissionKey)

func mustCreate(t *testing.T, svc service.CustomerService, uid int, req types.CustomerRequest) int {
	t.Helper()
	id, err := svc.CreateCustomer(customerCtx, req, uid)
	if err != nil {
		t.Fatalf("create %+v: %v", req, err)
	}
	return id
}

func requireStatus(t *testing.T, err error, code int) {
	t.Helper()
	var httperr *httperror.HTTPError
	if !errors.As(err, &httperr) || httperr.Code != code {
		t.Fatalf("expected status %d, got %v", code, err)
	}
}

func TestCustomerLifecycleAgainstTheDatabase(t *testing.T) {
	svc, f := newCustomerFixture(t)
	ada := mustCreate(t, svc, f.uid, types.CustomerRequest{
		Name: "Ada Obi", Phone: "0803 123 4567", Email: "ada@example.com", ContactConsent: true,
	})
	mustCreate(t, svc, f.uid, types.CustomerRequest{Name: "Bola 50%_off"})
	// Has 803 inside the number, but does not start with 0803.
	mustCreate(t, svc, f.uid, types.CustomerRequest{Name: "Chika", Phone: "0703 803 1111"})

	got, err := svc.GetCustomer(customerCtx, ada)
	if err != nil {
		t.Fatal(err)
	}
	given := got
	if got.Phone != "+2348031234567" || got.Version != 1 || !got.Active || !got.ContactConsent ||
		got.ConsentUpdatedAt == nil {
		t.Fatalf("stored customer %+v", got)
	}

	// The same number is refused however it is written.
	_, err = svc.CreateCustomer(customerCtx, types.CustomerRequest{
		Name: "Someone", Phone: "+234 (0)803-123-4567",
	}, f.uid)
	requireStatus(t, err, http.StatusConflict)
	_, err = svc.CreateCustomer(customerCtx, types.CustomerRequest{
		Name: "Someone", Phone: "2348031234567",
	}, f.uid)
	requireStatus(t, err, http.StatusConflict)

	// Search by name (any case, part of a word) and by phone written any way.
	for search, want := range map[string]int{
		"ada": 1, "OBI": 1, "bola": 1, "50%_off": 1, "_off": 1, "zzz": 0,
		// % and _ typed in a name search are text, not wildcards.
		"o_f": 0, "a%o": 0, "x%": 0,
		"0803": 1, "+234 803 123": 1, "8031234567": 1, "09": 0,
		// 803 not at the start of the number matches wherever it appears; 0803 only at the start.
		"803": 2, "2348031234": 1, "0703": 1, "0": 2, "+234": 2, "+44": 0,
	} {
		page, err := svc.ListCustomers(customerCtx, search, false, 1)
		if err != nil {
			t.Fatal(err)
		}
		if page.TotalCount != want || len(page.Data) != want {
			t.Errorf(
				"search %q found %d (%d listed), want %d",
				search,
				page.TotalCount,
				len(page.Data),
				want,
			)
		}
	}
	// Edit: stale copies are refused, a good edit moves the version on.
	edit := types.CustomerRequest{
		ID:             ada,
		Version:        1,
		Name:           "Ada Okafor",
		Phone:          "08031234567",
		ContactConsent: false,
	}
	if err = svc.UpdateCustomer(customerCtx, edit, f.uid); err != nil {
		t.Fatal(err)
	}
	requireStatus(t, svc.UpdateCustomer(customerCtx, edit, f.uid), http.StatusConflict)
	got, _ = svc.GetCustomer(customerCtx, ada)
	if got.Name != "Ada Okafor" || got.Version != 2 || got.ContactConsent || got.Email != "" {
		t.Fatalf("after edit %+v", got)
	}
	// Withdrawing consent moved its time on; an edit that leaves consent alone does not.
	if got.ConsentUpdatedAt == nil || !got.ConsentUpdatedAt.After(*given.ConsentUpdatedAt) {
		t.Fatalf("consent time should move when consent changes: %v then %v",
			given.ConsentUpdatedAt, got.ConsentUpdatedAt)
	}
	withdrawn := *got.ConsentUpdatedAt
	rename := types.CustomerRequest{ID: ada, Version: 2, Name: "Ada O.", Phone: "08031234567"}
	if err = svc.UpdateCustomer(customerCtx, rename, f.uid); err != nil {
		t.Fatal(err)
	}
	if got, _ = svc.GetCustomer(customerCtx, ada); !got.ConsentUpdatedAt.Equal(withdrawn) ||
		got.Version != 3 {
		t.Fatalf("consent time must stay when consent is unchanged: %+v", got)
	}

	// Deactivated customers leave the default list but keep their number.
	if err = svc.SetCustomerActive(customerCtx, ada, 3, false, f.uid); err != nil {
		t.Fatal(err)
	}
	if page, _ := svc.ListCustomers(customerCtx, "ada", false, 1); page.TotalCount != 0 {
		t.Errorf("deactivated customers are hidden by default, found %d", page.TotalCount)
	}
	if page, _ := svc.ListCustomers(customerCtx, "ada", true, 1); page.TotalCount != 1 {
		t.Errorf("deactivated customers show when asked, found %d", page.TotalCount)
	}
	_, err = svc.CreateCustomer(
		customerCtx,
		types.CustomerRequest{Name: "New", Phone: "08031234567"},
		f.uid,
	)
	requireStatus(t, err, http.StatusConflict)

	// Everything above was audited, in order.
	var actions []string
	err = f.db.Select(
		&actions,
		`SELECT action FROM audit_log WHERE entity_type = 'customer' AND entity_id = $1 ORDER BY id`,
		ada,
	)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(actions, ",") != "created,updated,updated,deactivated" {
		t.Fatalf("audit actions %v", actions)
	}
}

func TestTwoRequestsForTheSamePhoneSaveOneCustomer(t *testing.T) {
	svc, f := newCustomerFixture(t)
	const racers = 8
	var wg sync.WaitGroup
	results := make(chan error, racers)
	for i := 0; i < racers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.CreateCustomer(customerCtx, types.CustomerRequest{
				Name: "Racer", Phone: "08031234567",
			}, f.uid)
			results <- err
		}()
	}
	wg.Wait()
	close(results)

	saved := 0
	for err := range results {
		if err == nil {
			saved++
			continue
		}
		requireStatus(t, err, http.StatusConflict)
	}
	if saved != 1 {
		t.Fatalf("exactly one request should save the customer, %d did", saved)
	}
	page, _ := svc.ListCustomers(customerCtx, "", true, 1)
	if page.TotalCount != 1 {
		t.Fatalf("expected one customer, found %d", page.TotalCount)
	}
}

func TestListCustomersPages(t *testing.T) {
	svc, f := newCustomerFixture(t)
	for i := 0; i < 27; i++ {
		mustCreate(
			t,
			svc,
			f.uid,
			types.CustomerRequest{
				Name: "Customer " + string(rune('A'+i%26)) + string(rune('a'+i/26)),
			},
		)
	}
	first, _ := svc.ListCustomers(customerCtx, "", false, 1)
	second, _ := svc.ListCustomers(customerCtx, "", false, 2)
	if first.TotalCount != 27 || len(first.Data) != 25 || len(second.Data) != 2 {
		t.Fatalf(
			"total %d, page sizes %d and %d",
			first.TotalCount,
			len(first.Data),
			len(second.Data),
		)
	}
	if first.Data[24].ID == second.Data[0].ID {
		t.Fatal("pages must not overlap")
	}
}
