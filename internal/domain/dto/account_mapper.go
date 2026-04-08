package dto

import (
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

func NewAccountResponse(it entity.Account) AccountResponse {
	return AccountResponse{
		ID:              it.ID,
		Name:            it.Name,
		AccountNumber:   it.AccountNumber,
		AccountType:     it.AccountType,
		Currency:        it.Currency,
		ParentAccountID: it.ParentAccountID,
		Status:          it.Status,
		Balance:         it.Balance,
		Settings:        NewAccountSettingsResponse(it.Settings),
	}
}



func NewAccountResponses(items []entity.Account) []AccountResponse {
	out := make([]AccountResponse, len(items))
	for i, it := range items {
		out[i] = NewAccountResponse(it)
	}
	return out
}

func NewAccountBalanceResponse(it entity.AccountBalance) AccountBalanceResponse {
	return AccountBalanceResponse{
		AccountID: it.AccountID,
		Currency:  it.Currency,
		Balance:   it.Balance,
	}
}

func NewAccountBalanceResponses(items []entity.AccountBalance) []AccountBalanceResponse {
	out := make([]AccountBalanceResponse, len(items))
	for i, it := range items {
		out[i] = NewAccountBalanceResponse(it)
	}
	return out
}
