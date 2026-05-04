package service

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"

	"pharmacy/internal/types"
	"pharmacy/model"
	"pharmacy/repository"
)

type duplicateSaleRepo struct {
	repository.PharmacyRepository
	t *testing.T

	beginCalls          int
	rollbackCalls       int
	createSaleCalls     int
	createSaleItemCalls int
	paymentCalls        int
	stockMovementCalls  int
}

func (r *duplicateSaleRepo) BeginTx(context.Context) (*sqlx.Tx, error) {
	r.beginCalls++
	return nil, nil
}

func (r *duplicateSaleRepo) RollbackTx(*sqlx.Tx) {
	r.rollbackCalls++
}

func (r *duplicateSaleRepo) CreateSaleTx(_ context.Context, _ *sqlx.Tx, sale model.Sale) (int, error) {
	r.createSaleCalls++
	if sale.IdempotencyKey == nil || *sale.IdempotencyKey != "sale-key-1" {
		r.t.Fatalf("expected normalized sale idempotency key, got %#v", sale.IdempotencyKey)
	}
	return 0, nil
}

func (r *duplicateSaleRepo) BulkCreateSaleItemsTx(context.Context, *sqlx.Tx, []model.SaleItem) error {
	r.createSaleItemCalls++
	r.t.Fatal("duplicate sale should not create sale items")
	return nil
}

func (r *duplicateSaleRepo) BulkCreateSalePaymentsTX(context.Context, *sqlx.Tx, []model.SalePayment) error {
	r.paymentCalls++
	r.t.Fatal("duplicate sale should not create payments")
	return nil
}

func (r *duplicateSaleRepo) BulkCreateStockMovementTx(context.Context, *sqlx.Tx, []model.StockMovement) error {
	r.stockMovementCalls++
	r.t.Fatal("duplicate sale should not create stock movements")
	return nil
}

func TestCreateSaleDuplicateIdempotencyKeyShortCircuits(t *testing.T) {
	repo := &duplicateSaleRepo{t: t}
	svc := NewSaleService(repo)

	err := svc.CreateSale(context.Background(), types.Sale{
		CashierID:      1,
		IdempotencyKey: "  sale-key-1  ",
		Subtotal:       100,
		Total:          100,
		Items: []types.SaleItem{
			{ProductID: 1, Quantity: 1, PriceID: 1, UnitPrice: 100, Total: 100},
		},
		Payments: []types.SalePayment{
			{Amount: 100, PaymentMethod: "Cash"},
		},
	})
	if err != nil {
		t.Fatalf("CreateSale returned error: %v", err)
	}

	if repo.beginCalls != 1 {
		t.Fatalf("expected BeginTx once, got %d", repo.beginCalls)
	}
	if repo.createSaleCalls != 1 {
		t.Fatalf("expected CreateSaleTx once, got %d", repo.createSaleCalls)
	}
	if repo.rollbackCalls != 1 {
		t.Fatalf("expected duplicate transaction rollback once, got %d", repo.rollbackCalls)
	}
}

type duplicateReceiveRepo struct {
	repository.PharmacyRepository
	t *testing.T

	beginCalls              int
	rollbackCalls           int
	createBatchCalls        int
	createProductBatchCalls int
	stockMovementCalls      int
	priceUpdateCalls        int
}

func (r *duplicateReceiveRepo) BeginTx(context.Context) (*sqlx.Tx, error) {
	r.beginCalls++
	return nil, nil
}

func (r *duplicateReceiveRepo) RollbackTx(*sqlx.Tx) {
	r.rollbackCalls++
}

func (r *duplicateReceiveRepo) CreateReceivingBatchTx(_ context.Context, _ *sqlx.Tx, batch model.ReceivingBatch) (int, error) {
	r.createBatchCalls++
	if batch.IdempotencyKey == nil || *batch.IdempotencyKey != "receive-key-1" {
		r.t.Fatalf("expected normalized receiving idempotency key, got %#v", batch.IdempotencyKey)
	}
	return 0, nil
}

func (r *duplicateReceiveRepo) BulkCreateProductBatchTx(context.Context, *sqlx.Tx, []model.ProductBatch) ([]types.BatchInsertReturn, error) {
	r.createProductBatchCalls++
	r.t.Fatal("duplicate receive should not create product batches")
	return nil, nil
}

func (r *duplicateReceiveRepo) BulkCreateStockMovementTx(context.Context, *sqlx.Tx, []model.StockMovement) error {
	r.stockMovementCalls++
	r.t.Fatal("duplicate receive should not create stock movements")
	return nil
}

func (r *duplicateReceiveRepo) BulkUpdateProductPricesTx(context.Context, *sqlx.Tx, []map[string]any) error {
	r.priceUpdateCalls++
	r.t.Fatal("duplicate receive should not update product prices")
	return nil
}

func TestReceiveProductSupplyDuplicateIdempotencyKeyShortCircuits(t *testing.T) {
	repo := &duplicateReceiveRepo{t: t}
	svc := NewInventoryService(repo)

	err := svc.ReceiveProductSupply(context.Background(), types.ReceiveSupplyRequest{
		Supplier:       "Supplier",
		UserID:         1,
		IdempotencyKey: "  receive-key-1  ",
		Products: []types.ReceiveItem{
			{ID: 1, CostPrice: 50, SellingPrice: 100, Quantity: 2},
		},
	})
	if err != nil {
		t.Fatalf("ReceiveProductSupply returned error: %v", err)
	}

	if repo.beginCalls != 1 {
		t.Fatalf("expected BeginTx once, got %d", repo.beginCalls)
	}
	if repo.createBatchCalls != 1 {
		t.Fatalf("expected CreateReceivingBatchTx once, got %d", repo.createBatchCalls)
	}
	if repo.rollbackCalls != 1 {
		t.Fatalf("expected duplicate transaction rollback once, got %d", repo.rollbackCalls)
	}
}
