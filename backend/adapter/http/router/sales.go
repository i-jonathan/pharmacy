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

func InitSalesRouter(svc service.SaleService, tmpl *template.Template) http.Handler {
	saleController := controller.NewSaleController(svc)
	saleMux := http.NewServeMux()

	saleMux.HandleFunc(http.MethodPost+" /", saleController.CreateSale)
	saleMux.Handle(http.MethodGet+" /filter", middleware.AddPermissionsToContext(
		http.HandlerFunc(saleController.FilterSales),
	))
	saleMux.HandleFunc(http.MethodPost+" /{id}/reprint", saleController.ReprintReceipt)
	// Changing the customer on a saved sale needs customer:manage and customer:view (the customer
	// is never shown without the second); the service adds the rule for an earlier day's sale.
	saleMux.Handle(http.MethodPut+" /{id}/customer", middleware.RequirePermissions(
		constant.RequireAllPermissions,
		constant.ManageCustomersPermissionKey, constant.ViewCustomersPermissionKey,
	)(http.HandlerFunc(saleController.ChangeSaleCustomer)))
	saleMux.HandleFunc(http.MethodPost+" /hold", saleController.HoldSaleTransaction)
	saleMux.HandleFunc(http.MethodDelete+" /held/{reference}", saleController.DeleteHeldSale)
	saleMux.HandleFunc(http.MethodPost+" /returns", saleController.ReturnItems)
	saleMux.HandleFunc(http.MethodGet+" /api/held", saleController.FetchHeldTransactionsJSON)

	saleMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		httperror.NotFound("Sales Page Not Found", nil).Render(w, tmpl)
	})
	return http.StripPrefix("/sales", saleMux)
}
