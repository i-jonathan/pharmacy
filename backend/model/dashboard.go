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

type LowStockItem struct {
	ProductName  string `json:"product_name" db:"product_name"`
	Manufacturer string `json:"manufacturer" db:"manufacturer"`
	ProductID    int    `json:"product_id" db:"product_id"`
	CurrentStock int    `json:"current_stock" db:"current_stock"`
	ReorderLevel int    `json:"reorder_level" db:"reorder_level"`
}
