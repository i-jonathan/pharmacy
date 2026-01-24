package repository

import (
	"context"
	"pharmacy/internal/types"
	"pharmacy/model"
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
