package model

import "fmt"

type Product struct {
	baseModel
	Name           string
	BarCode        *string
	CategoryID     int
	Category       Category
	ReorderLevel   int
	Manufacturer   *string
	CostPriceKobo  int
	DefaultPriceID int
	DefaultPrice   ProductPrice
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
