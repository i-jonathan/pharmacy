package repository

import (
	"context"
	"pharmacy/internal/types"
	"pharmacy/model"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

func (r *repo) CreateSaleTx(ctx context.Context, tx *sqlx.Tx, sale model.Sale) (int, error) {
	var id int

	err := tx.QueryRowContext(
		ctx, createSaleQuery, sale.ReceiptNumber, sale.Status,
		sale.CashierID, sale.Subtotal, sale.Discount, sale.Total,
	).Scan(&id)

	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *repo) BulkCreateSaleItemsTx(ctx context.Context, tx *sqlx.Tx, saleItems []model.SaleItem) error {
	_, err := tx.NamedExecContext(ctx, bulkCreateSaleItemQuery, saleItems)
	if err != nil {
		return err
	}
	return nil
}

func (r *repo) BulkCreateSalePaymentsTX(ctx context.Context, tx *sqlx.Tx, salePayments []model.SalePayment) error {
	_, err := tx.NamedExecContext(ctx, bulkCreateSalePaymentQuery, salePayments)
	if err != nil {
		return err
	}
	return nil
}

func (r *repo) FetchSalesTx(ctx context.Context, tx *sqlx.Tx, filter types.SaleFilter) ([]model.Sale, error) {
	var sales []model.Sale
	var startDate, endDate any
	if filter.StartDate != nil {
		startDate = *filter.StartDate
	}
	if filter.EndDate != nil {
		endDate = *filter.EndDate
	}

	err := tx.SelectContext(ctx, &sales, fetchSalesQuery, startDate, endDate)
	if err != nil {
		return nil, err
	}
	return sales, nil
}

func (r *repo) BulkFetchSaleItems(ctx context.Context, tx *sqlx.Tx, saleIDs []int) ([]model.SaleItem, error) {
	var items []model.SaleItem

	err := tx.SelectContext(ctx, &items, bulkFetchSaleItemsQuery, pq.Array(saleIDs))
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (r *repo) BulkFetchSalePayments(ctx context.Context, tx *sqlx.Tx, saleIDs []int) ([]model.SalePayment, error) {
	var payments []model.SalePayment

	err := tx.SelectContext(ctx, &payments, bulkFetchSalePaymentsQuery, pq.Array(saleIDs))
	if err != nil {
		return nil, err
	}

	return payments, nil
}

func (r *repo) StoreHoldSale(ctx context.Context, saleTransaction model.HeldTransaction) error {
	_, err := r.Data.ExecContext(
		ctx, insertIntoHeldTransactionQuery, saleTransaction.Type,
		saleTransaction.Reference, saleTransaction.Payload,
	)
	return err
}
