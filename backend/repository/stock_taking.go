package repository

import (
	"context"
	"database/sql"
	"errors"
	"pharmacy/internal/types"
	"pharmacy/model"
	"time"
)

func (r *repo) CreateStockTaking(ctx context.Context, stockTakingData model.StockTaking) (int, error) {
	var id int
	err := r.Data.QueryRowContext(
		ctx, createStockTakingQuery, stockTakingData.Name, stockTakingData.Status, stockTakingData.CreatedByID,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *repo) GetStockTaking(ctx context.Context, stockTakingID int) (model.StockTaking, error) {
	var stockTaking model.StockTaking
	err := r.Data.GetContext(ctx, &stockTaking, getStockTakingByIDQuery, stockTakingID)

	if err != nil {
		return model.StockTaking{}, err
	}

	return stockTaking, nil
}

func (r *repo) GetStockTakingItems(ctx context.Context, stockTakingID int) (types.StockTakingItems, error) {
	var stockTakingItems types.StockTakingItems

	err := r.Data.SelectContext(ctx, &stockTakingItems, getStockTakingItemsQuery, stockTakingID)
	if err != nil {
		return nil, err
	}

	return stockTakingItems, nil
}

func (r *repo) HasActiveStockTaking(ctx context.Context) (bool, error) {
	var exists bool
	err := r.Data.GetContext(ctx, &exists, checkIfActiveStockTaking, model.StockTakingInProgress)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *repo) FetchCurrentStockLevel(ctx context.Context, productID int) (int, error) {
	var stock int
	err := r.Data.GetContext(ctx, &stock, currentProductStockQuery, productID)
	if err != nil {
		return 0, err
	}

	return stock, nil
}

func (r *repo) FetchStockTakingItem(ctx context.Context, stockTakingID, productID int) (*model.StockTakingItem, error) {
	var item model.StockTakingItem
	err := r.Data.GetContext(ctx, &item, getStockTakingItemByProductIDQuery, stockTakingID, productID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *repo) CreateStockTakingItem(ctx context.Context, item *model.StockTakingItem) (int, error) {
	var id int
	err := r.Data.QueryRowContext(
		ctx, createStockTakingItemQuery, item.StockTakingID,
		item.ProductID, item.SnapshotQuantity,
	).Scan(&id)

	if err != nil {
		return 0, nil
	}

	return id, nil
}

func (r *repo) UpdateStockTakingItem(ctx context.Context, item *model.StockTakingItem) error {
	_, err := r.Data.ExecContext(
		ctx, updateStockTakingItemQuery, item.DispensaryCount,
		item.StoreCount, item.Notes, item.LastUpdatedByID, item.ID,
	)
	return err
}

func (r *repo) UpdateProductCurrentExpiry(ctx context.Context, productID int, currentExpiry *time.Time) error {
	_, err := r.Data.ExecContext(ctx, updateProductCurrentExpiry, productID, currentExpiry)
	return err
}
