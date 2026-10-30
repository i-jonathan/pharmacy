package types

import "time"

// StockAdjustmentItemRequest changes one product's stock: "set" to a counted quantity, or "add"
// or "remove" a quantity, in base units. Expiry, as YYYY-MM-DD, optionally corrects the
// product's current expiry date.
type StockAdjustmentItemRequest struct {
	ProductID int     `json:"product_id"`
	Mode      string  `json:"mode"`
	Quantity  int     `json:"quantity"`
	Expiry    *string `json:"expiry"`
}

// StockAdjustmentRequest is a manager correcting stock for one reason. FlagIDs are open flags
// this adjustment answers; they are resolved with it.
type StockAdjustmentRequest struct {
	Reason         string                       `json:"reason"`
	Note           string                       `json:"note"`
	IdempotencyKey string                       `json:"idempotency_key"`
	Items          []StockAdjustmentItemRequest `json:"items"`
	FlagIDs        []int                        `json:"flag_ids"`
}

// ProductStock is a product's current stock, read under a row lock inside a transaction.
type ProductStock struct {
	ID            int        `db:"id"`
	Name          string     `db:"name"`
	CostPriceKobo int        `db:"cost_price"`
	CurrentExpiry *time.Time `db:"current_expiry"`
	Stock         int        `db:"stock"`
}

// NewStockAdjustment is the header row of an adjustment.
type NewStockAdjustment struct {
	Reason         string
	Note           string
	CreatedByID    int
	IdempotencyKey *string
}

// NewStockAdjustmentItem is one product's change within an adjustment.
type NewStockAdjustmentItem struct {
	ProductID      int
	QuantityBefore int
	QuantityChange int
	UnitCostKobo   int
	ExpiryBefore   *time.Time
	ExpiryAfter    *time.Time
}

// StockAdjustmentItem is a stored item with the product's name.
type StockAdjustmentItem struct {
	ProductID      int        `json:"product_id"      db:"product_id"`
	ProductName    string     `json:"product_name"    db:"product_name"`
	Manufacturer   string     `json:"manufacturer"    db:"manufacturer"`
	QuantityBefore int        `json:"quantity_before" db:"quantity_before"`
	QuantityChange int        `json:"quantity_change" db:"quantity_change"`
	ExpiryBefore   *time.Time `json:"expiry_before"   db:"expiry_before"`
	ExpiryAfter    *time.Time `json:"expiry_after"    db:"expiry_after"`
}

// StockAdjustment is a stored adjustment with who made it and its items.
type StockAdjustment struct {
	ID        int                   `json:"id"         db:"id"`
	Reason    string                `json:"reason"     db:"reason"`
	Note      string                `json:"note"       db:"note"`
	CreatedBy string                `json:"created_by" db:"created_by"`
	CreatedAt time.Time             `json:"created_at" db:"created_at"`
	Items     []StockAdjustmentItem `json:"items"`
}

// StockAdjustmentPage is one page of the adjustment history.
type StockAdjustmentPage struct {
	Data       []StockAdjustment `json:"data"`
	Page       int               `json:"page"`
	PerPage    int               `json:"per_page"`
	TotalCount int               `json:"total_count"`
}

// StockFlagRequest is someone asking a manager to check a product's stock.
type StockFlagRequest struct {
	ProductID       int    `json:"product_id"`
	Note            string `json:"note"`
	CountedQuantity *int   `json:"counted_quantity"`
}

// StockFlag is an open flag with the product's current stock. MovedSince is what sales, receipts
// and returns changed since the flag was raised; stock takings and adjustments are corrections,
// not shelf movements, so they are left out.
type StockFlag struct {
	ID              int       `json:"id"               db:"id"`
	ProductID       int       `json:"product_id"       db:"product_id"`
	ProductName     string    `json:"product_name"     db:"product_name"`
	Manufacturer    string    `json:"manufacturer"     db:"manufacturer"`
	Note            string    `json:"note"             db:"note"`
	CountedQuantity *int      `json:"counted_quantity" db:"counted_quantity"`
	CurrentStock    int       `json:"current_stock"    db:"current_stock"`
	StockAtFlag     int       `json:"stock_at_flag"    db:"stock_at_flag"`
	MovedSince      int       `json:"moved_since"      db:"moved_since"`
	FlaggedBy       string    `json:"flagged_by"       db:"flagged_by"`
	CreatedAt       time.Time `json:"created_at"       db:"created_at"`
	Status          string    `json:"status"           db:"status"`
}

// DismissStockFlagRequest closes a flag without adjusting stock.
type DismissStockFlagRequest struct {
	Note string `json:"note"`
}
