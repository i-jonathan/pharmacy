package types

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
	SellingPrice float64 `json:"selling_price"`
}

type ProductResult struct {
	ID           int                `json:"id"`
	Name         string             `json:"name"`
	Manufacturer string             `json:"manufacturer"`
	Barcode      string             `json:"barcode"`
	CostPrice    float64            `json:"cost_price"`
	DefaultPrice ProductPriceResult `json:"default_price"`
}
