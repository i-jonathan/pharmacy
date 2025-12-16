package router

import (
	"html/template"
	"net/http"
	"pharmacy/adapter/http/controller"
	"pharmacy/adapter/http/middleware"
	"pharmacy/service"
)

func InitSalesRouter(svc service.SaleService, tmpl *template.Template) http.Handler {
	saleController := controller.NewSaleController(svc, tmpl)
	saleMux := http.NewServeMux()

	saleMux.HandleFunc(http.MethodPost+" /", saleController.CreateSale)
	saleMux.HandleFunc(http.MethodGet+" /", saleController.RenderSalesReceipt)
	saleMux.Handle(http.MethodGet+" /history", middleware.AddPermissionsToContext(
		http.HandlerFunc(saleController.RenderSalesHistory),
	))
	saleMux.Handle(http.MethodGet+" /filter", middleware.AddPermissionsToContext(
		http.HandlerFunc(saleController.FilterSales),
	))
	saleMux.HandleFunc(http.MethodPost+" /hold", saleController.HoldSaleTransaction)
	saleMux.HandleFunc(http.MethodGet+" /held", saleController.RenderHeldSaleReceipts)
	saleMux.HandleFunc(http.MethodDelete+" /held/{reference}", saleController.DeleteHeldSale)
	saleMux.HandleFunc(http.MethodPost+" /returns", saleController.ReturnItems)
	return http.StripPrefix("/sales", saleMux)
}
