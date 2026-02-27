package model

import (
	"encoding/json"
	"fmt"
	"time"
)

type Product struct {
	baseModel
	Name           string        `db:"name"`
	Barcode        *string       `db:"barcode"`
	CategoryID     int           `db:"category_id"`
	ReorderLevel   int           `db:"reorder_level"`
	Manufacturer   *string       `db:"manufacturer"`
	CostPriceKobo  int           `db:"cost_price"`
	DefaultPriceID int           `db:"default_price_id"`
	Category       Category      `db:"category"`
	DefaultPrice   ProductPrice  `db:"default_price"`
	PriceOptions   ProductPrices `db:"price_options" json:"price_options"`
}

type ProductPrices []ProductPrice

func (pp *ProductPrices) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan ProductPrices: %v", value)
	}
	return json.Unmarshal(b, pp)
}

type ProductPrice struct {
	baseModel
	Name             string `db:"name" json:"name"`
	ProductID        int    `db:"product_id"`
	QuantityPerUnit  int    `db:"quantity_per_unit" json:"quantity_per_unit"`
	SellingPriceKobo int    `db:"selling_price" json:"selling_price"`
}

type ReceivingBatch struct {
	baseModel
	SupplierName string  `db:"supplier_name"`
	ReceviedByID int     `db:"received_by_id"`
	Note         *string `db:"note"`
	ReceivedBy   User    `db:"user"`
}

type ProductBatch struct {
	baseModel
	ProductID        int        `db:"product_id"`
	PriceID          int        `db:"price_id"`
	Quantity         int        `db:"quantity"`
	CostPriceKobo    int        `db:"cost_price"`
	ReceivingBatchID int        `db:"receiving_batch_id"`
	BatchNo          string     `db:"batch_no"`
	ExpiryDate       *time.Time `db:"expiry_date"`
	Product          Product
	ReceivingBatch   ReceivingBatch
}

type StockMovement struct {
	baseModel
	ProductID    int    `db:"product_id"`
	MovementType string `db:"movement_type"`
	Quantity     int    `db:"quantity"`
	ReferenceID  int    `db:"reference_id"`
}

type InventoryItem struct {
	ID             int        `db:"id" json:"id"`
	CategoryID     int        `db:"category_id" json:"category_id"`
	DefaultPriceID int        `db:"default_price_id" json:"default_price_id"`
	DefaultPrice   int        `db:"default_price" json:"default_price"`
	ReorderLevel   int        `db:"reorder_level" json:"reorder_level"`
	Stock          int        `db:"stock" json:"stock"`
	Category       string     `db:"category" json:"category"`
	Manufacturer   string     `db:"manufacturer" json:"manufacturer"`
	Name           string     `db:"name" json:"name"`
	EarliestExpiry *time.Time `db:"earliest_expiry" json:"earliest_expiry,omitempty"`
}

type Inventory struct {
	Items      []InventoryItem
	Categories []Category
}

func (p ProductPrice) PriceString() string {
	return fmt.Sprintf("₦%.2f", float64(p.SellingPriceKobo)/100)
}

func (p Product) CostPriceString() string {
	return fmt.Sprintf("₦%.2f", float64(p.CostPriceKobo)/100)
}

func (p Product) CostPriceFloat() float64 {
	return float64(p.CostPriceKobo) / 100
}

func (p Product) DefaultSellingPriceKobo() int {
	return p.DefaultPrice.SellingPriceKobo
}

func (p Product) DefaultSellingPriceString() string {
	return p.DefaultPrice.PriceString()
}

func (pp ProductPrice) SellingPriceFloat() float64 {
	return float64(pp.SellingPriceKobo) / 100
}
