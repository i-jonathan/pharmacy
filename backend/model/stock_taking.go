package model

import "time"

type StockTakingStatus string

const (
	StockTakingInProgress StockTakingStatus = "in_progress"
	StockTakingCompleted  StockTakingStatus = "completed"
	StockTakingCancelled  StockTakingStatus = "cancelled"
)

type StockTaking struct {
	baseModel
	Name          string            `db:"name"`
	Status        StockTakingStatus `db:"status"`
	CreatedByID   int               `db:"created_by_id"`
	StartedAt     time.Time         `db:"started_at"`
	CompletedAt   *time.Time        `db:"completed_at"`
	CreatedByName string            `db:"created_by_name"`
}

type StockTakingItem struct {
	baseModel
	StockTakingID    int        `db:"stock_taking_id"`
	ProductID        int        `db:"product_id"`
	SnapshotQuantity int        `db:"snapshot_quantity"`
	DispensaryCount  int        `db:"dispensary_count"`
	StoreCount       int        `db:"store_count"`
	LastUpdatedByID  int        `db:"last_updated_by_id"`
	LastUpdatedAt    *time.Time `db:"last_updated_at"`
	Notes            string     `db:"notes"`
}
