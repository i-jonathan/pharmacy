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

// InitCustomerRouter serves the customer list. Seeing customers needs customer:view; adding,
// editing, deactivating and reactivating them needs customer:manage.
func InitCustomerRouter(svc service.CustomerService, tmpl *template.Template) http.Handler {
	customers := controller.NewCustomerController(svc)
	mux := http.NewServeMux()

	mayView := middleware.RequirePermissions(
		constant.RequireAllPermissions, constant.ViewCustomersPermissionKey,
	)
	mayManage := middleware.RequirePermissions(
		constant.RequireAllPermissions, constant.ManageCustomersPermissionKey,
	)
	mux.Handle(http.MethodGet+" /{$}", mayView(http.HandlerFunc(customers.ListCustomers)))
	mux.Handle(http.MethodGet+" /{id}", mayView(http.HandlerFunc(customers.GetCustomer)))
	mux.Handle(http.MethodPost+" /{$}", mayManage(http.HandlerFunc(customers.CreateCustomer)))
	mux.Handle(http.MethodPut+" /{id}", mayManage(http.HandlerFunc(customers.UpdateCustomer)))
	mux.Handle(
		http.MethodPost+" /{id}/deactivate",
		mayManage(http.HandlerFunc(customers.Deactivate)),
	)
	mux.Handle(
		http.MethodPost+" /{id}/reactivate",
		mayManage(http.HandlerFunc(customers.Reactivate)),
	)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		httperror.NotFound("Customer Page Not Found", nil).Render(w, tmpl)
	})
	return http.StripPrefix("/customers", mux)
}
