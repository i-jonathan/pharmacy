package router

import (
	"html/template"
	"net/http"
	"pharmacy/adapter/http/controller"
	"pharmacy/adapter/http/middleware"
	"pharmacy/internal/constant"
	"pharmacy/service"
)

func InitStockTakingRouter(svc service.StockTakingService, tmpl *template.Template) http.Handler {
	stockTakingController := controller.NewStockTakingController(svc, tmpl)
	stockTakingMux := http.NewServeMux()

	stockTakingMux.HandleFunc(http.MethodGet+" /{$}", stockTakingController.RenderStockTakingDashboard)
	stockTakingMux.HandleFunc(http.MethodGet+" /api/list", stockTakingController.ListStockTakings)
	stockTakingMux.HandleFunc(http.MethodPost+" /api/create", stockTakingController.CreateStockTaking)
	stockTakingMux.HandleFunc(http.MethodGet+" /{id}", stockTakingController.RenderStockTakingPage)
	stockTakingMux.Handle(http.MethodGet+" /api/{id}", middleware.AddPermissionsToContext(
		http.HandlerFunc(stockTakingController.FetchStockTaking),
	))
	stockTakingMux.Handle(
		http.MethodPost+" /api/{id}", middleware.RequirePermissions(
			constant.RequireAllPermissions, constant.CompleteStockTakingPermissionKey,
		)(
			http.HandlerFunc(stockTakingController.CompleteStockTaking),
		),
	)
	stockTakingMux.HandleFunc(http.MethodGet+" /api/{id}/items", stockTakingController.FetchStockTakingItems)
	stockTakingMux.HandleFunc(http.MethodPost+" /api/{id}/item/{product_id}", stockTakingController.UpdateStockTakingItemCount)

	return http.StripPrefix("/stock-taking", stockTakingMux)
}
