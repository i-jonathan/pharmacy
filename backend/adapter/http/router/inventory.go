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

func InitInventoryRouter(svc service.InventoryService, tmpl *template.Template) http.Handler {
	inventoryController := controller.NewInventoryController(svc)
	inventoryMux := http.NewServeMux()

	inventoryMux.HandleFunc(http.MethodPost+" /add-item", inventoryController.CreateProduct)
	// Anyone signed in can read what they are allowed to see (the service scopes it). Drafting
	// needs the draft or approve permission; the rest needs approve.
	mayDraft := middleware.RequirePermissions(
		constant.RequireAnyPermissions,
		constant.DraftPurchaseOrderPermissionKey, constant.ApprovePurchaseOrderPermissionKey,
	)
	mayApprove := middleware.RequirePermissions(
		constant.RequireAllPermissions, constant.ApprovePurchaseOrderPermissionKey,
	)
	inventoryMux.HandleFunc(http.MethodGet+" /purchase-orders", inventoryController.ListPurchaseOrders)
	inventoryMux.Handle(http.MethodGet+" /purchase-orders/pending-count",
		mayApprove(http.HandlerFunc(inventoryController.CountPendingPurchaseOrders)))
	inventoryMux.Handle(http.MethodPost+" /purchase-orders",
		mayDraft(http.HandlerFunc(inventoryController.SavePurchaseOrder)))
	inventoryMux.HandleFunc(http.MethodGet+" /purchase-orders/{id}", inventoryController.GetPurchaseOrder)
	inventoryMux.HandleFunc(
		http.MethodGet+" /purchase-orders/{id}/history", inventoryController.GetPurchaseOrderHistory,
	)
	inventoryMux.Handle(http.MethodPost+" /purchase-orders/{id}/submit",
		mayDraft(http.HandlerFunc(inventoryController.SubmitPurchaseOrder)))
	inventoryMux.Handle(http.MethodPost+" /purchase-orders/{id}/approve",
		mayApprove(http.HandlerFunc(inventoryController.ApprovePurchaseOrder)))
	inventoryMux.Handle(http.MethodPost+" /purchase-orders/{id}/reject",
		mayApprove(http.HandlerFunc(inventoryController.RejectPurchaseOrder)))
	inventoryMux.Handle(http.MethodPost+" /purchase-orders/{id}/close",
		mayApprove(http.HandlerFunc(inventoryController.ClosePurchaseOrder)))
	inventoryMux.Handle(http.MethodPost+" /purchase-orders/{id}/move-outstanding",
		mayApprove(http.HandlerFunc(inventoryController.MovePurchaseOrderOutstanding)))
	// Stock adjustments are for managers (stock:adjust); flagging a product for a manager to
	// check is open to anyone who can flag or adjust.
	mayAdjustStock := middleware.RequirePermissions(
		constant.RequireAllPermissions, constant.AdjustStockPermissionKey,
	)
	mayFlagStock := middleware.RequirePermissions(
		constant.RequireAnyPermissions,
		constant.FlagStockPermissionKey, constant.AdjustStockPermissionKey,
	)
	inventoryMux.Handle(http.MethodPost+" /stock-adjustments",
		mayAdjustStock(http.HandlerFunc(inventoryController.CreateStockAdjustment)))
	inventoryMux.Handle(http.MethodGet+" /stock-adjustments",
		mayAdjustStock(http.HandlerFunc(inventoryController.ListStockAdjustments)))
	inventoryMux.Handle(http.MethodPost+" /stock-flags",
		mayFlagStock(http.HandlerFunc(inventoryController.FlagStock)))
	inventoryMux.Handle(http.MethodGet+" /stock-flags",
		mayAdjustStock(http.HandlerFunc(inventoryController.ListStockFlags)))
	inventoryMux.Handle(http.MethodGet+" /stock-flags/count",
		mayAdjustStock(http.HandlerFunc(inventoryController.CountStockFlags)))
	inventoryMux.Handle(http.MethodPost+" /stock-flags/{id}/dismiss",
		mayAdjustStock(http.HandlerFunc(inventoryController.DismissStockFlag)))
	inventoryMux.HandleFunc(http.MethodPost+" /receive-items/hold", inventoryController.HoldReceivingItems)
	inventoryMux.HandleFunc(http.MethodGet+" /receive-items/held/api", inventoryController.FetchHeldReceivingItemsJSON)
	inventoryMux.HandleFunc(http.MethodDelete+" /receive-items/held/{reference}", inventoryController.DeleteHeldReceivingItems)
	inventoryMux.HandleFunc(http.MethodGet+" /search", inventoryController.SearchForProduct)
	inventoryMux.HandleFunc(http.MethodGet+" /suppliers/search", inventoryController.SearchForSuppliers)
	inventoryMux.HandleFunc(http.MethodPost+" /receive-items", inventoryController.ReceiveSupply)
	inventoryMux.HandleFunc(http.MethodGet+" /report/stock", inventoryController.DownloadInventoryReport)
	inventoryMux.HandleFunc(http.MethodGet+" /item-list", inventoryController.FetchInventory)
	inventoryMux.HandleFunc(http.MethodGet+" /top-selling", inventoryController.GetTopSellingProducts)
	inventoryMux.HandleFunc(http.MethodGet+" /product/{id}", inventoryController.GetProductDetails)
	inventoryMux.HandleFunc(http.MethodGet+" /received-items-history/api", inventoryController.FetchReceivedItemsHistory)
	inventoryMux.Handle(http.MethodPut+" /product/{id}",
		middleware.RequirePermissions(constant.RequireAllPermissions, constant.EditInventoryPermissionKey)(
			http.HandlerFunc(inventoryController.UpdateProduct),
		),
	)

	inventoryMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		httperror.NotFound("Inventory Page Not Found", nil).Render(w, tmpl)
	})

	return http.StripPrefix("/inventory", inventoryMux)
}
