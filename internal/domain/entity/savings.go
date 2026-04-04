package entity

import (
	"time"
)

// Savings represents a simple savings product like a Term Deposit or Goal.
type Savings struct {
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
