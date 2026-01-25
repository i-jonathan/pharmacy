package router

import (
	"html/template"
	"net/http"
	"pharmacy/adapter/http/controller"
	"pharmacy/adapter/http/middleware"
	"pharmacy/service"
)

func InitStockTakingRouter(svc service.StockTakingService, tmpl *template.Template) http.Handler {
	stockTakingController := controller.NewStockTakingController(svc, tmpl)
	stockTakingMux := http.NewServeMux()

	// stockTakingMux.HandleFunc(http.MethodPost+" /", stockTakingController.CreateStockTaking)
	stockTakingMux.HandleFunc(http.MethodGet+" /{stock_taking_id}", stockTakingController.RenderStockTakingPage)
	stockTakingMux.Handle(http.MethodGet+" /api/{stock_taking_id}", middleware.AddPermissionsToContext(
		http.HandlerFunc(stockTakingController.FetchStockTaking),
	))

	return http.StripPrefix("/stock-taking", stockTakingMux)
}
