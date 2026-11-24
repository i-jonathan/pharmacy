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
	CreateUserTx(ctx context.Context, tx *sqlx.Tx, user model.User) (int, error)
	FetchUsernameByID(ctx context.Context, userID int) (string, error)
	FetchRoleIDByName(ctx context.Context, name string) (int, error)
	FetchUsernamesByRole(ctx context.Context, roleID int) ([]string, error)
	BulkFetchUserByIDTx(ctx context.Context, tx *sqlx.Tx, userIDs []int) ([]model.User, error)
}

type InventoryRepository interface {
	FindOrCreateSupplierTx(ctx context.Context, tx *sqlx.Tx, name string) (types.Supplier, error)
	CreatePurchaseOrderTx(ctx context.Context, tx *sqlx.Tx, order types.NewPurchaseOrder) (int, error)
	GetPurchaseOrderForUpdateTx(ctx context.Context, tx *sqlx.Tx, id int) (types.PurchaseOrder, error)
	UpdatePurchaseOrderHeaderTx(ctx context.Context, tx *sqlx.Tx, id int, header types.NewPurchaseOrder) error
	InsertPurchaseOrderItemTx(ctx context.Context, tx *sqlx.Tx, orderID int, item types.PurchaseOrderItem) error
	UpdatePurchaseOrderItemTx(ctx context.Context, tx *sqlx.Tx, orderID int, item types.PurchaseOrderItem) error
	DeletePurchaseOrderItemTx(ctx context.Context, tx *sqlx.Tx, orderID, itemID int) error
	ReceivePurchaseOrderItemTx(ctx context.Context, tx *sqlx.Tx, orderID, itemID, productID, quantity int) error
	RefreshPurchaseOrderStatusTx(ctx context.Context, tx *sqlx.Tx, id int) error
	ClosePurchaseOrderTx(ctx context.Context, tx *sqlx.Tx, id int) error
	SetPurchaseOrderStatusTx(ctx context.Context, tx *sqlx.Tx, id int, status string) error
	RevertPurchaseOrderToDraftTx(ctx context.Context, tx *sqlx.Tx, id int) error
	ApprovePurchaseOrderTx(ctx context.Context, tx *sqlx.Tx, id, approverID int) error
	CountPendingPurchaseOrders(ctx context.Context) (map[string]int, error)
	ListPurchaseOrders(ctx context.Context, statuses []string, limit int, createdBy *int) ([]types.PurchaseOrder, error)
	GetPurchaseOrder(ctx context.Context, id int) (types.PurchaseOrder, error)
	PurchaseOrderExists(ctx context.Context, id int) (bool, error)
	InsertAuditLogTx(ctx context.Context, tx *sqlx.Tx, entry types.AuditEntry) error
	FetchAuditLog(ctx context.Context, entityType string, entityID int) ([]types.AuditRecord, error)
	// FetchSaleCustomerChanges returns the customer_changed audit entries of the given sales,
	// newest first.
	FetchSaleCustomerChanges(
		ctx context.Context, tx *sqlx.Tx, saleIDs []int,
	) ([]types.SaleCustomerChangeRow, error)
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
	UpdateProductCurrentExpiryAfterReceiving(ctx context.Context, productID int, expiryDate time.Time) error
	FetchInventoryItems(ctx context.Context) ([]model.InventoryItem, error)
	FetchPriceByID(ctx context.Context, priceID int) (model.ProductPrice, error)
	FetchProductByIDWithPrices(ctx context.Context, id int) (model.Product, error)
	FetchReceivingBatches(ctx context.Context, startDate, endDate *time.Time) ([]types.ReceivedBatch, error)
	FetchBatchItemsByReceivingBatchID(ctx context.Context, batchID int) ([]types.ReceivedBatchItem, error)
}

type StockAdjustmentRepository interface {
	ProductsInActiveStockTakingTx(
		ctx context.Context, tx *sqlx.Tx, productIDs []int,
	) ([]int, error)
	LockProductStockTx(
		ctx context.Context, tx *sqlx.Tx, productIDs []int,
	) ([]types.ProductStock, error)
	StockAdjustmentByKeyTx(
		ctx context.Context, tx *sqlx.Tx, key string,
	) (*types.SavedStockAdjustment, error)
	CreateStockAdjustmentTx(
		ctx context.Context, tx *sqlx.Tx, adj types.NewStockAdjustment,
	) (int, bool, error)
	InsertStockAdjustmentItemTx(
		ctx context.Context, tx *sqlx.Tx, adjustmentID int, item types.NewStockAdjustmentItem,
	) error
	SetProductCurrentExpiryTx(
		ctx context.Context, tx *sqlx.Tx, productID int, expiry time.Time,
	) error
	ListStockAdjustments(
		ctx context.Context, limit, offset int,
	) ([]types.StockAdjustment, int, error)
	CreateStockFlagTx(
		ctx context.Context, tx *sqlx.Tx, flag types.StockFlagRequest, userID int,
	) (int, error)
	ListOpenStockFlags(ctx context.Context) ([]types.StockFlag, error)
	CountOpenStockFlags(ctx context.Context) (int, error)
	GetStockFlagsForUpdateTx(ctx context.Context, tx *sqlx.Tx, ids []int) ([]types.StockFlag, error)
	ResolveStockFlagsTx(ctx context.Context, tx *sqlx.Tx, ids []int, adjustmentID, userID int) error
	DismissStockFlagTx(ctx context.Context, tx *sqlx.Tx, id, userID int, note string) error
}

type HealthCheckRepository interface {
	// CreateHealthCheckTx inserts a reading and returns its id and true, or 0 and false when
	// another save already used the key (the caller looks that one up).
	CreateHealthCheckTx(
		ctx context.Context, tx *sqlx.Tx, check types.NewHealthCheck,
	) (int, bool, error)
	// HealthCheckByKeyTx returns the reading saved under a retry key, or nil.
	HealthCheckByKeyTx(
		ctx context.Context, tx *sqlx.Tx, key string,
	) (*types.SavedHealthCheck, error)
	GetHealthCheckTx(ctx context.Context, tx *sqlx.Tx, id int) (types.HealthCheck, error)
	// GetHealthCheckForUpdateTx is GetHealthCheckTx with the reading's row locked.
	GetHealthCheckForUpdateTx(
		ctx context.Context, tx *sqlx.Tx, id int,
	) (types.HealthCheck, error)
	// UpdateHealthCheckTx saves corrected readings and moves the version on by one.
	UpdateHealthCheckTx(
		ctx context.Context, tx *sqlx.Tx, id int, readings types.HealthReadings,
	) error
	// ListCustomerHealthChecksTx returns one page of a customer's readings, newest first, and how
	// many they have in all.
	ListCustomerHealthChecksTx(
		ctx context.Context, tx *sqlx.Tx, customerID, limit, offset int,
	) ([]types.HealthCheck, int, error)
}

type CustomerRepository interface {
	CreateCustomerTx(ctx context.Context, tx *sqlx.Tx, customer types.NewCustomer) (int, error)
	GetCustomerForUpdateTx(ctx context.Context, tx *sqlx.Tx, id int) (types.Customer, error)
	// UpdateCustomerTx saves the edited fields and moves the version on by one. The consent time
	// is the database's own and moves only when consent changes.
	UpdateCustomerTx(ctx context.Context, tx *sqlx.Tx, customer types.Customer) error
	// SetCustomerActiveTx also moves the version on, so a stale edit is refused.
	SetCustomerActiveTx(ctx context.Context, tx *sqlx.Tx, id int, active bool) error
	// CustomerByPhoneTx returns the customer with this normalized phone, or nil.
	CustomerByPhoneTx(ctx context.Context, tx *sqlx.Tx, phone string) (*types.Customer, error)
	GetCustomer(ctx context.Context, id int) (types.Customer, error)
	// GetCustomerRefTx reads a customer's name, whether they are active and whether they gave
	// health consent, and holds a share lock
	// on the row until the transaction ends.
	GetCustomerRefTx(ctx context.Context, tx *sqlx.Tx, id int) (types.CustomerRef, error)
	// LockCustomerForHealthCheckTx reads a customer's name, whether they are active and whether
	// they gave health consent, and locks the row for an update until the transaction ends.
	LockCustomerForHealthCheckTx(
		ctx context.Context, tx *sqlx.Tx, id int,
	) (types.CustomerRef, error)
	// SetCustomerHealthConsentTx records the customer's health consent, stamps when it changed and
	// moves the version on.
	SetCustomerHealthConsentTx(ctx context.Context, tx *sqlx.Tx, id int, consent bool) error
	// FetchCustomerNames returns the names of the given customers, deactivated ones included, on
	// the caller's transaction.
	FetchCustomerNames(ctx context.Context, tx *sqlx.Tx, ids []int) (map[int]string, error)
	ListCustomers(
		ctx context.Context, filter types.CustomerFilter, limit, offset int,
	) ([]types.Customer, int, error)
}

type SaleRepository interface {
	CreateSaleTx(ctx context.Context, tx *sqlx.Tx, sale model.Sale) (int, error)
	BulkCreateSaleItemsTx(ctx context.Context, tx *sqlx.Tx, saleItems []model.SaleItem) error
	BulkCreateSalePaymentsTX(ctx context.Context, tx *sqlx.Tx, salePayments []model.SalePayment) error
	FetchSalesTx(ctx context.Context, tx *sqlx.Tx, filter types.SaleFilter) ([]model.Sale, error)
	CountSales(ctx context.Context, filter types.SaleFilter) (int, error)
	SumSalesTotal(ctx context.Context, filter types.SaleFilter) (int, error)
	BulkFetchSaleItems(ctx context.Context, tx *sqlx.Tx, saleIDs []int) ([]model.SaleItem, error)
	BulkFetchSalePayments(ctx context.Context, tx *sqlx.Tx, saleIDs []int) ([]model.SalePayment, error)
	SaveHeldTransaction(ctx context.Context, transaction model.HeldTransaction) error
	FetchHeldTransactionsByType(ctx context.Context, transactionType constant.HoldTransactionType) ([]model.HeldTransaction, error)
	DeleteHeldTransactionByReference(ctx context.Context, reference string) error
	DeleteHeldTransactionByReferenceTx(ctx context.Context, tx *sqlx.Tx, reference string) error
	FetchSaleByID(ctx context.Context, tx *sqlx.Tx, saleID int) (model.Sale, error)
	FetchSaleReceiptNumberByID(ctx context.Context, tx *sqlx.Tx, saleID int) (string, error)
	FetchSaleByKeyTx(ctx context.Context, tx *sqlx.Tx, key string) (types.SavedSale, error)
	// FetchSaleCustomerForUpdateTx reads a sale's receipt number, customer and whether it was made
	// today, and locks the sale until the transaction ends (sql.ErrNoRows if there is none).
	FetchSaleCustomerForUpdateTx(
		ctx context.Context, tx *sqlx.Tx, saleID int,
	) (types.SaleCustomerState, error)
	// UpdateSaleCustomerTx sets a sale's customer; nil makes it a walk-in.
	UpdateSaleCustomerTx(ctx context.Context, tx *sqlx.Tx, saleID int, customerID *int) error

	ReturnRepository
}

type ReturnRepository interface {
	CreateReturnTx(ctx context.Context, tx *sqlx.Tx, rtn model.Return) (int, error)
	BulkCreateReturnItemsTx(ctx context.Context, tx *sqlx.Tx, returnItems []model.ReturnItems) error
	FetchAllSaleReturns(ctx context.Context, tx *sqlx.Tx, saleID int) ([]model.ReturnItems, error)
	BulkFetchReturnItemsBySaleIDs(ctx context.Context, saleIDs []int) ([]model.ReturnItemWithSale, error)
	SumReturnTotal(ctx context.Context, filter types.SaleFilter) (int, error)
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
	GetExpiringCount(ctx context.Context) (int, error)
	GetLowStockItems(ctx context.Context) ([]model.LowStockItem, error)
	GetSalesByCategory(ctx context.Context, startDate, endDate time.Time) ([]model.SalesByCategory, error)
	GetExpiringItems(ctx context.Context, startDate, endDate time.Time) ([]model.ExpiringItem, error)
	GetExpiringItemsByCategory(ctx context.Context, startDate, endDate time.Time) ([]model.ExpiryByCategory, error)
	GetTopSellingProducts(ctx context.Context, startDate, endDate time.Time, limit int) ([]model.TopSellingProduct, error)
	GetTopSellingProductsAllTime(ctx context.Context, limit int) ([]model.TopSellingProduct, error)
	GetRecentSales(ctx context.Context, startDate, endDate time.Time, limit int) ([]model.RecentTransaction, error)
}

type AdminRepository interface {
	GetAllPermissions(ctx context.Context) ([]model.Permission, error)
	CreatePermission(ctx context.Context, resource, action string) (int, error)
	DeletePermission(ctx context.Context, id int) error
	GetRolePermissions(ctx context.Context, roleID int) ([]model.Permission, error)
	AssignPermissionToRole(ctx context.Context, roleID, permissionID int) error
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID int) error
	GetAllRolesWithPermissions(ctx context.Context) ([]types.RoleWithPermissions, error)
	ListUsers(ctx context.Context) ([]types.UserListItem, error)
	UpdateUserRole(ctx context.Context, userID, roleID int) error
	UpdateUserPassword(ctx context.Context, userID int, hashedPassword string) error
	CreateCategory(ctx context.Context, name string) (int, error)
	UpdateCategory(ctx context.Context, id int, name string) error
	DeleteCategory(ctx context.Context, id int) error
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
	AdminRepository
	StockAdjustmentRepository
	CustomerRepository
	HealthCheckRepository
}
