package router

import (
	"html/template"
	"net/http"
	"pharmacy/adapter/http/controller"
)

func InitAppRouter(tmpl *template.Template) http.Handler {
	appController := controller.NewAppController(tmpl)
	appMux := http.NewServeMux()

	// Old UI dashboard (also handles ?ui=v2 redirect)
	appMux.HandleFunc("GET /dashboard", appController.GetDashboard)

	// V2 SPA shell: serve next-dashboard.html for /app/ and all sub-paths
	appMux.HandleFunc("GET /{$}", appController.ServeV2)
	appMux.HandleFunc("GET /{path...}", appController.ServeV2)

	return http.StripPrefix("/app", appMux)
}
