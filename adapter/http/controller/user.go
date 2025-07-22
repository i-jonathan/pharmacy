package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"pharmacy/httperror"
	"pharmacy/model"
	"pharmacy/service"
)

type userController struct {
	service service.UserService
}

func NewUserController(svc service.UserService) *userController {
	return &userController{service: svc}
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
