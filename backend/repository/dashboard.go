package repository

import (
	"context"
	"pharmacy/model"
	"time"
)

func (r *repo) GetTotalSales(ctx context.Context, startDate, endDate time.Time) (int, error) {
	query := `
		SELECT
			COALESCE(SUM(s.total), 0) - COALESCE(SUM(r.total_refunded), 0) as net_sales
		FROM sales s
		LEFT JOIN returns r ON s.id = r.sale_id
		WHERE s.created_at >= $1 AND s.created_at < $2
	`

	var netSales int
	err := r.Data.GetContext(ctx, &netSales, query, startDate, endDate)
	if err != nil {
		return 0, err
	}

	// Prevent negative sales
	if netSales < 0 {
		netSales = 0
	}

	return netSales, nil
}

func (r *repo) GetTransactionCount(ctx context.Context, startDate, endDate time.Time) (int, error) {
	query := `
		SELECT COUNT(*) as transaction_count
		FROM sales
		WHERE created_at >= $1 AND created_at < $2
	`

	var count int
	err := r.Data.GetContext(ctx, &count, query, startDate, endDate)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *repo) GetTotalInventoryItems(ctx context.Context) (int, error) {
	query := `
		SELECT COALESCE(SUM(stock), 0) as total_items
		FROM inventory_view
	`

	var totalItems int
	err := r.Data.GetContext(ctx, &totalItems, query)
	if err != nil {
		return 0, err
	}

	return totalItems, nil
}

func (r *repo) GetLowStockCount(ctx context.Context) (int, error) {
	query := `
		SELECT COUNT(*) as low_stock_count
		FROM inventory_view
		WHERE stock <= reorder_level
	`

	var count int
	err := r.Data.GetContext(ctx, &count, query)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *repo) GetLowStockItems(ctx context.Context) ([]model.LowStockItem, error) {
	query := `
		SELECT
			p.id as product_id,
			p.name as product_name,
			p.manufacturer as manufacturer,
			iv.stock as current_stock,
			iv.reorder_level as reorder_level
		FROM inventory_view iv
		JOIN product p ON iv.id = p.id
		WHERE iv.stock <= iv.reorder_level
		ORDER BY iv.stock ASC, p.name ASC
	`

	var lowStockItems []model.LowStockItem
	err := r.Data.SelectContext(ctx, &lowStockItems, query)
	if err != nil {
		return nil, err
	}

	return lowStockItems, nil
}

func (r *repo) GetSalesByCategory(ctx context.Context, startDate, endDate time.Time) ([]model.SalesByCategory, error) {
	query := `
		SELECT
			c.name as category,
			COUNT(si.id) as sales
		FROM sales s
		JOIN sales_item si ON s.id = si.sale_id
		JOIN product p ON si.product_id = p.id
		JOIN category c ON p.category_id = c.id
		WHERE s.created_at >= $1 AND s.created_at < $2
		GROUP BY c.name
		ORDER BY sales DESC
	`

	var salesByCategory []model.SalesByCategory
	err := r.Data.SelectContext(ctx, &salesByCategory, query, startDate, endDate)
	if err != nil {
		return nil, err
	}

	return salesByCategory, nil
}

func (r *repo) GetExpiringItems(ctx context.Context, startDate, endDate time.Time) ([]model.ExpiringItem, error) {
	query := `
		SELECT
			p.id as product_id,
			p.name as product_name,
			COALESCE(iv.stock, 0) as quantity,
			p.current_expiry as expiry_date,
			(p.current_expiry - CURRENT_DATE)::integer as days_until_expiry
		FROM product p
		LEFT JOIN inventory_view iv ON p.id = iv.id
		WHERE p.current_expiry IS NOT NULL
		  AND p.current_expiry >= CURRENT_DATE
		  AND p.current_expiry <= CURRENT_DATE + INTERVAL '90 days'
		  AND iv.stock > 0
		ORDER BY days_until_expiry ASC, p.current_expiry ASC
	`

	var expiringItems []model.ExpiringItem
	err := r.Data.SelectContext(ctx, &expiringItems, query)
	if err != nil {
		return nil, err
	}

	return expiringItems, nil
}

func (r *repo) GetExpiringItemsByCategory(ctx context.Context, startDate, endDate time.Time) ([]model.ExpiryByCategory, error) {
	query := `
		SELECT
			c.name as category,
			COUNT(DISTINCT p.id) as count,
			COALESCE(SUM(p.cost_price * COALESCE(iv.stock, 0)), 0) as total_cost_kobo
		FROM product p
		JOIN category c ON p.category_id = c.id
		LEFT JOIN inventory_view iv ON p.id = iv.id
		WHERE p.current_expiry IS NOT NULL
		  AND p.current_expiry >= $1
		  AND p.current_expiry <= $2
		  AND COALESCE(iv.stock, 0) > 0
		GROUP BY c.name
		ORDER BY total_cost_kobo DESC
	`

	var items []model.ExpiryByCategory
	err := r.Data.SelectContext(ctx, &items, query, startDate, endDate)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (r *repo) GetTopSellingProducts(ctx context.Context, startDate, endDate time.Time, limit int) ([]model.TopSellingProduct, error) {
	query := `
		SELECT
			p.name as product_name,
			SUM(si.quantity) as quantity,
			SUM(si.total_price) as revenue_kobo
		FROM sales s
		JOIN sales_item si ON s.id = si.sale_id
		JOIN product p ON si.product_id = p.id
		WHERE s.created_at >= $1 AND s.created_at < $2
		  AND s.status = 'COMPLETED'
		GROUP BY p.id, p.name
		ORDER BY quantity DESC
		LIMIT $3
	`

	var products []model.TopSellingProduct
	err := r.Data.SelectContext(ctx, &products, query, startDate, endDate, limit)
	if err != nil {
		return nil, err
	}

	return products, nil
}

func (r *repo) GetSalesByTime(ctx context.Context, startTime, endTime time.Time) ([]model.Sale, error) {
	query := `
		SELECT id, receipt_number, cashier_id, subtotal, discount, total, status, created_at
		FROM sales
		WHERE created_at >= $1 AND created_at < $2
		ORDER BY created_at DESC
	`

	var sales []model.Sale
	err := r.Data.SelectContext(ctx, &sales, query, startTime, endTime)
	if err != nil {
		return nil, err
	}

	return sales, nil
}
