package service

import (
	"context"
	"log"
	"pharmacy/httperror"
	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"pharmacy/model"
	"pharmacy/repository"
	"strings"
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
		Name:             constant.DefaultPriceName,
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
		ID:           productID,
		Name:         product.Name,
		Manufacturer: *product.Manufacturer,
		Barcode:      *product.Barcode,
	}, nil
}

func (s *inventoryService) FetchCategories(ctx context.Context) ([]types.CategoriesResponse, error) {
	categories, err := s.repo.FetchProductCategories(ctx)
	if err != nil {
		log.Println(err)
		return nil, httperror.ServerError("failed to fetch categories", err)
	}

	response := make([]types.CategoriesResponse, len(categories))
	for i, value := range categories {
		response[i] = types.CategoriesResponse{
			ID:   value.ID,
			Name: value.Name,
		}
	}
	return response, nil
}

func (s *inventoryService) SearchProducts(ctx context.Context, query string) ([]types.ProductResult, error) {
	products, err := s.repo.SearchProductByName(ctx, strings.ToLower(query))
	if err != nil {
		log.Println(err)
		return nil, httperror.ServerError("failed to execute search", err)
	}

	response := make([]types.ProductResult, len(products))
	for i, value := range products {
		priceOptions := make([]types.ProductPriceResult, len(value.PriceOptions))
		for j, po := range value.PriceOptions {
			priceOptions[j] = types.ProductPriceResult{
				ID:           po.ID,
				Name:         po.Name,
				SellingPrice: po.SellingPriceFloat(),
				Quantity:     po.QuantityPerUnit,
			}
		}

		response[i] = types.ProductResult{
			ID:           value.ID,
			Name:         value.Name,
			Manufacturer: *value.Manufacturer,
			Barcode:      *value.Barcode,
			CostPrice:    value.CostPriceFloat(),
			DefaultPrice: types.ProductPriceResult{
				ID:           value.DefaultPrice.ID,
				SellingPrice: value.DefaultPrice.SellingPriceFloat(),
			},
			PriceOptions: priceOptions,
		}
	}
	return response, nil
}

func (s *inventoryService) SearchForSuppliers(ctx context.Context, query string) ([]string, error) {
	suppliers, err := s.repo.SearchSuppliersName(ctx, query)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return suppliers, nil
}

func (s *inventoryService) ReceiveProductSupply(ctx context.Context, params types.ReceiveSupplyRequest) error {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		log.Println(err)
		return httperror.ServerError("failed to start transaction", err)
	}

	receivingBatch := model.ReceivingBatch{
		SupplierName: params.Supplier,
		ReceviedByID: params.UserID,
	}

	receivingBatchID, err := s.repo.CreateReceivingBatchTx(ctx, tx, receivingBatch)
	if err != nil {
		log.Println(err)
		return httperror.ServerError("failed to create receiving batch transaction", err)
	}

	var updatePriceData []map[string]any
	productBatch := make([]model.ProductBatch, len(params.Products))
	for i, value := range params.Products {
		priceID, err := s.repo.FetchDefaultPriceID(ctx, value.ID)
		if err != nil {
			log.Println(err)
			return httperror.ServerError("Failed to fetch default price id", err)
		}

		productBatch[i] = model.ProductBatch{
			ProductID:        value.ID,
			PriceID:          priceID,
			ReceivingBatchID: receivingBatchID,
			Quantity:         value.Quantity,
			CostPriceKobo:    int(value.CostPrice * 100),
			ExpiryDate:       &value.Expiry,
		}
		updatePriceData = append(updatePriceData, map[string]any{
			"product_id":    value.ID,
			"cost_price":    int(value.CostPrice * 100),
			"selling_price": int(value.SellingPrice * 100),
		})
	}

	err = s.repo.BulkUpdateProductPricesTx(ctx, tx, updatePriceData)
	if err != nil {
		log.Println(err)
		return httperror.ServerError("Failed to update product prices", err)
	}

	stockData, err := s.repo.BulkCreateProductBatchTx(ctx, tx, productBatch)
	if err != nil {
		log.Println(err)
		return httperror.ServerError("Failed to create product batch", err)
	}

	stockMovements := make([]model.StockMovement, len(stockData))
	for i, value := range stockData {
		stockMovements[i] = model.StockMovement{
			ProductID:    value.ProductID,
			Quantity:     value.Quantity,
			ReferenceID:  value.ID,
			MovementType: model.MovementTypeInPurchase,
		}
	}

	err = s.repo.BulkCreateStockMovementTx(ctx, tx, stockMovements)
	if err != nil {
		log.Println(err)
		return httperror.ServerError("Failed to bulk create stock movement", err)
	}

	err = s.repo.CommitTx(tx)
	if err != nil {
		log.Println(err)
		return httperror.ServerError("Failed to commit receiving products transaction", err)
	}
	return nil
}

func (s *inventoryService) FetchInventory(ctx context.Context) (*model.Inventory, error) {
	categories, err := s.repo.FetchProductCategories(ctx)
	if err != nil {
		log.Println(err)
		return nil, httperror.ServerError("failed to fetch categories", err)
	}

	inventoryItems, err := s.repo.FetchInventoryItems(ctx)
	if err != nil {
		log.Println(err)
		return nil, httperror.ServerError("failed to fetch inventory items", err)
	}

	return &model.Inventory{
		Items:      inventoryItems,
		Categories: categories,
	}, nil
}

func (s *inventoryService) UpdateProduct(ctx context.Context, id int, params types.UpdateProductRequest) error {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		log.Println(err)
		return httperror.ServerError("could not begin transaction", err)
	}

	product := model.Product{
		Name:          params.Name,
		Barcode:       &params.Barcode,
		CategoryID:    params.CategoryID,
		ReorderLevel:  params.ReorderLevel,
		Manufacturer:  &params.Manufacturer,
		CostPriceKobo: int(params.CostPrice * 100),
	}
	product.ID = id

	if err := s.repo.UpdateProductTx(ctx, tx, product); err != nil {
		log.Println(err)
		s.repo.RollbackTx(tx)
		return httperror.ServerError("failed to update product", err)
	}

	// Fetch current price IDs to identify which ones to delete
	currentProduct, err := s.repo.FetchProductByIDWithPrices(ctx, id)
	if err != nil {
		log.Println(err)
		s.repo.RollbackTx(tx)
		return httperror.ServerError("failed to fetch existing prices", err)
	}

	existingPriceMap := make(map[int]bool)
	for _, p := range currentProduct.PriceOptions {
		existingPriceMap[p.ID] = true
	}

	var defaultPriceID int
	updatedPriceIDs := make(map[int]bool)

	for _, p := range params.Prices {
		price := model.ProductPrice{
			ProductID:        id,
			Name:             p.Name,
			QuantityPerUnit:  p.QuantityPerUnit,
			SellingPriceKobo: int(p.SellingPrice * 100),
		}

		var priceID int
		if p.ID != nil {
			price.ID = *p.ID
			if err := s.repo.UpdateProductPriceTx(ctx, tx, price); err != nil {
				log.Println(err)
				s.repo.RollbackTx(tx)
				return httperror.ServerError("failed to update price", err)
			}
			priceID = *p.ID
			updatedPriceIDs[priceID] = true
		} else {
			id, err := s.repo.CreateProductPriceTx(ctx, tx, price)
			if err != nil {
				log.Println(err)
				s.repo.RollbackTx(tx)
				return httperror.ServerError("failed to create new price", err)
			}
			priceID = id
		}

		if p.IsDefault {
			defaultPriceID = priceID
		}
	}

	// Delete prices that were not in the update request
	for _, p := range currentProduct.PriceOptions {
		if !updatedPriceIDs[p.ID] && !containsPrice(params.Prices, p.ID) {
			if err := s.repo.DeleteProductPriceTx(ctx, tx, p.ID); err != nil {
				log.Println(err)
				s.repo.RollbackTx(tx)
				return httperror.ServerError("failed to delete removed price", err)
			}
		}
	}

	if defaultPriceID != 0 {
		if err := s.repo.UpdateProductDefaultPriceTx(ctx, tx, defaultPriceID, id); err != nil {
			log.Println(err)
			s.repo.RollbackTx(tx)
			return httperror.ServerError("failed to update default price", err)
		}
	}

	// Handle stock adjustment
	stockDiff := params.Stock - currentProduct.Stock
	if stockDiff != 0 {
		movementType := model.MovementTypeInManualAdjustment
		absDiff := stockDiff
		if stockDiff < 0 {
			movementType = model.MovementTypeOutManualAdjustment
			absDiff = -stockDiff
		}

		movement := model.StockMovement{
			ProductID:    id,
			Quantity:     absDiff,
			MovementType: movementType,
			ReferenceID:  id, // Self-referencing product ID for manual adjustments
		}

		if err := s.repo.CreateStockMovementTx(ctx, tx, movement); err != nil {
			log.Println(err)
			s.repo.RollbackTx(tx)
			return httperror.ServerError("failed to record stock adjustment", err)
		}
	}

	return s.repo.CommitTx(tx)
}

func (s *inventoryService) FetchProductByID(ctx context.Context, id int) (types.ProductResult, error) {
	product, err := s.repo.FetchProductByIDWithPrices(ctx, id)
	if err != nil {
		return types.ProductResult{}, httperror.NotFound("product not found", err)
	}

	priceOptions := make([]types.ProductPriceResult, len(product.PriceOptions))
	for i, po := range product.PriceOptions {
		priceOptions[i] = types.ProductPriceResult{
			ID:           po.ID,
			Name:         po.Name,
			SellingPrice: po.SellingPriceFloat(),
			Quantity:     po.QuantityPerUnit,
		}
	}

	return types.ProductResult{
		ID:           product.ID,
		Name:         product.Name,
		Manufacturer: *product.Manufacturer,
		Barcode:      *product.Barcode,
		CostPrice:    product.CostPriceFloat(),
		CategoryID:   product.CategoryID,
		ReorderLevel: product.ReorderLevel,
		Stock:        product.Stock,
		DefaultPrice: types.ProductPriceResult{
			ID:           product.DefaultPrice.ID,
			Name:         product.DefaultPrice.Name,
			SellingPrice: product.DefaultPrice.SellingPriceFloat(),
			Quantity:     product.DefaultPrice.QuantityPerUnit,
		},
		PriceOptions: priceOptions,
	}, nil
}

func containsPrice(prices []types.ProductPriceUpdate, id int) bool {
	for _, p := range prices {
		if p.ID != nil && *p.ID == id {
			return true
		}
	}
	return false
}
