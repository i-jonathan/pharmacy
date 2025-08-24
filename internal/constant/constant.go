package constant

import (
	"strings"
)

type contextKey string

const UserIDKey contextKey = "userID"
const UserSessionKey string = "user_id"

const DefaultPriceName string = "Base"

const ReceivingSupplyMovementName string = "IN_PURCHASE"

const CompleteSaleStatus = "COMPLETED"

const CashPaymentMethod = "Cash"
const CardPaymentMethod = "Card"
const TransferPaymentMethod = "Transfer"

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
