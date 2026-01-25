package types

import "time"

type StockTakingItemData struct {
	ProductID         int        `db:"product_id" json:"product_id"`
	StockTakingItemID *int       `db:"stock_taking_item_id" json:"stock_taking_item_id"`
	SnapshotQuantity  *int       `db:"snapshot_quantity" json:"snapshot_quantity"`
	DispensaryCount   *int       `db:"dispensary_count" json:"dispensary_count"`
	StoreCount        *int       `db:"store_count" json:"store_count"`
	ProductName       string     `db:"product_name" json:"product_name"`
	Manufacturer      string     `db:"manufacturer" json:"manufacturer"`
	Notes             *string    `db:"notes" json:"notes"`
	LastUpdatedBy     *string    `db:"last_updated_by" json:"last_updated_by"`
	EarliestExpiry    *time.Time `db:"earliest_expiry" json:"earliest_expiry"`
	ExpiryOptions     TimeArray  `db:"expiry_options" json:"expiry_options"`
	LastUpdatedAt     *time.Time `db:"last_updated_at" json:"last_updated_at,omitempty"`
}

type StockTakingData struct {
	Name        string     `json:"name"`
	Status      string     `json:"status"`
	CreatedBy   string     `json:"created_by"`
	CreatedByID int        `json:"created_by_id"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type StockTakingItems []StockTakingItemData

type StockTakingResponse struct {
	StockTakingData StockTakingData  `json:"stock_taking_data"`
	Items           StockTakingItems `json:"items"`
	Permissions     map[string]bool  `json:"permissions"` // temporary until login page moves to vue
}
