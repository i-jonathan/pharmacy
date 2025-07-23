package router

import (
	"html/template"
	"net/http"
	"pharmacy/adapter/http/controller"
	"pharmacy/service"
)

func InitUserRouter(svc service.UserService, tmpl *template.Template) http.Handler {
	userController := controller.NewUserController(svc, tmpl)
	userMux := http.NewServeMux()
	
	userMux.HandleFunc(http.MethodPost + " /register", userController.CreateUserAccount)
	userMux.HandleFunc(http.MethodGet + " /login", userController.GetLoginPage)
	userMux.HandleFunc(http.MethodPost + " /login", userController.HandleLogin)
	
	return http.StripPrefix("/user", userMux)
}