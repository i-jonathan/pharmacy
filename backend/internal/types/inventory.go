package types

import "time"

type ReceivedBatchItem struct {
	ProductID    int        `db:"product_id" json:"product_id"`
	ProductName  string     `db:"product_name" json:"product_name"`
	Manufacturer string     `db:"manufacturer" json:"manufacturer"`
	Quantity     int        `db:"quantity" json:"quantity"`
	CostPrice    float64    `db:"cost_price" json:"cost_price"`
	BatchNo      *string    `db:"batch_no" json:"batch_no"`
	ExpiryDate   *time.Time `db:"expiry_date" json:"expiry_date"`
}

type ReceivedBatch struct {
	ID           int                 `db:"id" json:"id"`
	SupplierName string              `db:"supplier_name" json:"supplier_name"`
	ReceivedBy   string              `db:"received_by" json:"received_by"`
	Note         *string             `db:"note" json:"note"`
	CreatedAt    time.Time           `db:"created_at" json:"created_at"`
	Items        []ReceivedBatchItem `json:"items"`
}

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
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Barcode      string `json:"barcode"`
	Manufacturer string `json:"manufacturer"`
}

type CategoriesResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ProductPriceResult struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	SellingPrice float64 `json:"selling_price"`
	Quantity     int     `json:"quantity"`
}

type ProductResult struct {
	ID           int                  `json:"id"`
	Name         string               `json:"name"`
	Manufacturer string               `json:"manufacturer"`
	Barcode      string               `json:"barcode"`
	CostPrice    float64              `json:"cost_price"`
	CategoryID   int                  `json:"category_id"`
	ReorderLevel int                  `json:"reorder_level"`
	Stock        int                  `json:"stock"`
	DefaultPrice ProductPriceResult   `json:"default_price"`
	PriceOptions []ProductPriceResult `json:"price_options"`
}

type BatchInsertReturn struct {
	ID        int `db:"id"`
	ProductID int `db:"product_id"`
	Quantity  int `db:"quantity"`
}

type ReceiveItem struct {
	ID                  int                 `json:"id"`
	Barcode             string              `json:"barcode"`
	CostPrice           float64             `json:"cost_price"`
	SellingPrice        float64             `json:"selling_price"`
	Quantity            int                 `json:"quantity"`
	Expiry              time.Time           `json:"expiry" time_format:"2006-01-02"`
	PriceOptionsChanges []PriceOptionChange `json:"price_options_changes"`
}

type PriceOptionChange struct {
	ID              *int    `json:"id"`
	Name            string  `json:"name"`
	SellingPrice    float64 `json:"selling_price"`
	QuantityPerUnit int     `json:"quantity_per_unit"`
}

type ReceiveSupplyRequest struct {
	Supplier               string        `json:"supplier"`
	HeldReceivingReference string        `json:"held_receiving_reference"`
	IdempotencyKey         string        `json:"idempotency_key"`
	Products               []ReceiveItem `json:"products"`
	UserID                 int
}

type ProductPriceUpdate struct {
	ID              *int    `json:"id"`
	Name            string  `json:"name"`
	SellingPrice    float64 `json:"selling_price"`
	QuantityPerUnit int     `json:"quantity_per_unit"`
	IsDefault       bool    `json:"is_default"`
}

type UpdateProductRequest struct {
	ID           int                  `json:"id"`
	Name         string               `json:"name"`
	Barcode      string               `json:"barcode"`
	Manufacturer string               `json:"manufacturer"`
	CategoryID   int                  `json:"category_id"`
	ReorderLevel int                  `json:"reorder_level"`
	Stock        int                  `json:"stock"`
	CostPrice    float64              `json:"cost_price"`
	Prices       []ProductPriceUpdate `json:"prices"`
}
