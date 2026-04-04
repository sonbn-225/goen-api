package savings

import (
	"context"
	"time"

	"github.com/sonbn-225/goen-api-v2/internal/domains/transaction"
)

type SavingsInstrument struct {
	ID               string     `json:"id"`
	SavingsAccountID string     `json:"savings_account_id"`
	ParentAccountID  string     `json:"parent_account_id"`
	Principal        string     `json:"principal"`
	InterestRate     *string    `json:"interest_rate,omitempty"`
	TermMonths       *int       `json:"term_months,omitempty"`
	StartDate        *string    `json:"start_date,omitempty"`
	MaturityDate     *string    `json:"maturity_date,omitempty"`
	AutoRenew        bool       `json:"auto_renew"`
	AccruedInterest  string     `json:"accrued_interest"`
	Status           string     `json:"status"`
	ClosedAt         *time.Time `json:"closed_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type AccountRef struct {
	ID              string
	Name            string
	Type            string
	Currency        string
	ParentAccountID *string
}

type CreateInput struct {
	Name             *string `json:"name,omitempty"`
	SavingsAccountID *string `json:"savings_account_id,omitempty"`
	ParentAccountID  *string `json:"parent_account_id,omitempty"`
	Principal        string  `json:"principal"`
	InterestRate     *string `json:"interest_rate,omitempty"`
	TermMonths       *int    `json:"term_months,omitempty"`
	StartDate        *string `json:"start_date,omitempty"`
	MaturityDate     *string `json:"maturity_date,omitempty"`
	AutoRenew        *bool   `json:"auto_renew,omitempty"`
	AccruedInterest  *string `json:"accrued_interest,omitempty"`
	Status           *string `json:"status,omitempty"`
}

type PatchInput struct {
	Principal       *string `json:"principal,omitempty"`
	InterestRate    *string `json:"interest_rate,omitempty"`
	TermMonths      *int    `json:"term_months,omitempty"`
	StartDate       *string `json:"start_date,omitempty"`
	MaturityDate    *string `json:"maturity_date,omitempty"`
	AutoRenew       *bool   `json:"auto_renew,omitempty"`
	AccruedInterest *string `json:"accrued_interest,omitempty"`
	Status          *string `json:"status,omitempty"`
}

type Repository interface {
	GetAccountForUser(ctx context.Context, userID, accountID string) (*AccountRef, error)
	CreateLinkedSavingsAccount(ctx context.Context, userID, parentAccountID, accountName, currency string) (*AccountRef, error)
	DeleteAccountForUser(ctx context.Context, userID, accountID string) error

	CreateSavingsInstrument(ctx context.Context, userID string, item SavingsInstrument) error
	GetSavingsInstrument(ctx context.Context, userID, instrumentID string) (*SavingsInstrument, error)
	ListSavingsInstruments(ctx context.Context, userID string) ([]SavingsInstrument, error)
	UpdateSavingsInstrument(ctx context.Context, userID string, item SavingsInstrument) error
	DeleteSavingsInstrument(ctx context.Context, userID, instrumentID string) error
}

type TransactionService interface {
	Create(ctx context.Context, userID string, input transaction.CreateInput) (*transaction.Transaction, error)
}

type Service interface {
	Create(ctx context.Context, userID string, input CreateInput) (*SavingsInstrument, error)
	Get(ctx context.Context, userID, instrumentID string) (*SavingsInstrument, error)
	List(ctx context.Context, userID string) ([]SavingsInstrument, error)
	Patch(ctx context.Context, userID, instrumentID string, input PatchInput) (*SavingsInstrument, error)
	Delete(ctx context.Context, userID, instrumentID string) error
}

type ModuleDeps struct {
	Repo      Repository
	TxService TransactionService
	Service   Service
}

type Module struct {
	Service Service
	Handler *Handler
}
