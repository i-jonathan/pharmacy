package types

import "time"

type StockTakingItemData struct {
	ProductID         int         `db:"product_id"`
	StockTakingItemID int         `db:"stock_taking_item_id"`
	SnapshotQuantity  int         `db:"snapshot_quantity"`
	DispensaryCount   int         `db:"dispensary_count"`
	StoreCount        int         `db:"store_count"`
	ProductName       string      `db:"product_name"`
	Manufacturer      string      `db:"manufacturer"`
	Notes             string      `db:"notes"`
	LastUpdatedBy     string      `db:"last_updated_by"`
	EarliestExpiry    time.Time   `db:"earliest_expiry"`
	ExpiryOptions     []time.Time `db:"expiry_options"`
	LastUpdatedAt     *time.Time  `db:"last_updated_at"`
}

type StockTakingData struct {
	Name        string
	Status      string
	CreatedBy   string
	StartedAt   time.Time
	CompletedAt *time.Time
}

type StockTakingItems []StockTakingItemData
