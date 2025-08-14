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
