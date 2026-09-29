package router

import (
	"html/template"
	"net/http"
	"pharmacy/adapter/http/controller"
)

func InitAppRouter(tmpl *template.Template) http.Handler {
	appController := controller.NewAppController(tmpl)
	appMux := http.NewServeMux()

	// Preserve the old dashboard URL as a redirect to the Vue app.
	appMux.HandleFunc("GET /dashboard", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/app/", http.StatusSeeOther)
	})

	// V2 SPA shell: serve next-dashboard.html for /app/ and all sub-paths
	appMux.HandleFunc("GET /{$}", appController.ServeV2)
	appMux.HandleFunc("GET /{path...}", appController.ServeV2)

	return http.StripPrefix("/app", appMux)
}
