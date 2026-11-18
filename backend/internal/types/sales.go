package types

import (
	"encoding/json"
	"time"
)

type SaleFilter struct {
	StartDate *time.Time
	EndDate   *time.Time
	Page      int
	PerPage   int
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
	CashierID         int
	HeldSaleReference string        `json:"held_sale_reference"`
	CustomerID        int           `json:"customer_id"` // picked at the till; 0 is a walk-in
	IdempotencyKey    string        `json:"idempotency_key"`
	Subtotal          float64       `json:"subtotal"`
	Discount          float64       `json:"discount"`
	Total             float64       `json:"total"`
	Items             []SaleItem    `json:"items"`
	Payments          []SalePayment `json:"payments"`
}

type SaleItemResponse struct {
	ID               int     `json:"id"`
	ProductName      string  `json:"product_name"`
	Manufacturer     string  `json:"manufacturer"`
	Quantity         int     `json:"quantity"`
	ReturnedQuantity int     `json:"returned_quantity"`
	UnitPrice        float64 `json:"unit_price"`
	Discount         float64 `json:"discount"`
}

type SalePaymentResponse struct {
	MethodName string  `json:"method_name"`
	Amount     float64 `json:"amount"`
}

type ReturnItemResponse struct {
	SaleItemID   int     `json:"sale_item_id"`
	Name         string  `json:"product_name"`
	Manufacturer string  `json:"manufacturer"`
	UnitPrice    float64 `json:"unit_price"`
	Quantity     int     `json:"quantity"`
}

// SavedSale is what an earlier save under a retry key stored.
type SavedSale struct {
	ReceiptNumber string `db:"receipt_number"`
	CustomerID    *int   `db:"customer_id"`
}

// SaleCustomerState is a saved sale as a customer change needs to see it, read with the sale
// locked. IsToday is decided by the database, on the same date basis Sales History filters by.
type SaleCustomerState struct {
	ReceiptNumber string `db:"receipt_number"`
	CustomerID    *int   `db:"customer_id"`
	IsToday       bool   `db:"is_today"`
}

// ChangeSaleCustomerRequest changes or removes (CustomerID 0) the customer on a saved sale.
// ExpectedCustomerID is the customer the page last saw (0 for a walk-in).
type ChangeSaleCustomerRequest struct {
	SaleID             int    `json:"-"`
	CustomerID         int    `json:"customer_id"`
	ExpectedCustomerID int    `json:"expected_customer_id"`
	Reason             string `json:"reason"`
	Note               string `json:"note"`
}

// SaleCustomer is who a sale was made to, as Sales History shows it.
type SaleCustomer struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type SaleResponse struct {
	ID            int                   `json:"id"`
	ReceiptNumber string                `json:"receipt_number"`
	Cashier       string                `json:"cashier"`
	CreatedAt     string                `json:"created_at"`
	Subtotal      float64               `json:"subtotal"`
	Discount      float64               `json:"discount"`
	Total         float64               `json:"total"`
	Items         []SaleItemResponse    `json:"items"`
	Payments      []SalePaymentResponse `json:"payments"`
	Returns       []ReturnItemResponse  `json:"returns"`
	// MadeToday says whether the database dates the sale today (the basis of the rule that only a
	// manager changes an earlier day's sale), so the page need not guess from the browser's clock.
	MadeToday bool `json:"made_today"`
	// Customer is set only for a sale with a customer, and only for users who may see customers.
	Customer *SaleCustomer `json:"customer"`
}

// ReceiptReprint is the server's answer to a reprint: the sale's number and when it was
// reprinted, so the copy can say so.
type ReceiptReprint struct {
	ReceiptNumber string `json:"receipt_number"`
	ReprintedAt   string `json:"reprinted_at"`
}

type SaleHistory struct {
	TotalAmount float64        `json:"total"`
	Data        []SaleResponse `json:"data"`
	Page        int            `json:"page"`
	PerPage     int            `json:"per_page"`
	TotalCount  int            `json:"total_count"`
}

type HoldTransactionRequest struct {
	Reference string          `json:"reference"`
	Payload   json.RawMessage `json:"payload"`
}

type HeldTransactionResponse struct {
	Reference string          `json:"reference"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type ReturnSaleItem struct {
	SaleItemID int `json:"sale_item_id"`
	Quantity   int `json:"quantity"`
}

type ReturnSale struct {
	SaleID      int              `json:"sale_id"`
	CashierID   int              `json:"cashier_id"`
	Notes       string           `json:"notes"`
	ReturnItems []ReturnSaleItem `json:"return_items"`
}
