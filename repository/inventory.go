package repository

import (
	"context"
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
		productPrice.QuantityPerUnit, productPrice.SellingPriceKobo,
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
