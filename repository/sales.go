package repository

import (
	"context"
	"pharmacy/internal/constant"
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

func (r *repo) SaveHeldTransaction(ctx context.Context, transaction model.HeldTransaction) error {
	_, err := r.Data.ExecContext(
		ctx, upsertHeldTransactionQuery, transaction.Type,
		transaction.Reference, transaction.Payload,
	)
	return err
}

func (r *repo) FetchHeldTransactionsByType(ctx context.Context, transactionType constant.HoldTransactionType) ([]model.HeldTransaction, error) {
	var heldTransactions []model.HeldTransaction
	err := r.Data.SelectContext(ctx, &heldTransactions, fetchHeldTransactionByTypeQuery, string(transactionType))
	if err != nil {
		return nil, err
	}
	return heldTransactions, nil
}

func (r *repo) DeleteHeldTransactionByReference(ctx context.Context, reference string) error {
	_, err := r.Data.ExecContext(ctx, deleteHeldTransactionByReferenceQuery, reference)
	return err
}

func (r *repo) DeleteHeldTransactionByReferenceTx(ctx context.Context, tx *sqlx.Tx, reference string) error {
	_, err := tx.ExecContext(ctx, deleteHeldTransactionByReferenceQuery, reference)
	return err
}

func (r *repo) CreateReturnTx(ctx context.Context, tx *sqlx.Tx, rtn model.Return) (int, error) {
	var id int
	err := tx.QueryRowContext(
		ctx, createReturnQuery, rtn.SaleID, rtn.CashierID, rtn.TotalRefunded, rtn.Notes,
	).Scan(&id)

	if err != nil {
		return 0, nil
	}
	return id, nil
}

func (r *repo) BulkCreateReturnItemsTx(ctx context.Context, tx *sqlx.Tx, returnItems []model.ReturnItems) error {
	_, err := tx.NamedExecContext(ctx, bulkCreateReturnItemQuery, returnItems)
	if err != nil {
		return err
	}
	return nil
}

func (r *repo) FetchSaleByID(ctx context.Context, saleID int) (model.Sale, error) {
	var sale model.Sale

	err := r.Data.GetContext(ctx, &sale, fetchSalesByIDQuery, saleID)
	if err != nil {
		return model.Sale{}, err
	}

	var items []model.SaleItem
	err = r.Data.SelectContext(ctx, &items, fetchSaleItemsBySaleIDQuery, saleID)
	if err != nil {
		return model.Sale{}, err
	}

	sale.SaleItems = items
	return sale, nil
}

func (r *repo) FetchAllSaleReturns(ctx context.Context, saleID int) (model.ReturnItems, error) {
	var returnItems model.ReturnItems
	err := r.Data.GetContext(ctx, &returnItems, fetchReturnsForSaleBySaleIDQuery, saleID)
	if err != nil {
		return model.ReturnItems{}, err
	}
	return returnItems, nil
}

func (r *repo) BulkFetchReturnItemsBySaleIDs(ctx context.Context, saleIDs []int) ([]model.ReturnItemWithSale, error) {
	var returnItems []model.ReturnItemWithSale
	err := r.Data.SelectContext(ctx, &returnItems, bulkFetchReturnsForSaleBySaleIDQuery, saleIDs)
	if err != nil {
		return nil, err
	}
	return returnItems, nil
}
