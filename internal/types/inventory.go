package types

import "time"

type CreateProductRequest struct {
	Name         string  `json:"name"`
	Barcode      string  `json:"barcode"`
	Manufacturer string  `json:"manufacturer"`
	CategoryID   int     `json:"category_id"`
	ReorderLevel int     `json:"reorder_level"`
	CostPrice    float64 `json:"cost_price"`
	SellingPrice float64 `json:"selling_price"`
}

type AddItemResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type CategoriesResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ProductPriceResult struct {
	ID           int     `json:"id"`
	Name         string  `json:"name:`
	SellingPrice float64 `json:"selling_price"`
}

type ProductResult struct {
	ID           int                  `json:"id"`
	Name         string               `json:"name"`
	Manufacturer string               `json:"manufacturer"`
	Barcode      string               `json:"barcode"`
	CostPrice    float64              `json:"cost_price"`
	DefaultPrice ProductPriceResult   `json:"default_price"`
	PriceOptions []ProductPriceResult `json:"price_options"`
}

type BatchInsertReturn struct {
	ID        int `db:"id"`
	ProductID int `db:"product_id"`
	Quantity  int `db:"quantity"`
}

type ReceiveItem struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Barcode      string    `json:"barcode"`
	CostPrice    float64   `json:"cost_price"`
	SellingPrice float64   `json:"selling_price"`
	Quantity     int       `json:"quantity"`
	Expiry       time.Time `json:"expiry" time_format:"2006-01-02"`
}

type ReceiveSupplyRequest struct {
	Supplier string        `json:"supplier"`
	Products []ReceiveItem `json:"products"`
	UserID   int
}
