package types

import (
	"encoding/json"
	"time"
)

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
	Quantity     int     `json:"quantity_per_unit"`
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
	PurchaseOrderItemID int                 `json:"purchase_order_item_id"`
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
	PurchaseOrderID        int `json:"purchase_order_id"`
}

type PurchaseOrderItem struct {
	ID               int    `json:"id" db:"id"`
	ProductID        *int   `json:"product_id" db:"product_id"`
	Name             string `json:"name" db:"item_name"`
	Manufacturer     string `json:"manufacturer" db:"manufacturer"`
	Note             string `json:"note" db:"note"`
	Quantity         int    `json:"quantity" db:"quantity"`
	ReceivedQuantity int    `json:"received_quantity" db:"received_quantity"`
}

type Supplier struct {
	ID   int    `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

type PurchaseOrder struct {
	ID          int        `json:"id" db:"id"`
	Supplier    string     `json:"supplier" db:"supplier_name"`
	SupplierID  *int       `json:"supplier_id" db:"supplier_id"`
	Status      string     `json:"status" db:"status"`
	Note        string     `json:"note" db:"note"`
	CreatedByID *int       `json:"created_by_id" db:"created_by_id"`
	CreatedBy   string     `json:"created_by" db:"created_by"`
	ApprovedBy  string     `json:"approved_by" db:"approved_by"`
	ApprovedAt  *time.Time `json:"approved_at" db:"approved_at"`
	// Version changes on every material change; edits must send the version they loaded.
	Version int `json:"version" db:"version"`
	// SourceOrderID is the order whose outstanding items were moved onto this one, and
	// MovedToID the order that took over this one's outstanding items.
	SourceOrderID *int                `json:"source_purchase_order_id" db:"source_purchase_order_id"`
	MovedToID     *int                `json:"moved_to_purchase_order_id" db:"moved_to_id"`
	CreatedAt     time.Time           `json:"created_at" db:"created_at"`
	Items         []PurchaseOrderItem `json:"items"`
}

// NewPurchaseOrder holds the header fields of an order about to be created.
type NewPurchaseOrder struct {
	Status        string
	SupplierID    *int
	SupplierName  string
	SourceOrderID *int
	Note          string
	CreatedByID   int
}

type MoveOutstandingRequest struct {
	Version  int    `json:"version"`
	Supplier string `json:"supplier"`
	Reason   string `json:"reason"`
}

// PurchaseOrderActionRequest is the body of submit and approve: just the version the user saw.
type PurchaseOrderActionRequest struct {
	Version int `json:"version"`
}

// PendingPurchaseOrders counts the orders waiting for a manager.
type PendingPurchaseOrders struct {
	Pending int `json:"pending"`
	Drafts  int `json:"drafts"`
}

type ClosePurchaseOrderRequest struct {
	Version int    `json:"version"`
	Reason  string `json:"reason"`
}

// AuditEntry is one fact to append to the audit log.
type AuditEntry struct {
	ActorID    int
	EntityType string
	EntityID   int
	Action     string
	Reason     string
	Details    json.RawMessage
}

// AuditRecord is a stored audit entry with the actor's name resolved.
type AuditRecord struct {
	ID         int             `json:"id" db:"id"`
	OccurredAt time.Time       `json:"occurred_at" db:"occurred_at"`
	Actor      string          `json:"actor" db:"actor"`
	Action     string          `json:"action" db:"action"`
	Reason     string          `json:"reason" db:"reason"`
	Details    json.RawMessage `json:"details" db:"details"`
}

// PurchaseOrderRequest creates a purchase order (ID 0) or edits an open or partial one.
// On edit, items with ID 0 are new lines and existing lines left out are removed.
type PurchaseOrderRequest struct {
	ID       int                 `json:"id"`
	Version  int                 `json:"version"`
	Supplier string              `json:"supplier"`
	Note     string              `json:"note"`
	Items    []PurchaseOrderItem `json:"items"`
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
	Stock        *int                 `json:"stock"`
	CostPrice    float64              `json:"cost_price"`
	Prices       []ProductPriceUpdate `json:"prices"`
}
