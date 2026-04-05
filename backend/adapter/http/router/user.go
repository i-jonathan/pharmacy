package router

import (
	"html/template"
	"net/http"
	"pharmacy/adapter/http/controller"
	"pharmacy/adapter/http/middleware"
	"pharmacy/httperror"
	"pharmacy/service"
)

func InitUserRouter(svc service.UserService, tmpl *template.Template) http.Handler {
	userController := controller.NewUserController(svc, tmpl)
	userMux := http.NewServeMux()

	userMux.HandleFunc(http.MethodPost+" /register", userController.CreateUserAccount)
	userMux.HandleFunc(http.MethodGet+" /login", userController.GetLoginPage)
	userMux.HandleFunc(http.MethodPost+" /login", userController.HandleLogin)
	userMux.Handle(http.MethodGet+" /register", middleware.AuthMiddleware(http.HandlerFunc(userController.GetRegisterPage)))
	userMux.Handle("/logout", middleware.AuthMiddleware(http.HandlerFunc(userController.LogoutHandler)))

	userMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		httperror.NotFound("User Page Not Found", nil).Render(w, tmpl)
	})

	return http.StripPrefix("/user", userMux)
}
