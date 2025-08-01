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
	CreateProduct(ctx context.Context, params types.CreateProductRequest) error
}