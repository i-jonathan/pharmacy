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

func (s *dashboardService) GetDashboardData(ctx context.Context) (*types.DashboardResponse, error) {
	// Check if user has sales total permission
	hasSalesPermission := HasPermission(ctx, constant.ViewSalesTotalPermissionKey)

	// Get KPI data
	var todaySales int
	var yesterdaySales int
	var salesTrend []types.SalesTrendData

	if hasSalesPermission {
		// Only fetch sales data if user has permission
		var err error
		todaySales, err = s.repo.GetTotalSales(ctx, time.Now().Truncate(24*time.Hour), time.Now())
		if err != nil {
			log.Printf("failed to get today's sales: %v", err)
			return nil, httperror.ServerError("failed to get today's sales", err)
		}

		yesterdaySales, err = s.repo.GetTotalSales(ctx, time.Now().AddDate(0, 0, -1).Truncate(24*time.Hour), time.Now().AddDate(0, 0, -1))
		if err != nil {
			log.Printf("failed to get yesterday's sales: %v", err)
			return nil, httperror.ServerError("failed to get yesterday's sales", err)
		}

		// Get sales trend (last 7 days) - only if user has permission
		days := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}

		for i := 6; i >= 0; i-- {
			date := time.Now().AddDate(0, 0, -i)
			dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
			dayEnd := dayStart.Add(24 * time.Hour)

			sales, err := s.repo.GetTotalSales(ctx, dayStart, dayEnd)
			if err != nil {
				log.Printf("failed to get sales for day %d: %v", i, err)
				return nil, err
			}

			dayIndex := int(date.Weekday())
			if dayIndex == 0 { // Sunday is 0 in Go, but we want it last
				dayIndex = 6
			} else {
				dayIndex--
			}

			salesTrend = append(salesTrend, types.SalesTrendData{
				Day:   days[dayIndex],
				Sales: sales,
			})
		}
	}

	// Get other KPI data (these don't require sales permission)
	todayTransactions, err := s.repo.GetTransactionCount(ctx, time.Now().Truncate(24*time.Hour), time.Now())
	if err != nil {
		log.Printf("failed to get today's transactions: %v", err)
		return nil, httperror.ServerError("failed to get today's transactions", err)
	}

	yesterdayTransactions, err := s.repo.GetTransactionCount(ctx, time.Now().AddDate(0, 0, -1).Truncate(24*time.Hour), time.Now().AddDate(0, 0, -1))
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
	categorySales, err := s.repo.GetSalesByCategory(ctx, time.Now().AddDate(0, 0, -30), time.Now())
	if err != nil {
		log.Printf("failed to get sales by category: %v", err)
		return nil, httperror.ServerError("failed to get category sales", err)
	}

	// Convert to percentages and format for API
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

	// Get expiring items data
	expiringItems, err := s.repo.GetExpiringItems(ctx, time.Now(), time.Now().AddDate(0, 0, 90))
	if err != nil {
		log.Printf("failed to get expiring items: %v", err)
		return nil, httperror.ServerError("failed to get expiring items", err)
	}

	// Get low stock items and convert
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
		SalesTrend:    salesTrend,
		CategorySales: categorySalesData,
		ExpiringItems: convertExpiringItems(expiringItems),
		LowStockItems: convertLowStockItems(lowStockItems),
	}, nil
}

func calculatePercentageChange(oldValue, newValue float64) float64 {
	if oldValue == 0 {
		if newValue == 0 {
			return 0
		}
		return 100 // 100% increase from 0
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

func convertLowStockItems(items []model.LowStockItem) []types.LowStockItemData {
	var result []types.LowStockItemData
	for _, item := range items {
		result = append(result, types.LowStockItemData{
			ID:           item.ProductID,
			ProductName:  item.ProductName,
			CurrentStock: item.CurrentStock,
			ReorderLevel: item.ReorderLevel,
		})
	}
	return result
}
