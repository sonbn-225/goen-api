package dto

import (
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

func NewTransactionResponse(t entity.Transaction, debtLinks []DebtPaymentLinkResponse, groupParticipants []DebtResponse) TransactionResponse {
	lineItems := make([]TransactionLineItemResponse, len(t.LineItems))
	for i, li := range t.LineItems {
		lineItems[i] = TransactionLineItemResponse{
			ID:         li.ID,
			CategoryID: li.CategoryID,
			TagIDs:     li.TagIDs,
			Amount:     li.Amount,
			Note:       li.Note,
		}
	}

	return TransactionResponse{
		ID:              t.ID,
		ExternalRef:     t.ExternalRef,
		Type:            t.Type,
		OccurredAt:      t.OccurredAt,
		OccurredDate:    t.OccurredDate,
		Amount:          t.Amount,
		FromAmount:      t.FromAmount,
		ToAmount:        t.ToAmount,
		AccountCurrency: t.AccountCurrency,
		FromCurrency:    t.FromCurrency,
		ToCurrency:      t.ToCurrency,
		Description:     t.Description,
		AccountID:       t.AccountID,
		FromAccountID:   t.FromAccountID,
		ToAccountID:     t.ToAccountID,
		ExchangeRate:    t.ExchangeRate,
		Status:          t.Status,
		LineItems:       lineItems,
		TagIDs:          t.TagIDs,
		CategoryIDs:     t.CategoryIDs,
		CategoryNames:   t.CategoryNames,
		TagNames:        t.TagNames,
		CategoryColors:  t.CategoryColors,
		TagColors:       t.TagColors,
		DebtLinks:       debtLinks,
		GroupParticipants: groupParticipants,
	}
}

func NewTransactionResponses(items []entity.Transaction) []TransactionResponse {
	out := make([]TransactionResponse, len(items))
	for i, it := range items {
		out[i] = NewTransactionResponse(it, nil, nil)
	}
	return out
}


