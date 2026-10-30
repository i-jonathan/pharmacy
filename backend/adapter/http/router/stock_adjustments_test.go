package router

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"pharmacy/config"
	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"pharmacy/service"
	"strings"
	"testing"
)

// fakeStockService records the stock adjustment calls; the rest of the inventory service is
// unused here.
type fakeStockService struct {
	service.InventoryService
	adjustment types.StockAdjustmentRequest
	flag       types.StockFlagRequest
	userID     int
	page       int
	dismissed  int
	note       string
}

func (f *fakeStockService) CreateStockAdjustment(
	_ context.Context, req types.StockAdjustmentRequest, userID int,
) (int, error) {
	f.adjustment, f.userID = req, userID
	return 70, nil
}

func (f *fakeStockService) ListStockAdjustments(
	_ context.Context, page int,
) (types.StockAdjustmentPage, error) {
	f.page = page
	return types.StockAdjustmentPage{Page: page, Data: []types.StockAdjustment{}}, nil
}

func (f *fakeStockService) FlagStock(
	_ context.Context, req types.StockFlagRequest, userID int,
) (int, error) {
	f.flag, f.userID = req, userID
	return 90, nil
}

func (f *fakeStockService) CountStockFlags(context.Context) (int, error) { return 3, nil }

func (f *fakeStockService) DismissStockFlag(
	_ context.Context, id int, note string, userID int,
) error {
	f.dismissed, f.note, f.userID = id, note, userID
	return nil
}

// withSession signs a session cookie carrying the permissions, as the login handler does, and
// puts the user id in the request context, as the auth middleware does.
func withSession(t *testing.T, req *http.Request, perms ...string) *http.Request {
	t.Helper()
	granted := map[string]bool{}
	for _, p := range perms {
		granted[p] = true
	}
	body, _ := json.Marshal(granted)
	store := config.NewSessionStore()
	rec := httptest.NewRecorder()
	session, _ := store.New(req, "session")
	session.Values["permissions"] = string(body)
	if err := session.Save(req, rec); err != nil {
		t.Fatal(err)
	}
	for _, c := range rec.Result().Cookies() {
		req.AddCookie(c)
	}
	return req.WithContext(context.WithValue(req.Context(), constant.UserIDKey, 7))
}

func serve(h http.Handler, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestStockAdjustmentRoutesNeedTheAdjustPermission(t *testing.T) {
	h := InitInventoryRouter(&fakeStockService{}, errorPage)
	for _, r := range []struct{ method, path string }{
		{http.MethodPost, "/inventory/stock-adjustments"},
		{http.MethodGet, "/inventory/stock-adjustments"},
		{http.MethodGet, "/inventory/stock-flags"},
		{http.MethodGet, "/inventory/stock-flags/count"},
		{http.MethodPost, "/inventory/stock-flags/5/dismiss"},
	} {
		req := withSession(t, httptest.NewRequest(r.method, r.path, strings.NewReader("{}")),
			constant.FlagStockPermissionKey)
		if rec := serve(h, req); rec.Code != http.StatusForbidden {
			t.Errorf("%s %s with only stock:flag: want 403, got %d", r.method, r.path, rec.Code)
		}
	}
}

func TestCreateStockAdjustmentRoute(t *testing.T) {
	svc := &fakeStockService{}
	body := `{"reason":"damaged","items":[{"product_id":1,"mode":"remove","quantity":2}],"flag_ids":[5]}`
	req := withSession(
		t,
		httptest.NewRequest(
			http.MethodPost,
			"/inventory/stock-adjustments",
			strings.NewReader(body),
		),
		constant.AdjustStockPermissionKey,
	)

	rec := serve(InitInventoryRouter(svc, errorPage), req)

	if rec.Code != http.StatusCreated || !strings.Contains(rec.Body.String(), `"id":70`) {
		t.Fatalf("expected 201 {id:70}, got %d %s", rec.Code, rec.Body.String())
	}
	got := svc.adjustment
	if got.Reason != "damaged" || len(got.Items) != 1 || got.Items[0].Mode != "remove" ||
		got.Items[0].Quantity != 2 || len(got.FlagIDs) != 1 || svc.userID != 7 {
		t.Fatalf("the service got %+v from user %d", got, svc.userID)
	}
}

func TestListStockAdjustmentsRoutePassesThePage(t *testing.T) {
	svc := &fakeStockService{}
	req := withSession(t,
		httptest.NewRequest(http.MethodGet, "/inventory/stock-adjustments?page=2", nil),
		constant.AdjustStockPermissionKey)

	if rec := serve(
		InitInventoryRouter(svc, errorPage),
		req,
	); rec.Code != http.StatusOK ||
		svc.page != 2 {
		t.Fatalf("expected page 2, got %d (status %d)", svc.page, rec.Code)
	}
}

func TestFlagStockRouteIsOpenToFlaggers(t *testing.T) {
	svc := &fakeStockService{}
	body := `{"product_id":1,"note":"two packs crushed","counted_quantity":8}`
	req := withSession(t,
		httptest.NewRequest(http.MethodPost, "/inventory/stock-flags", strings.NewReader(body)),
		constant.FlagStockPermissionKey)

	rec := serve(InitInventoryRouter(svc, errorPage), req)

	if rec.Code != http.StatusCreated || svc.flag.Note != "two packs crushed" ||
		svc.flag.CountedQuantity == nil || *svc.flag.CountedQuantity != 8 {
		t.Fatalf("expected the flag to reach the service, got %d %+v", rec.Code, svc.flag)
	}

	bare := withSession(t,
		httptest.NewRequest(http.MethodPost, "/inventory/stock-flags", strings.NewReader(body)))
	if rec := serve(InitInventoryRouter(svc, errorPage), bare); rec.Code != http.StatusForbidden {
		t.Fatalf("without stock:flag or stock:adjust: want 403, got %d", rec.Code)
	}
}

func TestStockFlagCountAndDismissRoutes(t *testing.T) {
	svc := &fakeStockService{}
	h := InitInventoryRouter(svc, errorPage)

	count := withSession(
		t,
		httptest.NewRequest(http.MethodGet, "/inventory/stock-flags/count", nil),
		constant.AdjustStockPermissionKey,
	)
	if rec := serve(h, count); !strings.Contains(rec.Body.String(), `"open":3`) {
		t.Fatalf("expected {open:3}, got %d %s", rec.Code, rec.Body.String())
	}

	dismiss := withSession(t, httptest.NewRequest(http.MethodPost,
		"/inventory/stock-flags/5/dismiss", strings.NewReader(`{"note":"counted, fine"}`)),
		constant.AdjustStockPermissionKey)
	if rec := serve(h, dismiss); rec.Code != http.StatusNoContent || svc.dismissed != 5 ||
		svc.note != "counted, fine" {
		t.Fatalf("expected flag 5 dismissed, got %d (%d %q)", rec.Code, svc.dismissed, svc.note)
	}
}

// RequirePermissions in "any" mode used to let a request with none of the permissions through to
// the handler; the purchase order draft routes use that mode too.
func TestPurchaseOrderDraftRoutesRefuseUsersWithNeitherPermission(t *testing.T) {
	h := InitInventoryRouter(&fakeStockService{}, errorPage)
	for _, path := range []string{"/inventory/purchase-orders", "/inventory/purchase-orders/3/submit"} {
		req := withSession(t, httptest.NewRequest(http.MethodPost, path, strings.NewReader("{}")),
			constant.FlagStockPermissionKey)
		if rec := serve(h, req); rec.Code != http.StatusForbidden {
			t.Errorf("POST %s without draft or approve: want 403, got %d", path, rec.Code)
		}
	}
}
