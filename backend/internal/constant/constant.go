package constant

import (
	"strings"
)

type contextKey string

const UserIDKey contextKey = "userID"
const UserSessionKey string = "user_id"
const UserNameSessionKey string = "user_name"
const RoleSessionKey string = "role_id"
const RoleNameSessionKey string = "role_name"
const PermissionsSessionKey string = "permissions"

const DefaultPriceName string = "Base"
 
const CompleteSaleStatus = "COMPLETED"

const CashPaymentMethod = "Cash"
const CardPaymentMethod = "Card"
const TransferPaymentMethod = "Transfer"

type HoldTransactionType string

const HoldSaleType HoldTransactionType = "SALE"
const HoldReceivingItemType HoldTransactionType = "RECEIVING_ITEMS"

func NormalizePaymentMethod(input string) string {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "cash":
		return CashPaymentMethod
	case "card":
		return CardPaymentMethod
	case "transfer":
		return TransferPaymentMethod
	default:
		return CashPaymentMethod
	}
}

type RequirePermissionMode string

const RequireAllPermissions RequirePermissionMode = "ALL"
const RequireAnyPermissions RequirePermissionMode = "ANY"
