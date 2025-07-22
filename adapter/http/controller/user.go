package controller

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"pharmacy/httperror"
	"pharmacy/model"
	"pharmacy/service"
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
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := c.template.ExecuteTemplate(w, "login.html", nil)
	if err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
	}
}
