package router

import (
	"html/template"
	"net/http"
	"pharmacy/adapter/http/controller"
)

func InitAppRouter(tmpl *template.Template) http.Handler {
	appController := controller.NewAppController(tmpl)
	appMux := http.NewServeMux()

	// Serve the Vue SPA shell for /app/ and its routes.
	appMux.HandleFunc("GET /{$}", appController.ServeV2)
	appMux.HandleFunc("GET /{path...}", appController.ServeV2)

	return http.StripPrefix("/app", appMux)
}
