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
}