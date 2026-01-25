package service

import (
	"context"
	"log"
	"pharmacy/httperror"
	"pharmacy/internal/types"
	"pharmacy/model"
	"pharmacy/repository"
	"strings"
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
