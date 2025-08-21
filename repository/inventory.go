package repository

import (
	"context"
	"pharmacy/internal/types"
	"pharmacy/model"

	"github.com/jmoiron/sqlx"
)

func (r *repo) CreateProductTx(ctx context.Context, tx *sqlx.Tx, product model.Product) (int, error) {
	var id int

	err := tx.QueryRowContext(
		ctx, createProductQuery, product.Name, product.Barcode, product.CategoryID,
		product.ReorderLevel, product.Manufacturer, product.CostPriceKobo,
	).Scan(&id)

	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *repo) CreateProductPriceTx(ctx context.Context, tx *sqlx.Tx, productPrice model.ProductPrice) (int, error) {
	var id int
	err := tx.QueryRowContext(
		ctx, createProductPriceQuery, productPrice.ProductID,
		productPrice.QuantityPerUnit, productPrice.SellingPriceKobo, productPrice.Name,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *repo) UpdateProductDefaultPriceTx(ctx context.Context, tx *sqlx.Tx, price_id int, product_id int) error {
	_, err := tx.ExecContext(ctx, updateProductDefaultPrice, price_id, product_id)
	return err
}

func (r *repo) FetchProductCategories(ctx context.Context) ([]model.Category, error) {
	var categories []model.Category
	err := r.Data.SelectContext(ctx, &categories, fetchCategoriesQuery)
	if err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *repo) SearchProductByName(ctx context.Context, searchTerm string) ([]model.Product, error) {
	var products []model.Product

	err := r.Data.SelectContext(ctx, &products, searchProductsQuery, searchTerm)
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (r *repo) SearchSuppliersName(ctx context.Context, query string) ([]string, error) {
	var suppliers []string
	err := r.Data.SelectContext(ctx, &suppliers, searchSupplierQuery, query)
	if err != nil {
		return nil, err
	}
	return suppliers, nil
}

func (r *repo) CreateReceivingBatchTx(ctx context.Context, tx *sqlx.Tx, receivingBatch model.ReceivingBatch) (int, error) {
	var receivingBatchID int

	err := tx.QueryRowContext(ctx, createReceivingBatchQuery, receivingBatch.SupplierName, receivingBatch.ReceviedByID).Scan(&receivingBatchID)
	if err != nil {
		return 0, err
	}

	return receivingBatchID, nil
}

func (r *repo) BulkCreateProductBatchTx(_ context.Context, tx *sqlx.Tx, productBatches []model.ProductBatch) ([]types.BatchInsertReturn, error) {
	rows, err := tx.NamedQuery(createProductBatchQuery, productBatches)
	if err != nil {
		return nil, err
	}

	var batchReturn []types.BatchInsertReturn
	for rows.Next() {
		var batch types.BatchInsertReturn
		err = rows.StructScan(&batch)
		if err != nil {
			return nil, err
		}
		batchReturn = append(batchReturn, batch)
	}

	return batchReturn, nil
}

func (r *repo) BulkCreateStockMovementTx(ctx context.Context, tx *sqlx.Tx, stockMovements []model.StockMovement) error {
	_, err := tx.NamedExecContext(ctx, createMovementFromBatchQuery, stockMovements)
	if err != nil {
		return err
	}
	return nil
}

func (r *repo) FetchDefaultPriceID(ctx context.Context, productID int) (int, error) {
	var priceID int
	err := r.Data.GetContext(ctx, &priceID, fetchDefaultPriceIDQuery, productID)
	if err != nil {
		return 0, err
	}
	return priceID, nil
}

func (r *repo) BulkUpdateProductPricesTx(ctx context.Context, tx *sqlx.Tx, updateValues []map[string]any) error {
	_, err := tx.NamedExecContext(ctx, updateProductPricesQuery, updateValues)
	if err != nil {
		return err
	}
	return nil
}
