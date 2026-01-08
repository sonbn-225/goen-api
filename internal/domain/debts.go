package domain

import (
	"context"
	"errors"
	"time"
)

var ErrDebtNotFound = errors.New("debt not found")

type Debt struct {
	ID                 string     `json:"id"`
	ClientID           *string    `json:"client_id,omitempty"`
	UserID             string     `json:"user_id"`
	Direction          string     `json:"direction"`
	Name               *string    `json:"name,omitempty"`
	Principal          string     `json:"principal"`
	Currency           string     `json:"currency"`
	StartDate          string     `json:"start_date"`
	DueDate            string     `json:"due_date"`
	InterestRate       *string    `json:"interest_rate,omitempty"`
	InterestRule       *string    `json:"interest_rule,omitempty"`
	OutstandingPrincipal string   `json:"outstanding_principal"`
	AccruedInterest    string     `json:"accrued_interest"`
	Status             string     `json:"status"`
	ClosedAt           *time.Time `json:"closed_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type DebtPaymentLink struct {
	ID           string    `json:"id"`
	DebtID       string    `json:"debt_id"`
	TransactionID string   `json:"transaction_id"`
	PrincipalPaid *string  `json:"principal_paid,omitempty"`
	InterestPaid  *string  `json:"interest_paid,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type DebtInstallment struct {
	ID            string  `json:"id"`
	DebtID        string  `json:"debt_id"`
	InstallmentNo int     `json:"installment_no"`
	DueDate       string  `json:"due_date"`
	AmountDue     string  `json:"amount_due"`
	AmountPaid    string  `json:"amount_paid"`
	Status        string  `json:"status"`
}

type DebtRepository interface {
	CreateDebt(ctx context.Context, debt Debt) error
	GetDebt(ctx context.Context, userID string, debtID string) (*Debt, error)
	ListDebts(ctx context.Context, userID string) ([]Debt, error)

	CreatePaymentLink(ctx context.Context, userID string, link DebtPaymentLink, newOutstandingPrincipal string, newAccruedInterest string, newStatus string, closedAt *time.Time) error
	ListPaymentLinks(ctx context.Context, userID string, debtID string) ([]DebtPaymentLink, error)
	ListPaymentLinksByTransaction(ctx context.Context, userID string, transactionID string) ([]DebtPaymentLink, error)

	CreateInstallment(ctx context.Context, userID string, inst DebtInstallment) error
	ListInstallments(ctx context.Context, userID string, debtID string) ([]DebtInstallment, error)
}
