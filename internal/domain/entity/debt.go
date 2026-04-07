package entity

import (
	"time"

	"github.com/google/uuid"
)

type DebtDirection string

const (
	DebtDirectionLent     DebtDirection = "lent"
	DebtDirectionBorrowed DebtDirection = "borrowed"
)

type DebtStatus string

const (
	DebtStatusActive    DebtStatus = "active"
	DebtStatusPaid      DebtStatus = "paid"
	DebtStatusCancelled DebtStatus = "cancelled"
)

type Debt struct {
	AuditEntity
	UserID               uuid.UUID      `json:"user_id"`
	AccountID            *uuid.UUID     `json:"account_id,omitempty"`
	Direction            DebtDirection  `json:"direction"`
	Name                 *string        `json:"name,omitempty"`
	ContactID            *uuid.UUID     `json:"contact_id,omitempty"`
	ContactName          *string        `json:"contact_name,omitempty"`
	ContactAvatarURL     *string        `json:"contact_avatar_url,omitempty"`
	Principal            string         `json:"principal"`
	Currency             *string        `json:"currency,omitempty"`
	StartDate            string         `json:"start_date"`
	DueDate              string         `json:"due_date"`
	InterestRate         *string        `json:"interest_rate,omitempty"`
	InterestRule         *string        `json:"interest_rule,omitempty"`
	OutstandingPrincipal string         `json:"outstanding_principal"`
	AccruedInterest      string         `json:"accrued_interest"`
	Status               DebtStatus     `json:"status"`
	ClosedAt             *time.Time     `json:"closed_at,omitempty"`
}

type DebtPaymentLink struct {
	BaseEntity
	DebtID        uuid.UUID `json:"debt_id"`
	TransactionID uuid.UUID `json:"transaction_id"`
	PrincipalPaid *string   `json:"principal_paid,omitempty"`
	InterestPaid  *string   `json:"interest_paid,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type DebtInstallment struct {
	BaseEntity
	DebtID        uuid.UUID `json:"debt_id"`
	InstallmentNo int       `json:"installment_no"`
	DueDate       string    `json:"due_date"`
	AmountDue     string    `json:"amount_due"`
	AmountPaid    string    `json:"amount_paid"`
	Status        string    `json:"status"`
}
