package model

import "time"

type SalesByCategory struct {
	Category string `db:"category"`
	Sales    int    `db:"sales"`
}

type ExpiringItem struct {
	ProductID       int       `json:"product_id" db:"product_id"`
	ProductName     string    `json:"product_name" db:"product_name"`
	Quantity        int       `json:"quantity" db:"quantity"`
	CostPriceKobo   int       `json:"cost_price" db:"cost_price"`
	ExpiryDate      time.Time `json:"expiry_date" db:"expiry_date"`
	DaysUntilExpiry int       `json:"days_until_expiry" db:"days_until_expiry"`
}

type TopSellingProduct struct {
	ProductName string `db:"product_name"`
	Quantity    int    `db:"quantity"`
	RevenueKobo int    `db:"revenue_kobo"`
}

type LowStockItem struct {
	ProductName  string `json:"product_name" db:"product_name"`
	Manufacturer string `json:"manufacturer" db:"manufacturer"`
	ProductID    int    `json:"product_id" db:"product_id"`
	CurrentStock int    `json:"current_stock" db:"current_stock"`
	ReorderLevel int    `json:"reorder_level" db:"reorder_level"`
}

type RecentTransaction struct {
	ID            int       `db:"id"`
	ReceiptNumber string    `db:"receipt_number"`
	Total         int       `db:"total"`
	ItemCount     int       `db:"item_count"`
	Status        string    `db:"status"`
	CreatedAt     time.Time `db:"created_at"`
}

type ExpiryByCategory struct {
	Category      string `db:"category"`
	Count         int    `db:"count"`
	TotalCostKobo int    `db:"total_cost_kobo"`
}
