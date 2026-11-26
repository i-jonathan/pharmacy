package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"

	"pharmacy/httperror"
	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"pharmacy/model"

	"github.com/jmoiron/sqlx"
)

// oversoldProduct is a product a sale took below zero, in stock units.
type oversoldProduct struct {
	ProductID   int    `json:"product_id"`
	Name        string `json:"-"`
	StockBefore int    `json:"stock_before"`
	Sold        int    `json:"sold"`
}

// lockSaleProducts locks the products of a new sale in id order with LockProductStockTx (as stock
// adjustments do; receiving and product edits update product rows, which waits on this lock too,
// and receiving does so in the same order) and returns their stock. It is taken before the sale is
// priced, so a price edit holding the lock commits first and is priced against.
func (s *saleService) lockSaleProducts(
	ctx context.Context, tx *sqlx.Tx, items []types.SaleItem,
) ([]types.ProductStock, error) {
	seen := map[int]bool{}
	ids := make([]int, 0, len(items))
	for _, item := range items {
		if item.ProductID > 0 && !seen[item.ProductID] {
			seen[item.ProductID] = true
			ids = append(ids, item.ProductID)
		}
	}
	if len(ids) == 0 {
		return nil, nil // pricing refuses the sale
	}
	sort.Ints(ids)
	stock, err := s.repo.LockProductStockTx(ctx, tx, ids)
	if err != nil {
		log.Println(err)
		return nil, httperror.ServerError("failed to check stock", err)
	}
	return stock, nil
}

// units is a count of stock units for a message: "1 unit", "10 units". Stock is kept in the
// smallest unit sold, so a pack line takes several.
func units(n int) string {
	if n == 1 {
		return fmt.Sprintf("%d unit", n)
	}
	return fmt.Sprintf("%d units", n)
}

// checkSaleStock refuses a sale that takes any locked product below zero: 409
// `insufficient_stock`, or 403 when the user asked to oversell without sale:oversell. A permitted
// oversell returns the products it takes below zero. One product can be on several lines (single
// and pack prices), so what each line takes is summed per product.
func checkSaleStock(
	ctx context.Context, stock []types.ProductStock, items []model.SaleItem, allowOversell bool,
) ([]oversoldProduct, error) {
	taking := map[int]int{}
	for _, item := range items {
		taking[item.ProductID] += item.Quantity * item.StockQuantityPerUnit
	}

	var short []oversoldProduct
	for _, p := range stock {
		if p.Stock-taking[p.ID] < 0 {
			short = append(short, oversoldProduct{
				ProductID: p.ID, Name: p.Name, StockBefore: p.Stock, Sold: taking[p.ID],
			})
		}
	}
	if len(short) == 0 {
		return nil, nil
	}
	if !allowOversell {
		parts := make([]string, len(short))
		for i, p := range short {
			parts[i] = fmt.Sprintf(
				"%s (%s in stock, selling %s)", p.Name, units(p.StockBefore), units(p.Sold),
			)
		}
		refusal := httperror.Conflict(
			"Not enough stock: "+strings.Join(parts, "; ")+".",
			errors.New("sale takes stock below zero"),
		)
		refusal.Reason = "insufficient_stock"
		return nil, refusal
	}
	if !HasPermission(ctx, constant.OversellPermissionKey) {
		return nil, httperror.Forbidden("you are not allowed to sell more than is in stock", nil)
	}
	return short, nil
}

// recordOversell writes the audit entry for a permitted oversell and opens a stock flag, raised
// by the seller, on each product it took below zero, so someone counts it. A flag this user
// already has open on the product gets the note added instead.
func (s *saleService) recordOversell(
	ctx context.Context, tx *sqlx.Tx, userID, saleID int, receipt string,
	oversold []oversoldProduct,
) error {
	err := appendAudit(ctx, s.repo, tx, userID, constant.AuditEntitySale, saleID,
		constant.AuditActionSaleOversold, "", map[string]any{
			"receipt_number": receipt,
			"products":       oversold,
		})
	if err != nil {
		return err
	}
	for _, p := range oversold {
		note := fmt.Sprintf(
			"Sold %s with %s in stock (receipt %s). Count it and correct the stock.",
			units(p.Sold), units(p.StockBefore), receipt,
		)
		if err := s.repo.OpenSaleStockFlagTx(ctx, tx, p.ProductID, userID, note); err != nil {
			log.Println(err)
			return httperror.ServerError("failed to flag the oversold stock", err)
		}
	}
	return nil
}
