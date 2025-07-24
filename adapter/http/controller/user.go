package controller

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"pharmacy/config"
	"pharmacy/constant"
	"pharmacy/httperror"
	"pharmacy/model"
	"pharmacy/service"

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
	var u model.User

	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		httperror.BadRequest("invalid json", err).JSONRespond(w)
		return
	}

	err := c.service.CreateUserAccount(r.Context(), u)
	if err != nil {
		var httperr *httperror.HTTPError
		if errors.As(err, &httperr) {
			httperr.JSONRespond(w)
			return
		}
		json.NewEncoder(w).Encode(err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *userController) GetLoginPage(w http.ResponseWriter, r *http.Request) {
	session, err := config.SessionStore.Get(r, "session")
	if err != nil {
		session, _ = config.SessionStore.New(r, "session")
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

	session, err := config.SessionStore.Get(r, "session")
	if err != nil {
		session, _ = config.SessionStore.New(r, "session")
	}

	session.Values[constant.UserSessionKey] = u.ID
	_ = session.Save(r, w)
	http.Redirect(w, r, "/app/dashboard", http.StatusSeeOther)
}
