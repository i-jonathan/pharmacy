package repository

import (
	"context"
	"pharmacy/model"

	"github.com/jmoiron/sqlx"
)

type UserRepository interface {
	FetchUserWithPassword(ctx context.Context, userName string) (model.User, error)
	CheckUserNameExists(ctx context.Context, userName string) (bool, error)
	CreateUserAccount(ctx context.Context, user model.User) error
}

type InventoryRepository interface {
	CreateProductTx(ctx context.Context, tx *sqlx.Tx, product model.Product) (int, error)
	CreateProductPriceTx(ctx context.Context, tx *sqlx.Tx, productPrice model.ProductPrice) (int, error)
	UpdateProductDefaultPriceTx(ctx context.Context, tx *sqlx.Tx, priceID int, productID int) error
}

type PharmacyRepository interface {
	BeginTx(ctx context.Context) (*sqlx.Tx, error)
	CommitTx(tx *sqlx.Tx) error
	Rollback(tx *sqlx.Tx)
	UserRepository
	InventoryRepository
}
