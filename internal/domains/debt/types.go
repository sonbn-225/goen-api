package debt

import (
	"context"
	"time"

	"github.com/sonbn-225/goen-api-v2/internal/domains/contact"
	"github.com/sonbn-225/goen-api-v2/internal/domains/transaction"
)

type Debt struct {
	ID                   string     `json:"id"`
	ClientID             *string    `json:"client_id,omitempty"`
	UserID               string     `json:"user_id"`
	AccountID            *string    `json:"account_id,omitempty"`
	Direction            string     `json:"direction"`
	Name                 *string    `json:"name,omitempty"`
	ContactID            *string    `json:"contact_id,omitempty"`
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
	Status               string     `json:"status"`
	ClosedAt             *time.Time `json:"closed_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type DebtPaymentLink struct {
	ID            string    `json:"id"`
	DebtID        string    `json:"debt_id"`
	TransactionID string    `json:"transaction_id"`
	PrincipalPaid *string   `json:"principal_paid,omitempty"`
	InterestPaid  *string   `json:"interest_paid,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type DebtInstallment struct {
	ID            string `json:"id"`
	DebtID        string `json:"debt_id"`
	InstallmentNo int    `json:"installment_no"`
	DueDate       string `json:"due_date"`
	AmountDue     string `json:"amount_due"`
	AmountPaid    string `json:"amount_paid"`
	Status        string `json:"status"`
}

type CreateInput struct {
	ClientID     *string `json:"client_id,omitempty"`
	AccountID    string  `json:"account_id"`
	Direction    string  `json:"direction"`
	Name         *string `json:"name,omitempty"`
	ContactID    *string `json:"contact_id,omitempty"`
	Principal    string  `json:"principal"`
	StartDate    string  `json:"start_date"`
	DueDate      string  `json:"due_date"`
	InterestRate *string `json:"interest_rate,omitempty"`
	InterestRule *string `json:"interest_rule,omitempty"`
	Status       *string `json:"status,omitempty"`
}

type CreatePaymentInput struct {
	TransactionID string  `json:"transaction_id"`
	PrincipalPaid *string `json:"principal_paid,omitempty"`
	InterestPaid  *string `json:"interest_paid,omitempty"`
}

type CreateInstallmentInput struct {
	InstallmentNo int     `json:"installment_no"`
	DueDate       string  `json:"due_date"`
	AmountDue     string  `json:"amount_due"`
	AmountPaid    *string `json:"amount_paid,omitempty"`
	Status        *string `json:"status,omitempty"`
}

type DebtUpdate struct {
	Principal            string
	OutstandingPrincipal string
	AccruedInterest      string
	Status               string
	ClosedAt             *time.Time
	UpdatedAt            time.Time
}

type Repository interface {
	Create(ctx context.Context, userID string, input Debt) error
	GetByID(ctx context.Context, userID, debtID string) (*Debt, error)
	ListByUser(ctx context.Context, userID string) ([]Debt, error)
	CreatePaymentLink(ctx context.Context, userID string, input DebtPaymentLink, update DebtUpdate) error
	ListPaymentLinks(ctx context.Context, userID, debtID string) ([]DebtPaymentLink, error)
	ListPaymentLinksByTransaction(ctx context.Context, userID, transactionID string) ([]DebtPaymentLink, error)
	CreateInstallment(ctx context.Context, userID string, input DebtInstallment) error
	ListInstallments(ctx context.Context, userID, debtID string) ([]DebtInstallment, error)
}

type TransactionService interface {
	Get(ctx context.Context, userID, transactionID string) (*transaction.Transaction, error)
}

type ContactService interface {
	Create(ctx context.Context, userID string, input contact.CreateInput) (*contact.Contact, error)
	List(ctx context.Context, userID string) ([]contact.Contact, error)
}

type Service interface {
	Create(ctx context.Context, userID string, input CreateInput) (*Debt, error)
	Get(ctx context.Context, userID, debtID string) (*Debt, error)
	List(ctx context.Context, userID string) ([]Debt, error)
	CreatePayment(ctx context.Context, userID, debtID string, input CreatePaymentInput) (*DebtPaymentLink, error)
	ListPayments(ctx context.Context, userID, debtID string) ([]DebtPaymentLink, error)
	ListPaymentsByTransaction(ctx context.Context, userID, transactionID string) ([]DebtPaymentLink, error)
	CreateInstallment(ctx context.Context, userID, debtID string, input CreateInstallmentInput) (*DebtInstallment, error)
	ListInstallments(ctx context.Context, userID, debtID string) ([]DebtInstallment, error)
}

type ModuleDeps struct {
	Repo           Repository
	TxService      TransactionService
	ContactService ContactService
	Service        Service
}

type Module struct {
	Service Service
	Handler *Handler
}
