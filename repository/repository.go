package repository

import (
	"context"
	"pharmacy/internal/types"
	"pharmacy/model"

	"github.com/jmoiron/sqlx"
)

type UserRepository interface {
	FetchUserWithPassword(ctx context.Context, userName string) (model.User, error)
	CheckUserNameExists(ctx context.Context, userName string) (bool, error)
	CreateUserAccount(ctx context.Context, user model.User) error
	BulkFetchUserByIDTx(ctx context.Context, tx *sqlx.Tx, userIDs []int) ([]model.User, error)
}

type InventoryRepository interface {
	CreateProductTx(ctx context.Context, tx *sqlx.Tx, product model.Product) (int, error)
	CreateProductPriceTx(ctx context.Context, tx *sqlx.Tx, productPrice model.ProductPrice) (int, error)
	UpdateProductDefaultPriceTx(ctx context.Context, tx *sqlx.Tx, priceID int, productID int) error
	FetchProductCategories(ctx context.Context) ([]model.Category, error)
	SearchProductByName(ctx context.Context, searchTerm string) ([]model.Product, error)
	SearchSuppliersName(ctx context.Context, query string) ([]string, error)
	FetchDefaultPriceID(ctx context.Context, productID int) (int, error)
	CreateReceivingBatchTx(ctx context.Context, tx *sqlx.Tx, receivingBatch model.ReceivingBatch) (int, error)
	BulkCreateProductBatchTx(ctx context.Context, tx *sqlx.Tx, productBatches []model.ProductBatch) ([]types.BatchInsertReturn, error)
	BulkCreateStockMovementTx(ctx context.Context, tx *sqlx.Tx, stockMovements []model.StockMovement) error
	BulkUpdateProductPricesTx(ctx context.Context, tx *sqlx.Tx, updateValues []map[string]any) error
	BulkFetchProductByIDsTx(ctx context.Context, tx *sqlx.Tx, productIDs []int) ([]model.Product, error)
	FetchInventoryItems(ctx context.Context) ([]model.InventoryItem, error)
	FetchPriceByID(ctx context.Context, priceID int) (model.ProductPrice, error)
}

type SaleRepository interface {
	CreateSaleTx(ctx context.Context, tx *sqlx.Tx, sale model.Sale) (int, error)
	BulkCreateSaleItemsTx(ctx context.Context, tx *sqlx.Tx, saleItems []model.SaleItem) error
	BulkCreateSalePaymentsTX(ctx context.Context, tx *sqlx.Tx, salePayments []model.SalePayment) error
	FetchSalesTx(ctx context.Context, tx *sqlx.Tx, filter types.SaleFilter) ([]model.Sale, error)
	BulkFetchSaleItems(ctx context.Context, tx *sqlx.Tx, saleIDs []int) ([]model.SaleItem, error)
	BulkFetchSalePayments(ctx context.Context, tx *sqlx.Tx, saleIDs []int) ([]model.SalePayment, error)
}

type PharmacyRepository interface {
	BeginTx(ctx context.Context) (*sqlx.Tx, error)
	CommitTx(tx *sqlx.Tx) error
	Rollback(tx *sqlx.Tx)
	UserRepository
	InventoryRepository
	SaleRepository
}
