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
