package router

import (
	"html/template"
	"net/http"
	"pharmacy/adapter/http/controller"
	"pharmacy/httperror"
)

func InitAppRouter(tmpl *template.Template) http.Handler {
	appController := controller.NewAppController(tmpl)
	appMux := http.NewServeMux()
	
	appMux.HandleFunc(http.MethodGet + " /dashboard", appController.GetDashboard)
	
	appMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		httperror.NotFound("", nil).Render(w, tmpl)
	})
	
	return http.StripPrefix("/app", appMux)
}
