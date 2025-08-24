package types

type SaleItem struct {
	ProductID int     `json:"product_id"`
	Quantity  int     `json:"quantity"`
	PriceID   int     `json:"price_id"`
	UnitPrice float64 `json:"unit_price"`
	Discount  float64 `json:"discount"`
	Total     float64 `json:"total"`
}

type SalePayment struct {
	Amount        float64 `json:"amount"`
	PaymentMethod string  `json:"payment_method"`
}

type Sale struct {
	CashierID int
	Subtotal  float64       `json:"subtotal"`
	Discount  float64       `json:"discount"`
	Total     float64       `json:"total"`
	Items     []SaleItem    `json:"items"`
	Payments  []SalePayment `json:"payments"`
}
