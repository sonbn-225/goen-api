package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrSavingsInstrumentNotFound = errors.New("savings instrument not found")
)

type SavingsInstrument struct {
	ID              string     `json:"id"`
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

type SavingsRepository interface {
	CreateSavingsInstrument(ctx context.Context, userID string, s SavingsInstrument) error
	GetSavingsInstrument(ctx context.Context, userID string, savingsInstrumentID string) (*SavingsInstrument, error)
	ListSavingsInstruments(ctx context.Context, userID string) ([]SavingsInstrument, error)
	UpdateSavingsInstrument(ctx context.Context, userID string, s SavingsInstrument) error
	DeleteSavingsInstrument(ctx context.Context, userID string, savingsInstrumentID string) error
}
