package repository

import (
	"context"
	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"pharmacy/model"
	"time"

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
	UpdateProductTx(ctx context.Context, tx *sqlx.Tx, product model.Product) error

	CreateProductPriceTx(ctx context.Context, tx *sqlx.Tx, productPrice model.ProductPrice) (int, error)
	UpdateProductPriceTx(ctx context.Context, tx *sqlx.Tx, price model.ProductPrice) error
	UpdateProductPriceByIDTx(ctx context.Context, tx *sqlx.Tx, priceID int, price model.ProductPrice) error
	DeleteProductPriceTx(ctx context.Context, tx *sqlx.Tx, priceID int) error

	UpdateProductDefaultPriceTx(ctx context.Context, tx *sqlx.Tx, priceID int, productID int) error
	FetchProductCategories(ctx context.Context) ([]model.Category, error)
	SearchProductByName(ctx context.Context, searchTerm string) ([]model.Product, error)
	SearchSuppliersName(ctx context.Context, query string) ([]string, error)
	FetchDefaultPriceID(ctx context.Context, productID int) (int, error)
	CreateReceivingBatchTx(ctx context.Context, tx *sqlx.Tx, receivingBatch model.ReceivingBatch) (int, error)
	CreateStockMovementTx(ctx context.Context, tx *sqlx.Tx, stockMovement model.StockMovement) error
	BulkCreateProductBatchTx(ctx context.Context, tx *sqlx.Tx, productBatches []model.ProductBatch) ([]types.BatchInsertReturn, error)
	BulkCreateStockMovementTx(ctx context.Context, tx *sqlx.Tx, stockMovements []model.StockMovement) error
	BulkUpdateProductPricesTx(ctx context.Context, tx *sqlx.Tx, updateValues []map[string]any) error
	BulkFetchProductByIDsTx(ctx context.Context, tx *sqlx.Tx, productIDs []int) ([]model.Product, error)
	FetchInventoryItems(ctx context.Context) ([]model.InventoryItem, error)
	FetchPriceByID(ctx context.Context, priceID int) (model.ProductPrice, error)
	FetchProductByIDWithPrices(ctx context.Context, id int) (model.Product, error)
}

type SaleRepository interface {
	CreateSaleTx(ctx context.Context, tx *sqlx.Tx, sale model.Sale) (int, error)
	BulkCreateSaleItemsTx(ctx context.Context, tx *sqlx.Tx, saleItems []model.SaleItem) error
	BulkCreateSalePaymentsTX(ctx context.Context, tx *sqlx.Tx, salePayments []model.SalePayment) error
	FetchSalesTx(ctx context.Context, tx *sqlx.Tx, filter types.SaleFilter) ([]model.Sale, error)
	BulkFetchSaleItems(ctx context.Context, tx *sqlx.Tx, saleIDs []int) ([]model.SaleItem, error)
	BulkFetchSalePayments(ctx context.Context, tx *sqlx.Tx, saleIDs []int) ([]model.SalePayment, error)
	SaveHeldTransaction(ctx context.Context, transaction model.HeldTransaction) error
	FetchHeldTransactionsByType(ctx context.Context, transactionType constant.HoldTransactionType) ([]model.HeldTransaction, error)
	DeleteHeldTransactionByReference(ctx context.Context, reference string) error
	DeleteHeldTransactionByReferenceTx(ctx context.Context, tx *sqlx.Tx, reference string) error
	FetchSaleByID(ctx context.Context, saleID int) (model.Sale, error)

	ReturnRepository
}

type ReturnRepository interface {
	CreateReturnTx(ctx context.Context, tx *sqlx.Tx, rtn model.Return) (int, error)
	BulkCreateReturnItemsTx(ctx context.Context, tx *sqlx.Tx, returnItems []model.ReturnItems) error
	FetchAllSaleReturns(ctx context.Context, saleID int) ([]model.ReturnItems, error)
	BulkFetchReturnItemsBySaleIDs(ctx context.Context, saleIDs []int) ([]model.ReturnItemWithSale, error)
}

type StockTakingRepository interface {
	CreateStockTaking(ctx context.Context, stockTakingData model.StockTaking) (int, error)
	GetStockTaking(ctx context.Context, stockTakingID int) (model.StockTaking, error)
	GetStockTakingItems(ctx context.Context, stockTakingID int) (types.StockTakingItems, error)
	HasActiveStockTaking(ctx context.Context) (bool, error)
	FetchCurrentStockLevel(ctx context.Context, productID int) (int, error)
	FetchStockTakingItem(ctx context.Context, stockTakingID, productID int) (*model.StockTakingItem, error)
	CreateStockTakingItem(ctx context.Context, item *model.StockTakingItem) (int, error)
	UpdateStockTakingItem(ctx context.Context, item *model.StockTakingItem) error
	UpdateProductCurrentExpiry(ctx context.Context, productID int, currentExpiry *time.Time) error
	CompleteStockTakingTx(ctx context.Context, tx *sqlx.Tx, stockTakingID, userID int) error
	ListAllStockTakings(ctx context.Context) ([]model.StockTaking, error)
}

type DashboardRepository interface {
	GetTotalSales(ctx context.Context, startDate, endDate time.Time) (int, error)
	GetTransactionCount(ctx context.Context, startDate, endDate time.Time) (int, error)
	GetTotalInventoryItems(ctx context.Context) (int, error)
	GetLowStockCount(ctx context.Context) (int, error)
	GetLowStockItems(ctx context.Context) ([]model.LowStockItem, error)
	GetSalesByCategory(ctx context.Context, startDate, endDate time.Time) ([]model.SalesByCategory, error)
	GetExpiringItems(ctx context.Context, startDate, endDate time.Time) ([]model.ExpiringItem, error)
}

type PharmacyRepository interface {
	BeginTx(ctx context.Context) (*sqlx.Tx, error)
	CommitTx(tx *sqlx.Tx) error
	RollbackTx(tx *sqlx.Tx)
	GetSalesByTime(ctx context.Context, startTime, endTime time.Time) ([]model.Sale, error)
	UserRepository
	InventoryRepository
	SaleRepository
	StockTakingRepository
	DashboardRepository
}
