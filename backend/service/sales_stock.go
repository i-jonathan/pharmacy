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

// checkSaleStock locks the sale's products in id order with LockProductStockTx (as stock
// adjustments do; receiving updates product rows in the same order, so neither waits on the other
// in a cycle), reads their stock, and refuses a sale that takes any of them below zero: 409
// `insufficient_stock`, or 403 when the user asked to oversell without sale:oversell. A permitted
// oversell returns the products it takes below zero. One product can be on several lines (single
// and pack prices), so what each line takes is summed per product.
func (s *saleService) checkSaleStock(
	ctx context.Context, tx *sqlx.Tx, items []model.SaleItem, allowOversell bool,
) ([]oversoldProduct, error) {
	taking := map[int]int{}
	ids := make([]int, 0, len(items))
	for _, item := range items {
		if _, seen := taking[item.ProductID]; !seen {
			ids = append(ids, item.ProductID)
		}
		taking[item.ProductID] += item.Quantity * item.StockQuantityPerUnit
	}
	sort.Ints(ids)

	stock, err := s.repo.LockProductStockTx(ctx, tx, ids)
	if err != nil {
		log.Println(err)
		return nil, httperror.ServerError("failed to check stock", err)
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
			parts[i] = fmt.Sprintf("%s (%d in stock, selling %d)", p.Name, p.StockBefore, p.Sold)
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
	err := appendAudit(ctx, s.repo, tx, userID, "sale", saleID, "oversold", "", map[string]any{
		"receipt_number": receipt,
		"products":       oversold,
	})
	if err != nil {
		return err
	}
	for _, p := range oversold {
		note := fmt.Sprintf(
			"Sold %d with %d in stock (receipt %s). Count it and correct the stock.",
			p.Sold, p.StockBefore, receipt,
		)
		if err := s.repo.OpenSaleStockFlagTx(ctx, tx, p.ProductID, userID, note); err != nil {
			log.Println(err)
			return httperror.ServerError("failed to flag the oversold stock", err)
		}
	}
	return nil
}
