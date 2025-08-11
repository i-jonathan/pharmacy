package service

import (
	"context"
	"log"
	"pharmacy/httperror"
	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"pharmacy/model"
	"pharmacy/repository"
)

type inventoryService struct {
	repo repository.PharmacyRepository
}

func NewInventoryService(repo repository.PharmacyRepository) *inventoryService {
	return &inventoryService{repo: repo}
}

func (s *inventoryService) CreateProduct(ctx context.Context, params types.CreateProductRequest) (types.AddItemResponse, error) {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		log.Println(err)
		return types.AddItemResponse{}, httperror.ServerError("could not begin transaction", err)
	}

	product := model.Product{
		Name:          params.Name,
		Barcode:       &params.Barcode,
		CategoryID:    params.CategoryID,
		ReorderLevel:  params.ReorderLevel,
		Manufacturer:  &params.Manufacturer,
		CostPriceKobo: int(params.CostPrice * 100),
	}
	productID, err := s.repo.CreateProductTx(ctx, tx, product)
	if err != nil {
		log.Println(err)
		return types.AddItemResponse{}, httperror.ServerError("failed to create product", err)
	}

	productPrice := model.ProductPrice{
		UnitName:         constant.DefaultPriceName,
		ProductID:        productID,
		QuantityPerUnit:  1,
		SellingPriceKobo: int(params.SellingPrice * 100),
	}
	priceID, err := s.repo.CreateProductPriceTx(ctx, tx, productPrice)
	if err != nil {
		log.Println(err)
		return types.AddItemResponse{}, httperror.ServerError("failed to create base price", err)
	}

	err = s.repo.UpdateProductDefaultPriceTx(ctx, tx, priceID, productID)
	if err != nil {
		log.Println(err)
		return types.AddItemResponse{}, httperror.ServerError("failed to update product default price", err)
	}

	err = s.repo.CommitTx(tx)
	if err != nil {
		log.Println(err)
		return types.AddItemResponse{}, httperror.ServerError("failed to commit transaction", err)
	}

	return types.AddItemResponse{
		ID:   productID,
		Name: product.Name,
	}, nil
}
