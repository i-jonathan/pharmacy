package service

import (
	"context"
	"pharmacy/internal/constant"
	"pharmacy/internal/types"
)

func HasPermission(ctx context.Context, permissionKey string) bool {
	perms, ok := ctx.Value(constant.PermissionsSessionKey).(map[string]bool)
	if !ok {
		return false
	}

	return perms[permissionKey]
}

func calculateStockDifference(item types.StockTakingItemData) (int, string, bool) {
	total := 0

	if item.DispensaryCount != nil {
		total += *item.DispensaryCount
	}
	if item.StoreCount != nil {
		total += *item.StoreCount
	}

	snapshot := 0
	if item.SnapshotQuantity != nil {
		snapshot = *item.SnapshotQuantity
	}

	diff := total - snapshot
	if diff == 0 {
		return 0, "", true
	}

	movementType := constant.StockTakingIncrease
	if diff < 0 {
		movementType = constant.StockTakingDecrease
		diff = -diff
	}

	return diff, movementType, false
}
