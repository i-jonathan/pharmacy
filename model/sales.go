package model

import (
	"fmt"
	"math/rand"
	"time"
)

type Sale struct {
	baseModel
	ReceiptNumber string `db:"receipt_number"`
	Status        string `db:"status"`
	CashierID     int    `db:"cashier_id"`
	Subtotal      int    `db:"subtotal"`
	Discount      int    `db:"discount"`
	Total         int    `db:"total"`

	SaleItems []SaleItem
	Payments  []SalePayment
}

type SaleItem struct {
	baseModel
	SaleID     int `db:"sale_id"`
	ProductID  int `db:"product_id"`
	Quantity   int `db:"quantity"`
	UnitPrice  int `db:"unit_price"`
	Discount   int `db:"discount"`
	TotalPrice int `db:"total_price"`

	Product Product
}

type SalePayment struct {
	baseModel
	SaleID        int    `db:"sale_id"`
	Amount        int    `db:"amount"`
	PaymentMethod string `db:"payment_method"`
}

type Return struct {
	baseModel
	SaleID        int    `db:"sale_id"`
	CashierID     int    `db:"cashier_id"`
	TotalRefunded int    `db:"total_refunded"`
	Notes         string `db:"notes"`
	Sale          Sale
	Cashier       User
}

type ReturnItems struct {
	baseModel
	ReturnID   int `db:"return_id"`
	SaleItemID int `db:"sale_item_id"`
	Quantity   int `db:"quantity"`
	Return     Return
	SaleItem   SaleItem
}

func (s *Sale) GenerateReceiptNumber() {
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	days := int(time.Since(start).Hours() / 24)
	random := rand.Intn(9000) + 1000

	s.ReceiptNumber = fmt.Sprintf("%d%04d", days, random)
}
