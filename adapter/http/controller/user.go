package controller

import (
	"net/http"
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
	c.service.CreateUserAccount(model.User{})
}
