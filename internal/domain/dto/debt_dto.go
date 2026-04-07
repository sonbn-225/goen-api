package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type CreateDebtRequest struct {
	AccountID    string  `json:"account_id" binding:"required"`
	Direction    entity.DebtDirection  `json:"direction" binding:"required"` // lent, borrowed
	Name         *string `json:"name,omitempty"`
	ContactID    *string `json:"contact_id,omitempty"`
	ContactName  *string `json:"contact_name,omitempty"`
	Principal    string  `json:"principal" binding:"required"`
	StartDate    string  `json:"start_date" binding:"required"`
	DueDate      string  `json:"due_date" binding:"required"`
	InterestRate *string `json:"interest_rate,omitempty"`
	InterestRule *string `json:"interest_rule,omitempty"`
}

type UpdateDebtRequest struct {
	Name         *string `json:"name,omitempty"`
	DueDate      *string `json:"due_date,omitempty"`
	Status       *entity.DebtStatus    `json:"status,omitempty"`
	InterestRate *string `json:"interest_rate,omitempty"`
}

type DebtPaymentRequest struct {
	TransactionID string  `json:"transaction_id" binding:"required"`
	PrincipalPaid *string `json:"principal_paid,omitempty"`
	InterestPaid  *string `json:"interest_paid,omitempty"`
	AmountPaid    *string `json:"amount_paid,omitempty"` // Total paid, can be split by service
}

type DebtResponse struct {
	ID                   uuid.UUID  `json:"id"`
	UserID               uuid.UUID  `json:"user_id"`
	AccountID            *uuid.UUID `json:"account_id,omitempty"`
	Direction            entity.DebtDirection  `json:"direction"`
	Name                 *string    `json:"name,omitempty"`
	ContactID            *uuid.UUID `json:"contact_id,omitempty"`
	ContactName          *string    `json:"contact_name,omitempty"`
	ContactAvatarURL     *string    `json:"contact_avatar_url,omitempty"`
	Principal            string     `json:"principal"`
	Currency             *string    `json:"currency,omitempty"`
	StartDate            string     `json:"start_date"`
	DueDate              string     `json:"due_date"`
	InterestRate         *string    `json:"interest_rate,omitempty"`
	InterestRule         *string    `json:"interest_rule,omitempty"`
	OutstandingPrincipal string     `json:"outstanding_principal"`
	AccruedInterest      string     `json:"accrued_interest"`
	Status               entity.DebtStatus     `json:"status"`
	CreatedAt            time.Time  `json:"created_at"`
}

func NewDebtResponse(d entity.Debt) DebtResponse {
	return DebtResponse{
		ID:                   d.ID,
		UserID:               d.UserID,
		AccountID:            d.AccountID,
		Direction:            d.Direction,
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

type DebtPaymentLinkResponse struct {
	ID            uuid.UUID `json:"id"`
	DebtID        uuid.UUID `json:"debt_id"`
	TransactionID uuid.UUID `json:"transaction_id"`
	PrincipalPaid *string   `json:"principal_paid,omitempty"`
	InterestPaid  *string   `json:"interest_paid,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
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
