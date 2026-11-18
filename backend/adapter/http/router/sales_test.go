package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"pharmacy/service"
	"strings"
	"testing"
)

// fakeSaleService records the customer changes that reach the service.
type fakeSaleService struct {
	service.SaleService

	request types.ChangeSaleCustomerRequest
	userID  int
	calls   int
}

func (f *fakeSaleService) ChangeSaleCustomer(
	_ context.Context, req types.ChangeSaleCustomerRequest, userID int,
) (*types.SaleCustomer, error) {
	f.request, f.userID = req, userID
	f.calls++
	if req.CustomerID == 0 {
		return nil, nil
	}
	return &types.SaleCustomer{ID: req.CustomerID, Name: "Ada Obi"}, nil
}

const changeBody = `{"customer_id":5,"expected_customer_id":2,"reason":"other","note":"typo"}`

func TestChangeSaleCustomerRouteNeedsToManageAndSeeCustomers(t *testing.T) {
	for name, key := range map[string]string{
		"only view": constant.ViewCustomersPermissionKey, "only manage": constant.ManageCustomersPermissionKey,
	} {
		svc := &fakeSaleService{}
		req := withSession(t,
			httptest.NewRequest(http.MethodPut, "/sales/9/customer", strings.NewReader(changeBody)),
			key)

		rec := serve(InitSalesRouter(svc, errorPage), req)

		if rec.Code != http.StatusForbidden || svc.calls != 0 {
			t.Fatalf("%s: want 403 and no call, got %d with %d calls", name, rec.Code, svc.calls)
		}
	}
}

func TestChangeSaleCustomerRoutePassesTheRequestOn(t *testing.T) {
	svc := &fakeSaleService{}
	req := withSession(t,
		httptest.NewRequest(http.MethodPut, "/sales/9/customer",
			strings.NewReader(`{"sale_id":77,`+changeBody[1:])),
		constant.ManageCustomersPermissionKey, constant.ViewCustomersPermissionKey)

	rec := serve(InitSalesRouter(svc, errorPage), req)

	got := svc.request
	if rec.Code != http.StatusOK || got.SaleID != 9 || got.CustomerID != 5 ||
		got.ExpectedCustomerID != 2 || got.Reason != "other" || got.Note != "typo" ||
		svc.userID != 7 {
		t.Fatalf("got %d, %+v from user %d", rec.Code, got, svc.userID)
	}
	if !strings.Contains(rec.Body.String(), `"customer":{"id":5,"name":"Ada Obi"}`) {
		t.Fatalf("the reply names the new customer, got %s", rec.Body.String())
	}
}

func TestChangeSaleCustomerRouteAnswersAWalkInWithNull(t *testing.T) {
	req := withSession(t,
		httptest.NewRequest(
			http.MethodPut,
			"/sales/9/customer",
			strings.NewReader(
				`{"customer_id":0,"expected_customer_id":2,"reason":"customer_asked"}`,
			),
		),
		constant.ManageCustomersPermissionKey, constant.ViewCustomersPermissionKey)

	rec := serve(InitSalesRouter(&fakeSaleService{}, errorPage), req)

	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"customer":null`) {
		t.Fatalf("want 200 with a null customer, got %d %s", rec.Code, rec.Body.String())
	}
}

func TestChangeSaleCustomerRouteRefusesABadSaleID(t *testing.T) {
	req := withSession(t,
		httptest.NewRequest(http.MethodPut, "/sales/abc/customer", strings.NewReader(changeBody)),
		constant.ManageCustomersPermissionKey, constant.ViewCustomersPermissionKey)
	if rec := serve(
		InitSalesRouter(&fakeSaleService{}, errorPage),
		req,
	); rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}
