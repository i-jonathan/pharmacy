package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"pharmacy/config"
	"pharmacy/httperror"
	"pharmacy/internal/constant"
	"pharmacy/model"
	"pharmacy/service"
	"strings"

	"github.com/gorilla/csrf"
)

type userController struct {
	service  service.UserService
	template *template.Template
}

func NewUserController(svc service.UserService, tmpl *template.Template) *userController {
	return &userController{service: svc, template: tmpl}
}

func (c *userController) CreateUserAccount(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}

	u := model.User{
		UserName: r.FormValue("username"),
		Password: r.FormValue("password"),
	}

	err := c.service.CreateUserAccount(r.Context(), u)
	if err != nil {
		var httperr *httperror.HTTPError
		if errors.As(err, &httperr) {
			http.Error(w, httperr.Error(), httperr.Code)
			return
		}
		http.Error(w, "failed to create account", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/app/", http.StatusSeeOther)
}

func (c *userController) GetLoginPage(w http.ResponseWriter, r *http.Request) {
	store := config.NewSessionStore()
	session, err := store.Get(r, "session")
	if err != nil {
		session, _ = store.New(r, "session")
	}

	if userID, ok := session.Values[constant.UserSessionKey]; ok && userID != nil {
		http.Redirect(w, r, "/app/", http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = c.template.ExecuteTemplate(w, "login-v2.html", map[string]any{
		"CSRFToken":  csrf.Token(r),
		"LoginError": r.URL.Query().Has("error"),
	})
	if err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
	}
}

func (c *userController) HandleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var u model.User
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16*1024)).Decode(&u); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid login request."})
		return
	}

	// actually authenticate user
	err := c.service.AuthenticateUser(r.Context(), &u)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid username or password."})
		return
	}

	store := config.NewSessionStore()

	session, err := store.Get(r, "session")
	if err != nil {
		session, _ = store.New(r, "session")
	}

	permMap := map[string]bool{}
	for _, p := range u.Permissions {
		key := strings.ToLower(fmt.Sprintf("%s:%s", p.Resource, p.Action))
		permMap[key] = true
	}

	session.Values[constant.UserSessionKey] = u.ID
	session.Values[constant.UserNameSessionKey] = u.UserName
	session.Values[constant.RoleSessionKey] = u.RoleID
	session.Values[constant.RoleNameSessionKey] = u.RoleName

	permJSON, _ := json.Marshal(permMap)
	session.Values[constant.PermissionsSessionKey] = string(permJSON)

	nextURL, _ := session.Values["next"].(string)
	delete(session.Values, "next")
	if nextURL == "" || !strings.HasPrefix(nextURL, "/") || strings.HasPrefix(nextURL, "//") {
		nextURL = "/app/"
	}
	if err := session.Save(r, w); err != nil {
		http.Error(w, `{"error":"Unable to start your session."}`, http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]string{"redirect": nextURL})
}

func (c *userController) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	store := config.NewSessionStore()
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	// Remove the user session key
	delete(session.Values, constant.UserSessionKey)

	// Expire session immediately
	session.Options.MaxAge = -1

	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to log out", http.StatusInternalServerError)
		return
	}

	// Redirect to login
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}
