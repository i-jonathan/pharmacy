package service

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"pharmacy/httperror"
	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"pharmacy/model"
	"pharmacy/repository"
	"time"
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

		selectedPrice, err := s.repo.FetchPriceByID(ctx, v.PriceID)
		if err != nil {
			log.Println(err)
			return httperror.ServerError("failed to fetch product price", err)
		}

		stockMovement[i] = model.StockMovement{
			ProductID:    v.ProductID,
			Quantity:     v.Quantity * selectedPrice.QuantityPerUnit, // has to be quantity * selected price * quantity
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

func (s *saleService) FetchSalesHistory(ctx context.Context, filter types.SaleFilter) (types.SaleHistory, error) {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		log.Println(err)
		return types.SaleHistory{}, httperror.ServerError("Failed to begin transaction", err)
	}

	// fetch sales. Consider paginating
	sales, err := s.repo.FetchSalesTx(ctx, tx, filter)
	if err != nil {
		log.Println(err)
		return types.SaleHistory{}, httperror.ServerError("fetching sales failed", err)
	}

	// index sales by sale id and fetch all sale id's to use in getting sale items and payments
	salesByID := make(map[int]*model.Sale)
	saleIDs := make([]int, len(sales))
	cashierIDs := make(map[int]struct{})
	for i, s := range sales {
		salesByID[sales[i].ID] = &sales[i]
		saleIDs[i] = sales[i].ID
		cashierIDs[s.CashierID] = struct{}{}
	}

	// fetch sale items, from sale items, extract product id's.
	// Then, assign sale items to the already indexed sales map created earlier
	saleItems, err := s.repo.BulkFetchSaleItems(ctx, tx, saleIDs)
	if err != nil {
		log.Println(err)
		return types.SaleHistory{}, httperror.ServerError("fetching sale items failed", err)
	}

	productIDs := make([]int, len(saleItems))
	for i, item := range saleItems {
		salesByID[item.SaleID].SaleItems = append(salesByID[item.SaleID].SaleItems, item)
		productIDs[i] = item.ProductID
	}

	// fetch all sales payments, and assign to the already indexed sales. Again.
	payments, err := s.repo.BulkFetchSalePayments(ctx, tx, saleIDs)
	if err != nil {
		log.Println(err)
		return types.SaleHistory{}, httperror.ServerError("fetching sale payments failed", err)
	}

	for i := range payments {
		salesByID[payments[i].SaleID].Payments = append(salesByID[payments[i].SaleID].Payments, payments[i])
	}

	// fetch & index products to prepare for final structuring
	products, err := s.repo.BulkFetchProductByIDsTx(ctx, tx, productIDs)
	if err != nil {
		log.Println(err)
		return types.SaleHistory{}, httperror.ServerError("fetching products failed", err)
	}

	productsByID := make(map[int]model.Product)
	for _, p := range products {
		productsByID[p.ID] = p
	}

	ids := make([]int, 0, len(cashierIDs))
	for id := range cashierIDs {
		ids = append(ids, id)
	}

	cashiers, err := s.repo.BulkFetchUserByIDTx(ctx, tx, ids)
	if err != nil {
		log.Println(err)
		return types.SaleHistory{}, httperror.ServerError("fetching cashiers failed", err)
	}

	cashiersByID := make(map[int]string)
	for _, c := range cashiers {
		cashiersByID[c.ID] = c.UserName
	}

	responses := make([]types.SaleResponse, 0, len(sales))
	salesHistoryTotal := float64(0)
	for _, s := range sales {
		// build items
		items := make([]types.SaleItemResponse, 0, len(s.SaleItems))
		for _, item := range s.SaleItems {
			p := productsByID[item.ProductID]
			items = append(items, types.SaleItemResponse{
				ProductName:  p.Name,
				Manufacturer: *p.Manufacturer,
				Quantity:     item.Quantity,
				UnitPrice:    float64(item.UnitPrice) / 100,
				Discount:     float64(item.Discount) / 100,
			})
		}

		// build payments
		payments := make([]types.SalePaymentResponse, 0, len(s.Payments))
		for _, pay := range s.Payments {
			payments = append(payments, types.SalePaymentResponse{
				MethodName: pay.PaymentMethod,
				Amount:     float64(pay.Amount) / 100,
			})
		}

		// build the final response
		resp := types.SaleResponse{
			ReceiptNumber: s.ReceiptNumber,
			Cashier:       cashiersByID[s.CashierID],
			CreatedAt:     s.CreatedAt.Format(time.RFC3339),
			Subtotal:      float64(s.Subtotal) / 100,
			Discount:      float64(s.Discount) / 100,
			Total:         float64(s.Total) / 100,
			Items:         items,
			Payments:      payments,
		}
		responses = append(responses, resp)
		salesHistoryTotal += float64(s.Total) / 100
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		return types.SaleHistory{}, httperror.ServerError("failed to commit transaction", err)
	}

	return types.SaleHistory{
		TotalAmount: salesHistoryTotal,
		Data:        responses,
	}, nil
}

func (s *saleService) HoldSale(ctx context.Context, holdSaleRequest types.HoldTransactionRequest) error {
	reference := holdSaleRequest.Reference
	if reference == "" {
		reference = fmt.Sprintf("%s-%s-%04d",
			string(constant.HoldSaleType),
			time.Now().Format("20060102"),
			rand.Intn(10000),
		)
	}

	holdTransaction := model.HeldTransaction{
		Type:      string(constant.HoldSaleType),
		Reference: reference,
		Payload:   holdSaleRequest.Payload,
	}

	err := s.repo.SaveHeldTransaction(ctx, holdTransaction)
	if err != nil {
		log.Println(err)
		return httperror.ServerError("failed to hold sale", err)
	}
	return nil
}

func (s *saleService) FetchHeldSaleTransactions(ctx context.Context) ([]types.HeldTransactionResponse, error) {
	transactions, err := s.repo.FetchHeldTransactionsByType(ctx, constant.HoldSaleType)
	if err != nil {
		log.Println(err)
		return nil, httperror.ServerError("failed to fetch held transactions", err)
	}
	
	var response []types.HeldTransactionResponse
	for _, sale := range transactions {
		response = append(response, types.HeldTransactionResponse{
			Reference: sale.Reference,
			Payload:   sale.Payload,
			CreatedAt: sale.CreatedAt,
		})
	}
	
	return response, nil
}
