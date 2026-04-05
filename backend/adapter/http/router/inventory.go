package router

import (
	"html/template"
	"net/http"
	"pharmacy/adapter/http/controller"
	"pharmacy/adapter/http/middleware"
	"pharmacy/internal/constant"
	"pharmacy/service"
)

func InitInventoryRouter(svc service.InventoryService, tmpl *template.Template) http.Handler {
	inventoryController := controller.NewInventoryController(svc, tmpl)
	inventoryMux := http.NewServeMux()

	inventoryMux.HandleFunc(http.MethodPost+" /add-item", inventoryController.CreateProduct)
	inventoryMux.HandleFunc(http.MethodGet+" /receive-items", inventoryController.GetReceiveItems)
	inventoryMux.HandleFunc(http.MethodGet+" /search", inventoryController.SearchForProduct)
	inventoryMux.HandleFunc(http.MethodGet+" /suppliers/search", inventoryController.SearchForSuppliers)
	inventoryMux.HandleFunc(http.MethodPost+" /receive-items", inventoryController.ReceiveSupply)
	inventoryMux.HandleFunc(http.MethodGet+" /items", inventoryController.RenderInventoryPage)
	inventoryMux.HandleFunc(http.MethodGet+" /report/stock", inventoryController.DownloadInventoryReport)
	inventoryMux.HandleFunc(http.MethodGet+" /item-list", inventoryController.FetchInventory)
	inventoryMux.HandleFunc(http.MethodGet+" /product/{id}", inventoryController.GetProductDetails)
	inventoryMux.Handle(http.MethodPut+" /product/{id}",
		middleware.RequirePermissions(constant.RequireAllPermissions, constant.EditInventoryPermissionKey)(
			http.HandlerFunc(inventoryController.UpdateProduct),
		),
	)

	return http.StripPrefix("/inventory", inventoryMux)
}
