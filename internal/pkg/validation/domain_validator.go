package validation

import (
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// IsValidAccountType checks if the provided string is a valid account type.
func IsValidAccountType(t string) bool {
	switch entity.AccountType(t) {
	case entity.AccountTypeBank, entity.AccountTypeWallet, entity.AccountTypeCash,
		entity.AccountTypeBroker, entity.AccountTypeCard, entity.AccountTypeSavings:
		return true
	default:
		return false
	}
}

// IsValidTransactionType checks if the provided string is a valid transaction type.
func IsValidTransactionType(t string) bool {
	switch t {
	case "income", "expense", "transfer":
		return true
	default:
		return false
	}
}
