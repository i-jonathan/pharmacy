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
	got     types.CreateUserRequest
	actorID int
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
