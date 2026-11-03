package middleware

import (
	"net/http"
	"net/http/httptest"
	"pharmacy/internal/constant"
	"strings"
	"testing"
)

func TestAuthMiddlewareWithoutASession(t *testing.T) {
	reached := false
	h := AuthMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		reached = true
	}))

	// A page load goes to the sign-in page.
	page := httptest.NewRequest(http.MethodGet, "/app/stock-adjustments", nil)
	page.Header.Set("Accept", "text/html,application/xhtml+xml,*/*;q=0.8")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, page)
	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/user/login" {
		t.Fatalf("a page load should redirect to sign in, got %d %q", rec.Code,
			rec.Header().Get("Location"))
	}

	// A call from the app gets 401 with a message: following the redirect would hand the
	// caller a 200 login page, which it would take for success.
	api := httptest.NewRequest(http.MethodPost, "/inventory/stock-adjustments", nil)
	api.Header.Set("Accept", "*/*")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, api)
	if rec.Code != http.StatusUnauthorized || !strings.Contains(rec.Body.String(), "sign in") {
		t.Fatalf("an API call should get 401 with a message, got %d %s", rec.Code,
			rec.Body.String())
	}
	if strings.Contains(strings.Join(rec.Header().Values("Set-Cookie"), ";"), "session") {
		t.Fatal("an API call must not save itself as the page to return to after sign-in")
	}
	if reached {
		t.Fatal("the handler must not run without a session")
	}
}

// The app shows the server's message, so a refused API call says why in JSON; a page load keeps
// the plain refusal.
func TestRequirePermissionsRefusalSaysWhy(t *testing.T) {
	h := RequirePermissions(constant.RequireAllPermissions, "stock:adjust")(
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			t.Fatal("the handler must not run without the permission")
		}),
	)
	api := httptest.NewRequest(http.MethodPost, "/inventory/stock-adjustments", nil)
	api.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, api)
	if rec.Code != http.StatusForbidden ||
		!strings.Contains(rec.Header().Get("Content-Type"), "application/json") ||
		!strings.Contains(rec.Body.String(), "permission") {
		t.Fatalf("an API call should get a JSON 403 that mentions permission, got %d %q %s",
			rec.Code, rec.Header().Get("Content-Type"), rec.Body.String())
	}
}

// The SPA retries a request once when the CSRF refusal carries code csrf_failed.
func TestCSRFRefusalOfAnAPICallCarriesItsCode(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/inventory/stock-adjustments", nil)
	rec := httptest.NewRecorder()
	csrfErrorHandler(rec, req)
	if rec.Code != http.StatusForbidden ||
		!strings.Contains(rec.Body.String(), `"code":"csrf_failed"`) {
		t.Fatalf(
			"expected a JSON 403 with code csrf_failed, got %d %s",
			rec.Code,
			rec.Body.String(),
		)
	}
}
