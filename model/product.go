package model

import "fmt"

type Product struct {
	baseModel
	Name           string       `db:"name"`
	Barcode        *string      `db:"barcode"`
	CategoryID     int          `db:"category_id"`
	ReorderLevel   int          `db:"reorder_level"`
	Manufacturer   *string      `db:"manufacturer"`
	CostPriceKobo  int          `db:"cost_price"`
	DefaultPriceID int          `db:"default_price_id"`
	Category       Category     `db:"category"`
	DefaultPrice   ProductPrice `db:"default_price"`
	PriceOptions   []ProductPrice
}

type ProductPrice struct {
	baseModel
	ProductID        int
	UnitName         string
	QuantityPerUnit  int
	SellingPriceKobo int
	IsDefault        bool
}

func (p ProductPrice) PriceString() string {
	return fmt.Sprintf("₦%.2f", float64(p.SellingPriceKobo)/100)
}

func (p Product) CostPriceString() string {
	return fmt.Sprintf("₦%.2f", float64(p.CostPriceKobo)/100)
}

func (p Product) DefaultSellingPriceKobo() int {
	return p.DefaultPrice.SellingPriceKobo
}

func (p Product) DefaultSellingPriceString() string {
	return p.DefaultPrice.PriceString()
}
