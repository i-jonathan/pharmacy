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

// fakeCustomerService records the calls that reach the service.
type fakeCustomerService struct {
	request types.CustomerRequest
	userID  int
	id      int
	version int
	active  *bool
	search  string
	showOff bool
	page    int
}

func (f *fakeCustomerService) CreateCustomer(
	_ context.Context, req types.CustomerRequest, userID int,
) (int, error) {
	f.request, f.userID = req, userID
	return 51, nil
}

func (f *fakeCustomerService) UpdateCustomer(
	_ context.Context, req types.CustomerRequest, userID int,
) error {
	f.request, f.userID = req, userID
	return nil
}

func (f *fakeCustomerService) SetCustomerActive(
	_ context.Context, id, version int, active bool, userID int,
) error {
	f.id, f.version, f.active, f.userID = id, version, &active, userID
	return nil
}

func (f *fakeCustomerService) GetCustomer(_ context.Context, id int) (types.Customer, error) {
	f.id = id
	return types.Customer{ID: id, Name: "Ada", Active: true}, nil
}

func (f *fakeCustomerService) ListCustomers(
	_ context.Context, search string, includeInactive bool, page int,
) (types.CustomerPage, error) {
	f.search, f.showOff, f.page = search, includeInactive, page
	return types.CustomerPage{Page: page, Data: []types.Customer{}}, nil
}

func TestCustomerRoutesNeedTheirPermissions(t *testing.T) {
	h := InitCustomerRouter(&fakeCustomerService{}, errorPage)
	reads := []struct{ method, path string }{
		{http.MethodGet, "/customers/"},
		{http.MethodGet, "/customers/4"},
	}
	writes := []struct{ method, path string }{
		{http.MethodPost, "/customers/"},
		{http.MethodPut, "/customers/4"},
		{http.MethodPost, "/customers/4/deactivate"},
		{http.MethodPost, "/customers/4/reactivate"},
	}
	for _, r := range reads {
		req := withSession(t, httptest.NewRequest(r.method, r.path, nil))
		if rec := serve(h, req); rec.Code != http.StatusForbidden {
			t.Errorf("%s %s with no permission: want 403, got %d", r.method, r.path, rec.Code)
		}
		req = withSession(t, httptest.NewRequest(r.method, r.path, nil),
			constant.ViewCustomersPermissionKey)
		if rec := serve(h, req); rec.Code != http.StatusOK {
			t.Errorf("%s %s with customer:view: want 200, got %d", r.method, r.path, rec.Code)
		}
	}
	for _, r := range writes {
		req := withSession(t, httptest.NewRequest(r.method, r.path, strings.NewReader("{}")),
			constant.ViewCustomersPermissionKey)
		if rec := serve(h, req); rec.Code != http.StatusForbidden {
			t.Errorf("%s %s with only customer:view: want 403, got %d", r.method, r.path, rec.Code)
		}
	}
}

func TestCreateCustomerRoute(t *testing.T) {
	svc := &fakeCustomerService{}
	body := `{"name":"Ada","phone":"0803 123 4567","email":"a@b.co","notes":"n","contact_consent":true}`
	req := withSession(t,
		httptest.NewRequest(http.MethodPost, "/customers/", strings.NewReader(body)),
		constant.ManageCustomersPermissionKey)

	rec := serve(InitCustomerRouter(svc, errorPage), req)

	if rec.Code != http.StatusCreated || !strings.Contains(rec.Body.String(), `"id":51`) {
		t.Fatalf("expected 201 {id:51}, got %d %s", rec.Code, rec.Body.String())
	}
	got := svc.request
	if got.Name != "Ada" || got.Phone != "0803 123 4567" || !got.ContactConsent || svc.userID != 7 {
		t.Fatalf("the service got %+v from user %d", got, svc.userID)
	}
}

func TestUpdateCustomerRouteTakesTheIDFromThePath(t *testing.T) {
	svc := &fakeCustomerService{}
	req := withSession(t,
		httptest.NewRequest(http.MethodPut, "/customers/4",
			strings.NewReader(`{"id":99,"version":3,"name":"Ada"}`)),
		constant.ManageCustomersPermissionKey)

	rec := serve(InitCustomerRouter(svc, errorPage), req)

	if rec.Code != http.StatusNoContent || svc.request.ID != 4 || svc.request.Version != 3 {
		t.Fatalf("expected customer 4 at version 3, got %d %+v", rec.Code, svc.request)
	}
}

func TestDeactivateAndReactivateRoutes(t *testing.T) {
	for path, want := range map[string]bool{
		"/customers/4/deactivate": false, "/customers/4/reactivate": true,
	} {
		svc := &fakeCustomerService{}
		req := withSession(t,
			httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"version":3}`)),
			constant.ManageCustomersPermissionKey)

		rec := serve(InitCustomerRouter(svc, errorPage), req)

		if rec.Code != http.StatusNoContent || svc.id != 4 || svc.version != 3 ||
			svc.active == nil || *svc.active != want {
			t.Errorf(
				"%s: got %d, id %d, version %d, active %v",
				path,
				rec.Code,
				svc.id,
				svc.version,
				svc.active,
			)
		}
	}
}

func TestListCustomersRoutePassesTheSearchAndPage(t *testing.T) {
	svc := &fakeCustomerService{}
	req := withSession(t,
		httptest.NewRequest(http.MethodGet, "/customers/?q=0803&page=2&inactive=1", nil),
		constant.ViewCustomersPermissionKey)

	rec := serve(InitCustomerRouter(svc, errorPage), req)

	if rec.Code != http.StatusOK || svc.search != "0803" || svc.page != 2 || !svc.showOff {
		t.Fatalf(
			"got %d, search %q, page %d, inactive %v",
			rec.Code,
			svc.search,
			svc.page,
			svc.showOff,
		)
	}
}

func TestCustomerRouteRefusesABadID(t *testing.T) {
	req := withSession(t, httptest.NewRequest(http.MethodGet, "/customers/abc", nil),
		constant.ViewCustomersPermissionKey)
	if rec := serve(
		InitCustomerRouter(&fakeCustomerService{}, errorPage),
		req,
	); rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}
