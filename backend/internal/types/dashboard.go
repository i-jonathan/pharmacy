package types

import "time"

type DashboardResponse struct {
	KPI               KPIResponse            `json:"kpi"`
	SalesTrend        []SalesTrendData       `json:"sales_trend"`
	CategorySales     []CategorySalesData    `json:"category_sales"`
	TopSellingProducts []TopSellingProductData `json:"top_selling_products"`
	ExpiringItems      []ExpiringItemData      `json:"expiring_items"`
	ExpiryByCategory   []ExpiryByCategoryData  `json:"expiry_by_category"`
	LowStockItems      []LowStockItemData      `json:"low_stock_items"`
	RecentTransactions []RecentTransactionData `json:"recent_transactions"`
}

type KPIResponse struct {
	TodaySales        int     `json:"today_sales"`
	TodayTransactions int     `json:"today_transactions"`
	TotalInventory    int     `json:"total_inventory"`
	LowStockCount     int     `json:"low_stock_count"`
	ExpiringCount     int     `json:"expiring_count"`
	SalesTrend        float64 `json:"sales_trend"`       // percentage change
	TransactionTrend  float64 `json:"transaction_trend"` // percentage change
}

type SalesTrendData struct {
	Date  string `json:"date"`
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
	ProductName  string `json:"product_name"`
	Manufacturer string `json:"manufacturer"`
	ID           int    `json:"id"`
	CurrentStock int    `json:"current_stock"`
	ReorderLevel int    `json:"reorder_level"`
}

type ExpiryByCategoryData struct {
	Category      string `json:"category"`
	Count         int    `json:"count"`
	TotalCostKobo int    `json:"total_cost_kobo"`
}

type TopSellingProductData struct {
	ProductName string `json:"product_name"`
	Quantity    int    `json:"quantity"`
	RevenueKobo int    `json:"revenue_kobo"`
}

type RecentTransactionData struct {
	ID            int       `json:"id"`
	ReceiptNumber string    `json:"receipt_number"`
	Total         int       `json:"total"`
	ItemCount     int       `json:"item_count"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

type DashboardFilter struct {
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
}
