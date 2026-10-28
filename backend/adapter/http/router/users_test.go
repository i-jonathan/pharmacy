package router

import (
	"context"
	"encoding/json"
	"html/template"
	"net/http"
	"net/http/httptest"
	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"pharmacy/service"
	"strings"
	"testing"
)

var errorPage = template.Must(template.New("").Parse(`{{define "error.html"}}{{.Message}}{{end}}`))

func TestPublicRegistrationIsGone(t *testing.T) {
	handler := InitUserRouter(nil, errorPage)
	req := httptest.NewRequest(
		http.MethodPost, "/user/register", strings.NewReader("username=x&password=y"),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("POST /user/register should be gone, got status %d", rec.Code)
	}
}

// fakeAdminService captures the create-user call; everything else is unused here.
type fakeAdminService struct {
	service.AdminService
	got      types.CreateUserRequest
	actorID  int
	resetFor int
	resetTo  string
}

func (f *fakeAdminService) CreateUser(
	_ context.Context, req types.CreateUserRequest, actorID int,
) (int, error) {
	f.got, f.actorID = req, actorID
	return 42, nil
}

func TestAdminCreateUserPassesTheRequestAndActor(t *testing.T) {
	svc := &fakeAdminService{}
	handler := InitAdminRouter(svc, errorPage)
	body := `{"username":"ada","password":"ten chars!","role_id":3}`
	req := httptest.NewRequest(http.MethodPost, "/admin/api/users", strings.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), constant.UserIDKey, 7))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil || resp.ID != 42 {
		t.Fatalf("expected {\"id\":42}, got %s", rec.Body.String())
	}
	want := types.CreateUserRequest{Username: "ada", Password: "ten chars!", RoleID: 3}
	if svc.got != want || svc.actorID != 7 {
		t.Fatalf("service got %+v from actor %d", svc.got, svc.actorID)
	}
}

func TestAdminCreateUserNeedsASignedInUser(t *testing.T) {
	svc := &fakeAdminService{}
	handler := InitAdminRouter(svc, errorPage)
	body := `{"username":"ada","password":"ten chars!","role_id":3}`
	req := httptest.NewRequest(http.MethodPost, "/admin/api/users", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized || svc.got.Username != "" {
		t.Fatalf("expected 401 without calling the service, got %d (service saw %+v)",
			rec.Code, svc.got)
	}
}

func TestAdminCreateUserRefusesAnOversizedBody(t *testing.T) {
	svc := &fakeAdminService{}
	handler := InitAdminRouter(svc, errorPage)
	big := `{"username":"ada","password":"` + strings.Repeat("x", 32*1024) + `","role_id":3}`
	req := httptest.NewRequest(http.MethodPost, "/admin/api/users", strings.NewReader(big))
	req = req.WithContext(context.WithValue(req.Context(), constant.UserIDKey, 7))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge || svc.got.Username != "" {
		t.Fatalf("expected 413 without calling the service, got %d", rec.Code)
	}
}

func (f *fakeAdminService) ResetUserPassword(_ context.Context, userID int, password string) error {
	f.resetFor, f.resetTo = userID, password
	return nil
}

func TestAdminResetPasswordRefusesAnOversizedBody(t *testing.T) {
	svc := &fakeAdminService{}
	handler := InitAdminRouter(svc, errorPage)
	big := `{"new_password":"` + strings.Repeat("x", 32*1024) + `"}`
	req := httptest.NewRequest(
		http.MethodPost, "/admin/api/users/5/reset-password", strings.NewReader(big),
	)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge || svc.resetFor != 0 {
		t.Fatalf("expected 413 without calling the service, got %d", rec.Code)
	}
}
