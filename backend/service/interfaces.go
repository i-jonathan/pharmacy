package service

import (
	"context"
	"pharmacy/internal/types"
	"pharmacy/model"
)

type UserService interface {
	CreateUserAccount(ctx context.Context, user model.User) error
	AuthenticateUser(ctx context.Context, user *model.User) error
}

type InventoryService interface {
	CreateProduct(ctx context.Context, params types.CreateProductRequest) (types.AddItemResponse, error)
	UpdateProduct(ctx context.Context, id int, params types.UpdateProductRequest) error
	FetchCategories(ctx context.Context) ([]types.CategoriesResponse, error)
	SearchProducts(ctx context.Context, query string) ([]types.ProductResult, error)
	SearchForSuppliers(ctx context.Context, query string) ([]string, error)
	ReceiveProductSupply(ctx context.Context, params types.ReceiveSupplyRequest) error
	HoldReceivingItems(ctx context.Context, holdRequest types.HoldTransactionRequest) error
	FetchHeldReceivingItems(ctx context.Context) ([]types.HeldTransactionResponse, error)
	DeleteHeldTransaction(ctx context.Context, reference string) error
	FetchInventory(ctx context.Context) (*model.Inventory, error)
	FetchProductByID(ctx context.Context, id int) (types.ProductResult, error)
	FetchReceivingBatches(ctx context.Context, filter types.SaleFilter) ([]types.ReceivedBatch, error)
}

type SaleService interface {
	CreateSale(ctx context.Context, saleParams types.Sale) error
	FetchSalesHistory(ctx context.Context, filter types.SaleFilter) (types.SaleHistory, error)
	HoldSale(ctx context.Context, holdSaleRequest types.HoldTransactionRequest) error
	FetchHeldSaleTransactions(ctx context.Context) ([]types.HeldTransactionResponse, error)
	DeleteHeldTransaction(ctx context.Context, reference string) error
	ReturnItems(ctx context.Context, returnParams types.ReturnSale) error
}

type StockTakingService interface {
	CreateStockTaking(ctx context.Context, data types.StockTakingData) (int, error)
	FetchStockTaking(ctx context.Context, stockTakingID int) (types.StockTakingData, error)
	FetchStockTakingItems(ctx context.Context, stockTakingID int) (types.StockTakingItems, error)
	UpdateStockTakingItemCount(ctx context.Context, data types.StockTakingItemCount) error
	CompleteStockTaking(ctx context.Context, stockTakingID, userID int) error
	ListAllStockTakings(ctx context.Context) ([]types.StockTakingListItem, error)
}

type DashboardService interface {
	GetDashboardData(ctx context.Context) (*types.DashboardResponse, error)
}

type AdminService interface {
	GetModules(ctx context.Context) []types.AdminModuleInfo
	GetPermissions(ctx context.Context) ([]types.PermissionInfo, error)
	CreatePermission(ctx context.Context, resource, action string) error
	DeletePermission(ctx context.Context, id int) error
	GetRolePermissions(ctx context.Context, roleID int) ([]types.PermissionInfo, error)
	AssignPermissionToRole(ctx context.Context, roleID, permissionID int) error
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID int) error
	GetRoles(ctx context.Context) ([]types.RoleWithPermissions, error)
	GetUsers(ctx context.Context) ([]types.UserListItem, error)
	UpdateUserRole(ctx context.Context, userID, roleID int) error
	ResetUserPassword(ctx context.Context, userID int, newPassword string) error
	GetCategories(ctx context.Context) ([]types.CategoryListItem, error)
	CreateCategory(ctx context.Context, name string) error
	UpdateCategory(ctx context.Context, id int, name string) error
	DeleteCategory(ctx context.Context, id int) error
}
