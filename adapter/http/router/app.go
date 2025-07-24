package router

import (
	"html/template"
	"net/http"
	"pharmacy/adapter/http/controller"
)

func InitAppRouter(tmpl *template.Template) http.Handler {
	appController := controller.NewAppController(tmpl)
	appMux := http.NewServeMux()
	
	appMux.HandleFunc(http.MethodGet + " /dashboard", appController.GetDashboard)
	return http.StripPrefix("/app", appMux)
}
