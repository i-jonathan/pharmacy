package repository

import (
	"context"
	"pharmacy/model"

	"github.com/jmoiron/sqlx"
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
