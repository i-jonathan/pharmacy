package router

import (
	"html/template"
	"net/http"
	"pharmacy/adapter/http/controller"
	"pharmacy/adapter/http/middleware"
	"pharmacy/httperror"
	"pharmacy/internal/constant"
	"pharmacy/service"
)

// InitHealthCheckRouter serves blood pressure and glucose readings. Recording one needs
// healthcheck:record and customer:view (to find the customer); seeing a customer's readings and
// reprinting one need healthcheck:view; correcting one needs healthcheck:correct as well. The
// service checks again.
func InitHealthCheckRouter(svc service.HealthCheckService, tmpl *template.Template) http.Handler {
	checks := controller.NewHealthCheckController(svc)
	mux := http.NewServeMux()

	mayRecord := middleware.RequirePermissions(constant.RequireAllPermissions,
		constant.RecordHealthCheckPermissionKey, constant.ViewCustomersPermissionKey)
	mayView := middleware.RequirePermissions(constant.RequireAllPermissions,
		constant.ViewHealthChecksPermissionKey)
	mayCorrect := middleware.RequirePermissions(constant.RequireAllPermissions,
		constant.CorrectHealthCheckPermissionKey, constant.ViewHealthChecksPermissionKey)

	mux.Handle(http.MethodPost+" /{$}", mayRecord(http.HandlerFunc(checks.Record)))
	mux.Handle(
		http.MethodGet+" /customer/{id}",
		mayView(http.HandlerFunc(checks.CustomerHistory)),
	)
	mux.Handle(http.MethodPut+" /{id}", mayCorrect(http.HandlerFunc(checks.Correct)))
	mux.Handle(http.MethodPost+" /{id}/reprint", mayView(http.HandlerFunc(checks.Reprint)))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		httperror.NotFound("Health Check Page Not Found", nil).Render(w, tmpl)
	})
	return http.StripPrefix("/health-checks", mux)
}
