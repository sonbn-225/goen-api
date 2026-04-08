package dto

import (
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

func NewDebtResponse(d entity.Debt) DebtResponse {
	return DebtResponse{
		ID:                   d.ID,
		UserID:               d.UserID,
		AccountID:                d.AccountID,
		OriginatingTransactionID: d.OriginatingTransactionID,
		Direction:                d.Direction,
		Name:                 d.Name,
		ContactID:            d.ContactID,
		ContactName:          d.ContactName,
		ContactAvatarURL:     d.ContactAvatarURL,
		Principal:            d.Principal,
		Currency:             d.Currency,
		StartDate:            d.StartDate,
		DueDate:              d.DueDate,
		InterestRate:         d.InterestRate,
		InterestRule:         d.InterestRule,
		OutstandingPrincipal: d.OutstandingPrincipal,
		AccruedInterest:      d.AccruedInterest,
		Status:               d.Status,
		CreatedAt:            d.CreatedAt,
	}
}

func NewDebtResponses(items []entity.Debt) []DebtResponse {
	out := make([]DebtResponse, len(items))
	for i, it := range items {
		out[i] = NewDebtResponse(it)
	}
	return out
}

func NewDebtPaymentLinkResponse(l entity.DebtPaymentLink) DebtPaymentLinkResponse {
	return DebtPaymentLinkResponse{
		ID:            l.ID,
		DebtID:        l.DebtID,
		TransactionID: l.TransactionID,
		PrincipalPaid: l.PrincipalPaid,
		InterestPaid:  l.InterestPaid,
		CreatedAt:     l.CreatedAt,
	}
}

func NewDebtPaymentLinkResponses(items []entity.DebtPaymentLink) []DebtPaymentLinkResponse {
	out := make([]DebtPaymentLinkResponse, len(items))
	for i, it := range items {
		out[i] = NewDebtPaymentLinkResponse(it)
	}
	return out
}
