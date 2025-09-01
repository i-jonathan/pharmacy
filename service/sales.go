package service

import (
	"context"
	"log"
	"pharmacy/httperror"
	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"pharmacy/model"
	"pharmacy/repository"
)

type saleService struct {
	repo repository.PharmacyRepository
}

func NewSaleService(repo repository.PharmacyRepository) *saleService {
	return &saleService{repo: repo}
}

func (s *saleService) CreateSale(ctx context.Context, saleParams types.Sale) error {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		log.Println(err)
		return httperror.ServerError("could not begin transcation", err)
	}

	sale := model.Sale{
		Status:    constant.CompleteSaleStatus,
		CashierID: saleParams.CashierID,
		Discount:  int(saleParams.Discount * 100),
		Subtotal:  int(saleParams.Subtotal * 100),
		Total:     int(saleParams.Total * 100),
	}
	sale.GenerateReceiptNumber()

	saleID, err := s.repo.CreateSaleTx(ctx, tx, sale)
	if err != nil {
		log.Println(err)
		return httperror.ServerError("failed to create sale", err)
	}

	saleItems := make([]model.SaleItem, len(saleParams.Items))
	stockMovement := make([]model.StockMovement, len(saleParams.Items))
	for i, v := range saleParams.Items {
		saleItems[i] = model.SaleItem{
			SaleID:     saleID,
			ProductID:  v.ProductID,
			Quantity:   v.Quantity,
			UnitPrice:  int(v.UnitPrice * 100),
			Discount:   int(v.Discount * 100),
			TotalPrice: int(v.Total * 100),
		}

		stockMovement[i] = model.StockMovement{
			ProductID:    v.ProductID,
			Quantity:     v.Quantity,
			ReferenceID:  saleID,
			MovementType: constant.SaleMovementName,
		}
	}

	err = s.repo.BulkCreateSaleItemsTx(ctx, tx, saleItems)
	if err != nil {
		log.Println(err)
		return httperror.ServerError("Failed to create sale items", err)
	}

	salePayments := make([]model.SalePayment, len(saleParams.Payments))
	for i, v := range saleParams.Payments {
		salePayments[i] = model.SalePayment{
			SaleID:        saleID,
			PaymentMethod: constant.NormalizePaymentMethod(v.PaymentMethod),
			Amount:        int(v.Amount * 100),
		}
	}

	err = s.repo.BulkCreateSalePaymentsTX(ctx, tx, salePayments)
	if err != nil {
		log.Println(err)
		return httperror.ServerError("Failed to bulk create payments", err)
	}

	// todo create bulk stock movements
	err = s.repo.BulkCreateStockMovementTx(ctx, tx, stockMovement)
	if err != nil {
		log.Println(err)
		return httperror.ServerError("Failed to bulk create stock movements", err)
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
		return httperror.ServerError("failed to commit transaction", err)
	}
	return nil
}
