package dto

import (
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// Savings Requests/Responses
type CreateSavingsRequest struct {
	Name             string  `json:"name"`
	SavingsAccountID string  `json:"savings_account_id"`
	ParentAccountID  string  `json:"parent_account_id"`
	Principal        string  `json:"principal"`
	InterestRate     *string `json:"interest_rate,omitempty"`
	TermMonths       *int    `json:"term_months,omitempty"`
	StartDate        *string `json:"start_date,omitempty"`
	MaturityDate     *string `json:"maturity_date,omitempty"`
	AutoRenew        bool    `json:"auto_renew"`
}

type PatchSavingsRequest struct {
	Name             *string `json:"name,omitempty"`
	SavingsAccountID *string `json:"savings_account_id,omitempty"`
	Principal        *string `json:"principal,omitempty"`
	InterestRate     *string `json:"interest_rate,omitempty"`
	TermMonths       *int    `json:"term_months,omitempty"`
	MaturityDate     *string `json:"maturity_date,omitempty"`
	AutoRenew        *bool   `json:"auto_renew,omitempty"`
	Status           *string `json:"status,omitempty"`
}

type SavingsResponse struct {
	ID               string  `json:"id"`
	SavingsAccountID string  `json:"savings_account_id"`
	ParentAccountID  string  `json:"parent_account_id"`
	Principal        string  `json:"principal"`
	InterestRate     *string `json:"interest_rate,omitempty"`
	TermMonths       *int    `json:"term_months,omitempty"`
	StartDate        *string `json:"start_date,omitempty"`
	MaturityDate     *string `json:"maturity_date,omitempty"`
	AutoRenew        bool    `json:"auto_renew"`
	AccruedInterest  string  `json:"accrued_interest"`
	Status           string  `json:"status"`
}

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
