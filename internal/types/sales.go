package types

import (
	"encoding/json"
	"time"
)

type SaleFilter struct {
	StartDate *time.Time
	EndDate   *time.Time
}

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

type SaleItemResponse struct {
	ProductName  string  `json:"product_name"`
	Manufacturer string  `json:"manufacturer"`
	Quantity     int     `json:"quantity"`
	UnitPrice    float64 `json:"unit_price"`
	Discount     float64 `json:"discount"`
}

type SalePaymentResponse struct {
	MethodName string  `json:"method_name"`
	Amount     float64 `json:"amount"`
}

type SaleResponse struct {
	ReceiptNumber string                `json:"receipt_number"`
	Cashier       string                `json:"cashier"`
	CreatedAt     string                `json:"created_at"`
	Subtotal      float64               `json:"subtotal"`
	Discount      float64               `json:"discount"`
	Total         float64               `json:"total"`
	Items         []SaleItemResponse    `json:"items"`
	Payments      []SalePaymentResponse `json:"payments"`
}

type SaleHistory struct {
	TotalAmount float64        `json:"total"`
	Data        []SaleResponse `json:"data"`
}

type HoldTransactionRequest struct {
	Reference string          `json:"reference"`
	Payload   json.RawMessage `json:"payload"`
}
