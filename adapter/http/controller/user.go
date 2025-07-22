package controller

import (
	"encoding/json"
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
	w.WriteHeader(http.StatusNoContent)
}
