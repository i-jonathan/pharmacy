package router

import (
	"html/template"
	"net/http"
	"pharmacy/adapter/http/controller"
	"pharmacy/httperror"
	"pharmacy/service"
)

func InitAdminRouter(svc service.AdminService, tmpl *template.Template) http.Handler {
	ctrl := controller.NewAdminController(svc, tmpl)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", ctrl.GetAdminDashboard)
	mux.HandleFunc("GET /api/modules", ctrl.ListModules)

	mux.HandleFunc("GET /api/permissions", ctrl.ListPermissions)
	mux.HandleFunc("POST /api/permissions", ctrl.CreatePermission)
	mux.HandleFunc("DELETE /api/permissions/{id}", ctrl.DeletePermission)
	mux.HandleFunc("POST /api/permissions/{id}/role/{roleId}", ctrl.AssignPermissionToRole)
	mux.HandleFunc("DELETE /api/permissions/{id}/role/{roleId}", ctrl.RemovePermissionFromRole)

	mux.HandleFunc("GET /api/roles", ctrl.ListRoles)

	mux.HandleFunc("GET /api/users", ctrl.ListUsers)
	mux.HandleFunc("POST /api/users/{id}/role", ctrl.UpdateUserRole)
	mux.HandleFunc("POST /api/users/{id}/reset-password", ctrl.ResetUserPassword)

	mux.HandleFunc("GET /api/categories", ctrl.ListCategories)
	mux.HandleFunc("POST /api/categories", ctrl.CreateCategory)
	mux.HandleFunc("PUT /api/categories/{id}", ctrl.UpdateCategory)
	mux.HandleFunc("DELETE /api/categories/{id}", ctrl.DeleteCategory)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		httperror.NotFound("Admin Page Not Found", nil).Render(w, tmpl)
	})

	return http.StripPrefix("/admin", mux)
}
