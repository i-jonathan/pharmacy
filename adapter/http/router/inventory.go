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
	
	inventoryMux.HandleFunc(http.MethodPost + " /add-item", inventoryController.CreateProduct)
	inventoryMux.HandleFunc(http.MethodGet + " /receive-items", inventoryController.GetReceiveItems)
	
	return http.StripPrefix("/inventory", inventoryMux)
}