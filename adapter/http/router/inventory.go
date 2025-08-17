package router

import (
	"html/template"
	"net/http"
	"pharmacy/adapter/http/controller"
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
	inventoryMux.HandleFunc(http.MethodGet+" /sales/receipt", inventoryController.RenderSalesReceipt)

	return http.StripPrefix("/inventory", inventoryMux)
}
