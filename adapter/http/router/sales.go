package router

import (
	"html/template"
	"net/http"
	"pharmacy/adapter/http/controller"
	"pharmacy/service"
)

func InitSalesRouter(svc service.SaleService, tmpl *template.Template) http.Handler {
	saleController := controller.NewSaleController(svc, tmpl)
	saleMux := http.NewServeMux()

	saleMux.HandleFunc(http.MethodPost+" /", saleController.CreateSale)
	saleMux.HandleFunc(http.MethodGet+" /", saleController.RenderSalesReceipt)
	return http.StripPrefix("/sales", saleMux)
}
