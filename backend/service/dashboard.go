package service

import (
	"context"
	"log"
	"time"

	"pharmacy/httperror"
	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"pharmacy/model"
	"pharmacy/repository"
)

type dashboardService struct {
	repo repository.DashboardRepository
}

func NewDashboardService(repo repository.DashboardRepository) DashboardService {
	return &dashboardService{
		repo: repo,
	}
}

func (s *dashboardService) GetDashboardData(ctx context.Context, startDate, endDate *time.Time) (*types.DashboardResponse, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Use provided dates or default to today
	sd := todayStart
	ed := now
	prevStart := time.Now().AddDate(0, 0, -1).Truncate(24 * time.Hour)
	prevEnd := todayStart

	if startDate != nil {
		sd = *startDate
	}
	if endDate != nil {
		ed = *endDate
	}
	if startDate != nil && endDate != nil {
		// For custom ranges, previous period is same length
		duration := ed.Sub(sd)
		prevEnd = sd
		prevStart = prevEnd.Add(-duration)
	}

	// Check if user has sales total permission
	hasSalesPermission := HasPermission(ctx, constant.ViewSalesTotalPermissionKey)

	// Get KPI data
	var todaySales int
	var yesterdaySales int
	var salesTrend []types.SalesTrendData

	if hasSalesPermission {
		var err error
		todaySales, err = s.repo.GetTotalSales(ctx, sd, ed)
		if err != nil {
			log.Printf("failed to get today's sales: %v", err)
			return nil, httperror.ServerError("failed to get today's sales", err)
		}

		yesterdaySales, err = s.repo.GetTotalSales(ctx, prevStart, prevEnd)
		if err != nil {
			log.Printf("failed to get yesterday's sales: %v", err)
			return nil, httperror.ServerError("failed to get yesterday's sales", err)
		}

		// Get sales trend over the requested period
		days := int(ed.Sub(sd).Hours() / 24)
		if days < 1 {
			days = 1
		}
		if days > 31 {
			days = 31
		}

		for i := days; i >= 0; i-- {
			date := time.Now().AddDate(0, 0, -i)
			dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
			dayEnd := dayStart.Add(24 * time.Hour)

			sales, err := s.repo.GetTotalSales(ctx, dayStart, dayEnd)
			if err != nil {
				log.Printf("failed to get sales for day %d: %v", i, err)
				return nil, err
			}

			salesTrend = append(salesTrend, types.SalesTrendData{
				Date:  date.Format("Jan 2"),
				Sales: sales,
			})
		}
	}

	// Get other KPI data
	todayTransactions, err := s.repo.GetTransactionCount(ctx, sd, ed)
	if err != nil {
		log.Printf("failed to get today's transactions: %v", err)
		return nil, httperror.ServerError("failed to get today's transactions", err)
	}

	yesterdayTransactions, err := s.repo.GetTransactionCount(ctx, prevStart, prevEnd)
	if err != nil {
		log.Printf("failed to get yesterday's transactions: %v", err)
		return nil, httperror.ServerError("failed to get yesterday's transactions", err)
	}

	totalInventory, err := s.repo.GetTotalInventoryItems(ctx)
	if err != nil {
		log.Printf("failed to get total inventory: %v", err)
		return nil, httperror.ServerError("failed to get total inventory", err)
	}

	lowStockCount, err := s.repo.GetLowStockCount(ctx)
	if err != nil {
		log.Printf("failed to get low stock count: %v", err)
		return nil, httperror.ServerError("failed to get low stock count", err)
	}

	// Get category sales
	categorySales, err := s.repo.GetSalesByCategory(ctx, sd, ed)
	if err != nil {
		log.Printf("failed to get sales by category: %v", err)
		return nil, httperror.ServerError("failed to get category sales", err)
	}

	var categorySalesData []types.CategorySalesData
	totalSales := 0
	for _, sale := range categorySales {
		totalSales += sale.Sales
	}

	for _, sale := range categorySales {
		percentage := 0
		if totalSales > 0 {
			percentage = int((float64(sale.Sales) / float64(totalSales)) * 100)
		}
		categorySalesData = append(categorySalesData, types.CategorySalesData{
			Category: sale.Category,
			Sales:    percentage,
		})
	}

	// Get expiring items
	expiringItems, err := s.repo.GetExpiringItems(ctx, time.Now(), time.Now().AddDate(0, 0, 90))
	if err != nil {
		log.Printf("failed to get expiring items: %v", err)
		return nil, httperror.ServerError("failed to get expiring items", err)
	}

	// Get expiring items grouped by category with cost values
	expiryByCategory, err := s.repo.GetExpiringItemsByCategory(ctx, time.Now(), time.Now().AddDate(0, 0, 90))
	if err != nil {
		log.Printf("failed to get expiring items by category: %v", err)
		return nil, httperror.ServerError("failed to get expiring items by category", err)
	}

	// Get top selling products for the past 7 days
	weekStart := now.AddDate(0, 0, -7)
	topSelling, err := s.repo.GetTopSellingProducts(ctx, weekStart, now, 5)
	if err != nil {
		log.Printf("failed to get top selling products: %v", err)
		return nil, httperror.ServerError("failed to get top selling products", err)
	}

	// Get low stock items
	lowStockItems, err := s.repo.GetLowStockItems(ctx)
	if err != nil {
		log.Printf("failed to get low stock items: %v", err)
		return nil, httperror.ServerError("failed to get low stock items", err)
	}

	return &types.DashboardResponse{
		KPI: types.KPIResponse{
			TodaySales:        todaySales,
			TodayTransactions: todayTransactions,
			TotalInventory:    totalInventory,
			LowStockCount:     lowStockCount,
			SalesTrend:        calculatePercentageChange(float64(yesterdaySales), float64(todaySales)),
			TransactionTrend:  calculatePercentageChange(float64(yesterdayTransactions), float64(todayTransactions)),
		},
		SalesTrend:         salesTrend,
		CategorySales:      categorySalesData,
		TopSellingProducts: convertTopSellingProducts(topSelling),
		ExpiringItems:      convertExpiringItems(expiringItems),
		ExpiryByCategory:   convertExpiryByCategory(expiryByCategory),
		LowStockItems:      convertLowStockItems(lowStockItems),
	}, nil
}

func calculatePercentageChange(oldValue, newValue float64) float64 {
	if oldValue == 0 {
		if newValue == 0 {
			return 0
		}
		return 100
	}
	return ((newValue - oldValue) / oldValue) * 100
}

func convertExpiringItems(items []model.ExpiringItem) []types.ExpiringItemData {
	var result []types.ExpiringItemData
	for _, item := range items {
		result = append(result, types.ExpiringItemData{
			ID:              item.ProductID,
			ProductName:     item.ProductName,
			Quantity:        item.Quantity,
			CostPriceKobo:   item.CostPriceKobo,
			ExpiryDate:      item.ExpiryDate,
			DaysUntilExpiry: item.DaysUntilExpiry,
		})
	}
	return result
}

func convertExpiryByCategory(items []model.ExpiryByCategory) []types.ExpiryByCategoryData {
	var result []types.ExpiryByCategoryData
	for _, item := range items {
		result = append(result, types.ExpiryByCategoryData{
			Category:      item.Category,
			Count:         item.Count,
			TotalCostKobo: item.TotalCostKobo,
		})
	}
	return result
}

func convertTopSellingProducts(items []model.TopSellingProduct) []types.TopSellingProductData {
	var result []types.TopSellingProductData
	for _, item := range items {
		result = append(result, types.TopSellingProductData{
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			RevenueKobo: item.RevenueKobo,
		})
	}
	return result
}

func convertLowStockItems(items []model.LowStockItem) []types.LowStockItemData {
	var result []types.LowStockItemData
	for _, item := range items {
		result = append(result, types.LowStockItemData{
			ID:           item.ProductID,
			ProductName:  item.ProductName,
			Manufacturer: item.Manufacturer,
			CurrentStock: item.CurrentStock,
			ReorderLevel: item.ReorderLevel,
		})
	}
	return result
}
