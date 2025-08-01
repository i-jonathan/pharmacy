package types

type CreateProductRequest struct {
	Name         string
	Barcode      string
	Manufacturer string
	CategoryID   int
	ReorderLevel int
	CostPrice    float64
	SellingPrice float64
}
