package types

import "time"

type DashboardResponse struct {
	KPI           KPIResponse         `json:"kpi"`
	SalesTrend    []SalesTrendData    `json:"sales_trend"`
	CategorySales []CategorySalesData `json:"category_sales"`
	ExpiringItems []ExpiringItemData  `json:"expiring_items"`
	LowStockItems []LowStockItemData  `json:"low_stock_items"`
}

type KPIResponse struct {
	TodaySales        int     `json:"today_sales"`
	TodayTransactions int     `json:"today_transactions"`
	TotalInventory    int     `json:"total_inventory"`
	LowStockCount     int     `json:"low_stock_count"`
	SalesTrend        float64 `json:"sales_trend"`       // percentage change
	TransactionTrend  float64 `json:"transaction_trend"` // percentage change
}

type SalesTrendData struct {
	Day   string `json:"day"`
	Sales int    `json:"sales"`
}

type CategorySalesData struct {
	Category string `json:"category"`
	Sales    int    `json:"sales"` // percentage
}

type ExpiryAlertData struct {
	Count int     `json:"count"`
	Value float64 `json:"value"`
}

type ExpiringItemData struct {
	ID              int       `json:"id"`
	ProductName     string    `json:"product_name"`
	Quantity        int       `json:"quantity"`
	CostPriceKobo   int       `json:"cost_price_kobo"`
	ExpiryDate      time.Time `json:"expiry_date"`
	DaysUntilExpiry int       `json:"days_until_expiry"`
}

type LowStockItemData struct {
	ID           int    `json:"id"`
	ProductName  string `json:"product_name"`
	CurrentStock int    `json:"current_stock"`
	ReorderLevel int    `json:"reorder_level"`
}

type DashboardFilter struct {
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
}
