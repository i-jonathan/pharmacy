package service

import (
	"context"
	"pharmacy/internal/constant"
)

func HasPermission(ctx context.Context, permissionKey string) bool {
	perms, ok := ctx.Value(constant.PermissionsSessionKey).(map[string]bool)
	if !ok {
		return false
	}

	return perms[permissionKey]
}
