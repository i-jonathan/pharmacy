package service_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"pharmacy/httperror"
	"pharmacy/internal/constant"
	"pharmacy/internal/testdb"
	"pharmacy/internal/types"
	"pharmacy/repository"
	"pharmacy/service"
)

// These run the real service and SQL against a throwaway database; see internal/testdb.

var (
	recorderCtx = asUser(
		constant.RecordHealthCheckPermissionKey,
		constant.ViewCustomersPermissionKey,
	)
	managerCtx = asUser(
		constant.RecordHealthCheckPermissionKey, constant.ViewCustomersPermissionKey,
		constant.ManageCustomersPermissionKey, constant.ViewHealthChecksPermissionKey,
		constant.CorrectHealthCheckPermissionKey,
	)
)

type healthFixture struct {
	*fixtureDB
	svc       service.HealthCheckService
	customers service.CustomerService
	other     int // a second user
	ada       int // a customer with health consent
	bola      int // a customer without
}

func newHealthFixture(t *testing.T) *healthFixture {
	t.Helper()
	db := testdb.New(t, 0)
	db.MustExec(`INSERT INTO users (username, password) VALUES ('tunde', 'x'), ('kemi', 'x')`)
	f := &healthFixture{fixtureDB: &fixtureDB{db: db}}
	if err := db.Get(&f.uid, `SELECT id FROM users WHERE username = 'tunde'`); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&f.other, `SELECT id FROM users WHERE username = 'kemi'`); err != nil {
		t.Fatal(err)
	}
	store := repository.NewStore(db)
	f.svc = service.NewHealthCheckService(store)
	f.customers = service.NewCustomerService(store)
	f.ada = mustCreate(t, f.customers, f.uid, types.CustomerRequest{Name: "Ada Obi"})
	f.bola = mustCreate(t, f.customers, f.uid, types.CustomerRequest{Name: "Bola"})
	db.MustExec(`UPDATE customer SET health_consent = TRUE WHERE id = $1`, f.ada)
	return f
}

func ip(v int) *int         { return &v }
func fp(v float64) *float64 { return &v }

func bpReading(customer, sys, dia int, key string) types.HealthCheckRequest {
	return types.HealthCheckRequest{
		CustomerID:     customer,
		HealthReadings: types.HealthReadings{Systolic: ip(sys), Diastolic: ip(dia), Pulse: ip(72)},
		IdempotencyKey: key,
	}
}

func requireCodeAndReason(t *testing.T, err error, status int, reason string) {
	t.Helper()
	var httperr *httperror.HTTPError
	if !errors.As(err, &httperr) || httperr.Code != status || httperr.Reason != reason {
		t.Fatalf("expected %d %q, got %v", status, reason, err)
	}
}

func (f *healthFixture) audits(t *testing.T, entity string, id int) []string {
	t.Helper()
	var actions []string
	err := f.db.Select(&actions, `
		SELECT action FROM audit_log WHERE entity_type = $1 AND entity_id = $2 ORDER BY id`,
		entity, id)
	if err != nil {
		t.Fatal(err)
	}
	return actions
}

func TestRecordHealthCheckSavesAndAnswersARetry(t *testing.T) {
	f := newHealthFixture(t)
	req := bpReading(f.ada, 120, 80, "key-1")
	req.Glucose, req.GlucoseUnit, req.GlucoseContext = fp(5.6), "mmol/l", "Fasting"
	req.Note = " sat down first "

	saved, err := f.svc.RecordHealthCheck(recorderCtx, req, f.uid)
	if err != nil {
		t.Fatal(err)
	}
	if saved.CustomerName != "Ada Obi" || saved.TakenBy != "tunde" || *saved.Systolic != 120 ||
		*saved.Glucose != 5.6 || *saved.GlucoseUnit != "mmol/L" ||
		*saved.GlucoseContext != "fasting" || saved.Note != "sat down first" ||
		saved.Version != 1 || saved.Corrected {
		t.Fatalf("saved %+v", saved)
	}

	// The same request again (a double click, a lost reply) is the same save.
	again, err := f.svc.RecordHealthCheck(recorderCtx, req, f.uid)
	if err != nil || again.ID != saved.ID {
		t.Fatalf("retry: %v, id %d want %d", err, again.ID, saved.ID)
	}
	// Even after the customer is deactivated.
	f.db.MustExec(`UPDATE customer SET active = FALSE WHERE id = $1`, f.ada)
	if again, err = f.svc.RecordHealthCheck(recorderCtx, req, f.uid); err != nil ||
		again.ID != saved.ID {
		t.Fatalf("retry after deactivation: %v", err)
	}
	f.db.MustExec(`UPDATE customer SET active = TRUE WHERE id = $1`, f.ada)

	// The key with anything else, or from someone else, is refused.
	other := bpReading(f.ada, 130, 80, "key-1")
	_, err = f.svc.RecordHealthCheck(recorderCtx, other, f.uid)
	requireCodeAndReason(t, err, http.StatusConflict, "already_saved")
	_, err = f.svc.RecordHealthCheck(recorderCtx, req, f.other)
	requireCodeAndReason(t, err, http.StatusConflict, "already_saved")

	var n int
	f.db.Get(&n, `SELECT count(*) FROM health_check`)
	if n != 1 {
		t.Fatalf("expected one saved reading, got %d", n)
	}
	if got := f.audits(t, "health_check", saved.ID); len(got) != 1 || got[0] != "recorded" {
		t.Fatalf("audit %v", got)
	}
}

func TestRecordHealthCheckRefusesUnknownAndDeactivatedCustomers(t *testing.T) {
	f := newHealthFixture(t)
	_, err := f.svc.RecordHealthCheck(recorderCtx, bpReading(999, 120, 80, "k1"), f.uid)
	requireStatus(t, err, http.StatusBadRequest)

	f.db.MustExec(`UPDATE customer SET active = FALSE WHERE id = $1`, f.ada)
	_, err = f.svc.RecordHealthCheck(recorderCtx, bpReading(f.ada, 120, 80, "k2"), f.uid)
	requireCodeAndReason(t, err, http.StatusConflict, "customer_inactive")
}

func TestRecordHealthCheckNeedsConsentOnFile(t *testing.T) {
	f := newHealthFixture(t)
	req := bpReading(f.bola, 120, 80, "k1")
	_, err := f.svc.RecordHealthCheck(recorderCtx, req, f.uid)
	requireCodeAndReason(t, err, http.StatusConflict, "health_consent_needed")
	var n int
	f.db.Get(&n, `SELECT count(*) FROM health_check`)
	if n != 0 {
		t.Fatal("nothing may be saved without consent")
	}

	// Asked at the counter: the consent is recorded with the reading.
	req.ConsentGiven = true
	if _, err = f.svc.RecordHealthCheck(recorderCtx, req, f.uid); err != nil {
		t.Fatal(err)
	}
	bola, _ := f.customers.GetCustomer(customerCtx, f.bola)
	if !bola.HealthConsent || bola.HealthConsentUpdatedAt == nil || bola.Version != 2 {
		t.Fatalf("consent not recorded: %+v", bola)
	}
	if got := f.audits(t, "customer", f.bola); len(got) != 2 ||
		got[1] != "health_consent_changed" {
		t.Fatalf("customer audit %v", got)
	}
	// Once on file it is not needed again.
	if _, err = f.svc.RecordHealthCheck(
		recorderCtx,
		bpReading(f.bola, 125, 80, "k2"),
		f.uid,
	); err != nil {
		t.Fatal(err)
	}

	// Withdrawn on the customer's record: new readings stop, saved ones stay.
	no := false
	err = f.customers.UpdateCustomer(customerCtx, types.CustomerRequest{
		ID: f.bola, Version: 2, Name: "Bola", HealthConsent: &no,
	}, f.uid)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.svc.RecordHealthCheck(recorderCtx, bpReading(f.bola, 130, 80, "k3"), f.uid)
	requireCodeAndReason(t, err, http.StatusConflict, "health_consent_needed")
	f.db.Get(&n, `SELECT count(*) FROM health_check WHERE customer_id = $1`, f.bola)
	if n != 2 {
		t.Fatalf("saved readings must stay when consent is withdrawn, found %d", n)
	}
}

func TestCustomerEditLeavesHealthConsentAloneWhenNotSent(t *testing.T) {
	f := newHealthFixture(t)
	ada, _ := f.customers.GetCustomer(customerCtx, f.ada)
	err := f.customers.UpdateCustomer(customerCtx, types.CustomerRequest{
		ID: f.ada, Version: ada.Version, Name: "Ada Okafor",
	}, f.uid)
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := f.customers.GetCustomer(customerCtx, f.ada); !got.HealthConsent {
		t.Fatal("an edit without health_consent must not clear it")
	}

	var details []byte
	f.db.Get(&details, `
		SELECT details FROM audit_log WHERE entity_type = 'customer' AND entity_id = $1
		ORDER BY id DESC LIMIT 1`, f.ada)
	var d struct{ Changed []string }
	json.Unmarshal(details, &d)
	if len(d.Changed) != 1 || d.Changed[0] != "name" {
		t.Fatalf("audit changed %v", d.Changed)
	}
}

func TestHealthHistoryIsNewestFirstPagedAndAudited(t *testing.T) {
	f := newHealthFixture(t)
	for i := 0; i < 22; i++ {
		_, err := f.svc.RecordHealthCheck(
			recorderCtx, bpReading(f.ada, 100+i, 70, "k"+string(rune('a'+i))), f.uid,
		)
		if err != nil {
			t.Fatal(err)
		}
	}
	page, err := f.svc.CustomerHealthChecks(managerCtx, f.ada, 1, f.other)
	if err != nil {
		t.Fatal(err)
	}
	if page.TotalCount != 22 || len(page.Data) != 20 || *page.Data[0].Systolic != 121 ||
		*page.Data[19].Systolic != 102 {
		t.Fatalf("page 1: total %d, %d rows", page.TotalCount, len(page.Data))
	}
	page, err = f.svc.CustomerHealthChecks(managerCtx, f.ada, 2, f.other)
	if err != nil || len(page.Data) != 2 || *page.Data[1].Systolic != 100 {
		t.Fatalf("page 2: %v", err)
	}
	if got := f.audits(t, "customer", f.ada); len(got) != 3 ||
		got[1] != "health_history_viewed" || got[2] != "health_history_viewed" {
		t.Fatalf("customer audit %v", got)
	}

	_, err = f.svc.CustomerHealthChecks(managerCtx, 999, 1, f.other)
	requireStatus(t, err, http.StatusNotFound)
}

func TestCorrectHealthCheckKeepsATrail(t *testing.T) {
	f := newHealthFixture(t)
	saved, err := f.svc.RecordHealthCheck(recorderCtx, bpReading(f.ada, 180, 110, "k1"), f.uid)
	if err != nil {
		t.Fatal(err)
	}
	fix := types.HealthCheckCorrection{
		ID: saved.ID, Version: 1, Reason: "typed the wrong numbers",
		HealthReadings: types.HealthReadings{
			Systolic: ip(130), Diastolic: ip(80), Pulse: ip(72),
			Glucose: fp(6.1), GlucoseUnit: "mmol/L", GlucoseContext: "random",
		},
	}
	got, err := f.svc.CorrectHealthCheck(managerCtx, fix, f.other)
	if err != nil {
		t.Fatal(err)
	}
	if *got.Systolic != 130 || *got.Glucose != 6.1 || got.Version != 2 || !got.Corrected ||
		got.TakenBy != "tunde" {
		t.Fatalf("corrected %+v", got)
	}

	var entry struct {
		Reason  string `db:"reason"`
		Details []byte `db:"details"`
		Actor   int    `db:"actor_id"`
	}
	err = f.db.Get(&entry, `
		SELECT reason, details, actor_id FROM audit_log
		WHERE entity_type = 'health_check' AND action = 'corrected'`)
	if err != nil {
		t.Fatal(err)
	}
	var details struct {
		Changed map[string]struct{ From, To any }
	}
	json.Unmarshal(entry.Details, &details)
	c := details.Changed
	if entry.Reason != "typed the wrong numbers" || entry.Actor != f.other || len(c) != 5 ||
		c["systolic"].From != 180.0 || c["systolic"].To != 130.0 ||
		c["glucose"].From != nil || c["glucose"].To != 6.1 ||
		c["glucose_unit"].To != "mmol/L" || c["glucose_context"].To != "random" {
		t.Fatalf("audit %s %s", entry.Reason, entry.Details)
	}

	// A stale copy is refused; the same numbers again write nothing.
	_, err = f.svc.CorrectHealthCheck(managerCtx, fix, f.other)
	requireCodeAndReason(t, err, http.StatusConflict, "changed")
	fix.Version = 2
	if got, err = f.svc.CorrectHealthCheck(managerCtx, fix, f.other); err != nil ||
		got.Version != 2 {
		t.Fatalf("no-op correction: %v, version %d", err, got.Version)
	}
	if a := f.audits(t, "health_check", saved.ID); len(a) != 2 {
		t.Fatalf("a no-op correction must not be audited: %v", a)
	}

	fix.ID = 999
	_, err = f.svc.CorrectHealthCheck(managerCtx, fix, f.other)
	requireStatus(t, err, http.StatusNotFound)
}

func TestReprintHealthCheckIsAudited(t *testing.T) {
	f := newHealthFixture(t)
	saved, err := f.svc.RecordHealthCheck(recorderCtx, bpReading(f.ada, 120, 80, "k1"), f.uid)
	if err != nil {
		t.Fatal(err)
	}
	got, err := f.svc.ReprintHealthCheck(managerCtx, saved.ID, f.other)
	if err != nil || got.ID != saved.ID || *got.Diastolic != 80 {
		t.Fatalf("reprint: %v %+v", err, got)
	}
	if a := f.audits(t, "health_check", saved.ID); len(a) != 2 || a[1] != "reprinted" {
		t.Fatalf("audit %v", a)
	}
	_, err = f.svc.ReprintHealthCheck(managerCtx, 999, f.other)
	requireStatus(t, err, http.StatusNotFound)
}

func TestRecordHealthCheckTwiceAtOnceSavesOnce(t *testing.T) {
	f := newHealthFixture(t)
	// Bola has no consent yet, so both copies also try to record it.
	req := bpReading(f.bola, 120, 80, "double-click")
	req.ConsentGiven = true
	ids := make(chan int, 2)
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		go func() {
			saved, err := f.svc.RecordHealthCheck(recorderCtx, req, f.uid)
			ids <- saved.ID
			errs <- err
		}()
	}
	first, second := <-ids, <-ids
	for i := 0; i < 2; i++ {
		if err := <-errs; err != nil {
			t.Fatal(err)
		}
	}
	if first != second {
		t.Fatalf("both copies must answer with one save, got %d and %d", first, second)
	}
	var n int
	f.db.Get(&n, `SELECT count(*) FROM audit_log WHERE action = 'health_consent_changed'`)
	if n != 1 {
		t.Fatalf("consent must be recorded once, got %d", n)
	}
}

func TestHealthConsentGivenAtCreateAndChangedByEditIsSavedAndAudited(t *testing.T) {
	f := newHealthFixture(t)
	yes, no := true, false
	id, err := f.customers.CreateCustomer(customerCtx, types.CustomerRequest{
		Name: "Chika", HealthConsent: &yes,
	}, f.uid)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := f.customers.GetCustomer(customerCtx, id)
	if !got.HealthConsent || got.HealthConsentUpdatedAt == nil {
		t.Fatalf("consent given at create was not saved: %+v", got)
	}
	if _, err = f.svc.RecordHealthCheck(
		recorderCtx,
		bpReading(id, 120, 80, "k1"),
		f.uid,
	); err != nil {
		t.Fatalf("consent from create must let a reading be saved: %v", err)
	}

	err = f.customers.UpdateCustomer(customerCtx, types.CustomerRequest{
		ID: id, Version: got.Version, Name: "Chika", HealthConsent: &no,
	}, f.uid)
	if err != nil {
		t.Fatal(err)
	}
	// Every grant and withdrawal is findable under one action.
	var changes []string
	err = f.db.Select(&changes, `
		SELECT details->>'to' FROM audit_log
		WHERE entity_type = 'customer' AND entity_id = $1 AND action = 'health_consent_changed'
		ORDER BY id`, id)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 2 || changes[0] != "true" || changes[1] != "false" {
		t.Fatalf("consent trail %v", changes)
	}
}

func TestCorrectHealthCheckWithoutAVersionIsABadRequest(t *testing.T) {
	f := newHealthFixture(t)
	saved, err := f.svc.RecordHealthCheck(recorderCtx, bpReading(f.ada, 120, 80, "k1"), f.uid)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.svc.CorrectHealthCheck(managerCtx, types.HealthCheckCorrection{
		ID: saved.ID, Reason: "typo",
		HealthReadings: types.HealthReadings{Systolic: ip(130), Diastolic: ip(80)},
	}, f.uid)
	requireStatus(t, err, http.StatusBadRequest)
}

func TestCorrectionThatAddsAMeasurementNeedsConsentAndAnActiveCustomer(t *testing.T) {
	f := newHealthFixture(t)
	saved, err := f.svc.RecordHealthCheck(recorderCtx, bpReading(f.ada, 120, 80, "k1"), f.uid)
	if err != nil {
		t.Fatal(err)
	}
	addGlucose := types.HealthCheckCorrection{
		ID: saved.ID, Version: 1, Reason: "forgot the glucose",
		HealthReadings: types.HealthReadings{
			Systolic: ip(120), Diastolic: ip(80), Pulse: ip(72),
			Glucose: fp(6.1), GlucoseUnit: "mmol/L", GlucoseContext: "random",
		},
	}
	fixTypo := types.HealthCheckCorrection{
		ID: saved.ID, Version: 1, Reason: "typo",
		HealthReadings: types.HealthReadings{Systolic: ip(125), Diastolic: ip(80), Pulse: ip(72)},
	}

	f.db.MustExec(`UPDATE customer SET health_consent = FALSE WHERE id = $1`, f.ada)
	_, err = f.svc.CorrectHealthCheck(managerCtx, addGlucose, f.other)
	requireCodeAndReason(t, err, http.StatusConflict, "health_consent_needed")
	// Fixing numbers already stored keeps the record accurate, consent or not.
	if _, err = f.svc.CorrectHealthCheck(managerCtx, fixTypo, f.other); err != nil {
		t.Fatalf("a typo fix after withdrawal: %v", err)
	}

	f.db.MustExec(`UPDATE customer SET health_consent = TRUE, active = FALSE WHERE id = $1`, f.ada)
	addGlucose.Version = 2
	_, err = f.svc.CorrectHealthCheck(managerCtx, addGlucose, f.other)
	requireCodeAndReason(t, err, http.StatusConflict, "customer_inactive")

	f.db.MustExec(`UPDATE customer SET active = TRUE WHERE id = $1`, f.ada)
	if _, err = f.svc.CorrectHealthCheck(managerCtx, addGlucose, f.other); err != nil {
		t.Fatalf("adding glucose with consent: %v", err)
	}
}
