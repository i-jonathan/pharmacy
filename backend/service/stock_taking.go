package service

import (
	"context"
	"log"
	"pharmacy/httperror"
	"pharmacy/internal/types"
	"pharmacy/model"
	"pharmacy/repository"
	"strings"
	"time"
)

type stockTakingService struct {
	repo repository.StockTakingRepository
}

func NewStockTakingService(repo repository.StockTakingRepository) *stockTakingService {
	return &stockTakingService{
		repo: repo,
	}
}

func (s *stockTakingService) CreateStockTaking(ctx context.Context, data types.StockTakingData) (int, error) {
	active, err := s.repo.HasActiveStockTaking(ctx)
	if err != nil {
		log.Println(err)
		return 0, httperror.ServerError("failed checking active stock taking", err)
	}

	if active {
		return 0, httperror.BadRequest("another stock taking is currently in progress", nil)
	}

	name := strings.TrimSpace(data.Name)
	if name == "" {
		return 0, httperror.BadRequest("stock taking name is required", nil)
	}

	if len(name) > 100 {
		return 0, httperror.BadRequest("stock taking name is too long", nil)
	}

	if data.CreatedByID <= 0 {
		return 0, httperror.BadRequest("invalid creator", nil)
	}

	stockData := model.StockTaking{
		Name:        strings.TrimSpace(strings.ToTitle(data.Name)),
		CreatedByID: data.CreatedByID,
		Status:      model.StockTakingInProgress,
	}

	id, err := s.repo.CreateStockTaking(ctx, stockData)
	if err != nil {
		log.Println(err)
		return 0, httperror.ServerError("error occurred while creating stock", err)
	}

	return id, nil
}

func (s *stockTakingService) FetchStockTaking(ctx context.Context, stockTakingID int) (types.StockTakingData, error) {
	if stockTakingID <= 0 {
		return types.StockTakingData{}, httperror.BadRequest("invalid stock taking id", nil)
	}

	data, err := s.repo.GetStockTaking(ctx, stockTakingID)
	if err != nil {
		log.Println(err)
		return types.StockTakingData{}, httperror.ServerError("failed to fetch stock taking", err)
	}

	return types.StockTakingData{
		Name:        data.Name,
		Status:      data.Status.ToString(),
		CreatedBy:   data.CreatedByName,
		StartedAt:   data.StartedAt,
		CompletedAt: data.CompletedAt,
	}, nil
}

func (s *stockTakingService) FetchStockTakingItems(ctx context.Context, stockTakingID int) (types.StockTakingItems, error) {
	if stockTakingID <= 0 {
		return nil, httperror.BadRequest("invalid stock taking id", nil)
	}

	data, err := s.repo.GetStockTakingItems(ctx, stockTakingID)
	if err != nil {
		log.Println(err)
		return nil, httperror.ServerError("failed to fetch items", err)
	}

	return data, nil
}

func (s *stockTakingService) UpdateStockTakingItemCount(ctx context.Context, data types.StockTakingItemCount) error {
	item, err := s.repo.FetchStockTakingItem(ctx, data.StockTakingID, data.ProductID)
	if err != nil {
		log.Println(err)
		return httperror.NotFound("unable to find item for stock taking", err)
	}

	if item == nil {
		// create item with current stock level
		// fetch current stock level
		productStock, err := s.repo.FetchCurrentStockLevel(ctx, data.ProductID)
		if err != nil {
			log.Println(err)
			return httperror.ServerError("failed to fetch current stock level", err)
		}

		item = &model.StockTakingItem{
			StockTakingID:    data.StockTakingID,
			ProductID:        data.ProductID,
			SnapshotQuantity: productStock,
		}

		id, err := s.repo.CreateStockTakingItem(ctx, item)
		if err != nil {
			log.Println(err)
			return httperror.ServerError("failed to create stock taking item", err)
		}

		item.ID = id
	}

	if data.DispensaryCount != nil {
		item.DispensaryCount = *data.DispensaryCount
	}
	if data.StoreCount != nil {
		item.StoreCount = *data.StoreCount
	}
	if data.Notes != nil {
		item.Notes = *data.Notes
	}
	item.LastUpdatedByID = data.UpdatedByID

	if err := s.repo.UpdateStockTakingItem(ctx, item); err != nil {
		log.Println(err)
		return httperror.ServerError("failed to update stock taking item", err)
	}

	// update the expiry
	if data.UpdatedExpiry != nil {
		expiryTime, err := time.Parse("2006-01-02", *data.UpdatedExpiry)
		if err != nil {
			return httperror.BadRequest("invalid expiry date format", err)
		}

		if err := s.repo.UpdateProductCurrentExpiry(ctx, item.ProductID, &expiryTime); err != nil {
			return httperror.ServerError("failed to update expiry", err)
		}
	}
	return nil
}
