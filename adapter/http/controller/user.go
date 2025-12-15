package controller

import (
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

	http.Redirect(w, r, "/app/dashboard", http.StatusSeeOther)
}

func (c *userController) GetLoginPage(w http.ResponseWriter, r *http.Request) {
	store := config.NewSessionStore()
	session, err := store.Get(r, "session")
	if err != nil {
		session, _ = store.New(r, "session")
	}

	if userID, ok := session.Values[constant.UserSessionKey]; ok && userID != nil {
		http.Redirect(w, r, "/app/dashboard", http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = c.template.ExecuteTemplate(w, "login.html", map[string]any{
		"CSRFField": csrf.TemplateField(r),
	})
	if err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
	}
}

func (c *userController) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}

	u := model.User{
		UserName: r.FormValue("username"),
		Password: r.FormValue("password"),
	}

	// actually authenticate user
	err := c.service.AuthenticateUser(r.Context(), &u)
	if err != nil {
		http.Redirect(w, r, "/user/login?error="+err.Error(), http.StatusSeeOther)
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
	session.Values[constant.RoleSessionKey] = u.RoleID
	session.Values[constant.PermissionsSessionKey] = permMap
	_ = session.Save(r, w)

	nextURL, _ := session.Values["next"].(string)
	delete(session.Values, "next")
	session.Save(r, w)
	if nextURL == "" {
		nextURL = "/app/dashboard"
	}

	http.Redirect(w, r, nextURL, http.StatusSeeOther)
}

func (c *userController) GetRegisterPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := c.template.ExecuteTemplate(w, "register.html", map[string]any{
		"CSRFField": csrf.TemplateField(r),
	})
	if err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
	}
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
