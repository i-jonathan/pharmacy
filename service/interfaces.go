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
	FetchCategories(ctx context.Context) ([]types.CategoriesResponse, error)
	SearchProducts(ctx context.Context, query string) ([]types.ProductResult, error)
	SearchForSuppliers(ctx context.Context, query string) ([]string, error)
	ReceiveProductSupply(ctx context.Context, params types.ReceiveSupplyRequest) error
	FetchInventory(ctx context.Context) (*model.Inventory, error)
}

type SaleService interface {
	CreateSale(ctx context.Context, saleParams types.Sale) error
	FetchSalesHistory(ctx context.Context) ([]types.SaleResponse, error)
}
