package router

import (
	"net/http"
	"pharmacy/adapter/http/controller"
	"pharmacy/service"
)

func InitUserRouter(svc service.UserService) http.Handler {
	userController := controller.NewUserController(svc)
	userMux := http.NewServeMux()
	
	userMux.HandleFunc(http.MethodPost + " /register", userController.CreateUserAccount)
	return http.StripPrefix("/user", userMux)
}