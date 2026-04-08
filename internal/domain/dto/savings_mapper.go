package dto

import (
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

func NewSavingsResponse(s entity.Savings) SavingsResponse {
	return SavingsResponse{
		ID:               s.ID,
		SavingsAccountID: s.SavingsAccountID,
		ParentAccountID:  s.ParentAccountID,
		Principal:        s.Principal,
		InterestRate:     s.InterestRate,
		TermMonths:       s.TermMonths,
		StartDate:        s.StartDate,
		MaturityDate:     s.MaturityDate,
		AutoRenew:        s.AutoRenew,
		AccruedInterest:  s.AccruedInterest,
		Status:           s.Status,
	}
}

func NewSavingsResponses(items []entity.Savings) []SavingsResponse {
	out := make([]SavingsResponse, len(items))
	for i, it := range items {
		out[i] = NewSavingsResponse(it)
	}
	return out
}
