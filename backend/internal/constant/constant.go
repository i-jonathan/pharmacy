package constant

import (
	"strings"
)

type contextKey string

const UserIDKey contextKey = "userID"
const UserSessionKey string = "user_id"
const RoleSessionKey string = "role_id"
const PermissionsSessionKey string = "permissions"

const DefaultPriceName string = "Base"

const ReceivingSupplyMovementName string = "IN_PURCHASE"
const ReturnSaleMovementName string = "IN_SALE_RETURN"
const SaleMovementName string = "OUT_SALE"

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
