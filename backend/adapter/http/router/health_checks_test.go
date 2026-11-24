package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"strings"
	"testing"
)

// fakeHealthCheckService records the calls that reach the service.
type fakeHealthCheckService struct {
	record     types.HealthCheckRequest
	correction types.HealthCheckCorrection
	id         int
	page       int
	userID     int
}

func (f *fakeHealthCheckService) RecordHealthCheck(
	_ context.Context, req types.HealthCheckRequest, userID int,
) (types.HealthCheck, error) {
	f.record, f.userID = req, userID
	return types.HealthCheck{ID: 31}, nil
}

func (f *fakeHealthCheckService) CustomerHealthChecks(
	_ context.Context, customerID, page, userID int,
) (types.HealthCheckPage, error) {
	f.id, f.page, f.userID = customerID, page, userID
	return types.HealthCheckPage{Data: []types.HealthCheck{}, Page: page}, nil
}

func (f *fakeHealthCheckService) CorrectHealthCheck(
	_ context.Context, req types.HealthCheckCorrection, userID int,
) (types.HealthCheck, error) {
	f.correction, f.userID = req, userID
	return types.HealthCheck{ID: req.ID}, nil
}

func (f *fakeHealthCheckService) ReprintHealthCheck(
	_ context.Context, id, userID int,
) (types.HealthCheck, error) {
	f.id, f.userID = id, userID
	return types.HealthCheck{ID: id}, nil
}

func TestHealthCheckRoutesNeedTheirPermissions(t *testing.T) {
	h := InitHealthCheckRouter(&fakeHealthCheckService{})
	cashier := []string{
		constant.RecordHealthCheckPermissionKey, constant.ViewCustomersPermissionKey,
	}
	manager := append([]string{
		constant.ViewHealthChecksPermissionKey, constant.CorrectHealthCheckPermissionKey,
	}, cashier...)
	for _, r := range []struct {
		method, path string
		cashierMay   bool
	}{
		{http.MethodPost, "/health-checks/", true},
		{http.MethodGet, "/health-checks/customer/4", false},
		{http.MethodPut, "/health-checks/9", false},
		{http.MethodPost, "/health-checks/9/reprint", false},
	} {
		send := func(perms ...string) int {
			req := httptest.NewRequest(r.method, r.path, strings.NewReader("{}"))
			return serve(h, withSession(t, req, perms...)).Code
		}
		if code := send(); code != http.StatusForbidden {
			t.Errorf("%s %s with no permission: want 403, got %d", r.method, r.path, code)
		}
		want := http.StatusForbidden
		if r.cashierMay {
			want = http.StatusCreated
		}
		if code := send(cashier...); code != want {
			t.Errorf("%s %s as a cashier: want %d, got %d", r.method, r.path, want, code)
		}
		if code := send(manager...); code >= 400 {
			t.Errorf("%s %s as a manager: got %d", r.method, r.path, code)
		}
	}
	// Recording needs customer:view as well, and correcting needs healthcheck:view.
	req := httptest.NewRequest(http.MethodPost, "/health-checks/", strings.NewReader("{}"))
	if code := serve(
		h,
		withSession(t, req, constant.RecordHealthCheckPermissionKey),
	).Code; code != http.StatusForbidden {
		t.Errorf("record without customer:view: want 403, got %d", code)
	}
	req = httptest.NewRequest(http.MethodPut, "/health-checks/9", strings.NewReader("{}"))
	if code := serve(
		h,
		withSession(t, req, constant.CorrectHealthCheckPermissionKey),
	).Code; code != http.StatusForbidden {
		t.Errorf("correct without healthcheck:view: want 403, got %d", code)
	}
}

func TestRecordHealthCheckRoute(t *testing.T) {
	svc := &fakeHealthCheckService{}
	body := `{"customer_id":4,"systolic":120,"diastolic":80,"glucose":5.6,` +
		`"glucose_unit":"mmol/L","glucose_context":"fasting","consent_given":true,` +
		`"idempotency_key":"k"}`
	req := withSession(t,
		httptest.NewRequest(http.MethodPost, "/health-checks/", strings.NewReader(body)),
		constant.RecordHealthCheckPermissionKey, constant.ViewCustomersPermissionKey)

	rec := serve(InitHealthCheckRouter(svc), req)

	got := svc.record
	if rec.Code != http.StatusCreated || !strings.Contains(rec.Body.String(), `"id":31`) ||
		got.CustomerID != 4 || *got.Systolic != 120 || *got.Glucose != 5.6 ||
		got.GlucoseUnit != "mmol/L" || !got.ConsentGiven || got.IdempotencyKey != "k" ||
		svc.userID != 7 {
		t.Fatalf("got %d %s, service saw %+v", rec.Code, rec.Body.String(), got)
	}
}

func TestCorrectHealthCheckRouteTakesTheIDFromThePath(t *testing.T) {
	svc := &fakeHealthCheckService{}
	req := withSession(t,
		httptest.NewRequest(http.MethodPut, "/health-checks/9",
			strings.NewReader(`{"version":2,"reason":"typo","systolic":130,"diastolic":80}`)),
		constant.CorrectHealthCheckPermissionKey, constant.ViewHealthChecksPermissionKey)

	rec := serve(InitHealthCheckRouter(svc), req)

	c := svc.correction
	if rec.Code != http.StatusOK || c.ID != 9 || c.Version != 2 || c.Reason != "typo" ||
		*c.Systolic != 130 {
		t.Fatalf("got %d, service saw %+v", rec.Code, c)
	}
}

func TestHealthHistoryRoutePassesTheCustomerAndPage(t *testing.T) {
	svc := &fakeHealthCheckService{}
	req := withSession(t,
		httptest.NewRequest(http.MethodGet, "/health-checks/customer/4?page=3", nil),
		constant.ViewHealthChecksPermissionKey)

	rec := serve(InitHealthCheckRouter(svc), req)

	if rec.Code != http.StatusOK || svc.id != 4 || svc.page != 3 || svc.userID != 7 {
		t.Fatalf("got %d, customer %d, page %d", rec.Code, svc.id, svc.page)
	}
	req = withSession(t,
		httptest.NewRequest(http.MethodGet, "/health-checks/customer/abc", nil),
		constant.ViewHealthChecksPermissionKey)
	if rec = serve(InitHealthCheckRouter(svc), req); rec.Code != http.StatusBadRequest {
		t.Fatalf("a bad id: want 400, got %d", rec.Code)
	}
}

func TestUnknownHealthCheckRoutesAnswerJSON(t *testing.T) {
	h := InitHealthCheckRouter(&fakeHealthCheckService{})
	for _, r := range []struct{ method, path string }{
		{http.MethodGet, "/health-checks/9"},
		{http.MethodDelete, "/health-checks/9"},
		{http.MethodGet, "/health-checks/nothing/here"},
	} {
		req := withSession(t, httptest.NewRequest(r.method, r.path, nil),
			constant.ViewHealthChecksPermissionKey)
		rec := serve(h, req)
		if rec.Code != http.StatusNotFound ||
			!strings.HasPrefix(rec.Header().Get("Content-Type"), "application/json") {
			t.Errorf("%s %s: want a JSON 404, got %d %q", r.method, r.path, rec.Code,
				rec.Header().Get("Content-Type"))
		}
	}
}
