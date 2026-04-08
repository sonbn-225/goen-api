package entity

import (
	"time"

	"github.com/google/uuid"
)

// Savings represents a specialized financial goal managed as an account.
// All metadata is stored in the settings JSONB column of the accounts table.
type Savings struct {
	ID               uuid.UUID     `json:"id"`
	Name             string        `json:"name"`
	SavingsAccountID uuid.UUID     `json:"savings_account_id"`
	ParentAccountID  *uuid.UUID     `json:"parent_account_id,omitempty"`
	Principal        string        `json:"principal"`          // Initial deposit amount (decimal string)
	InterestRate     *string       `json:"interest_rate,omitempty"` // Annual interest rate (decimal string)
	TermMonths       *int          `json:"term_months,omitempty"`   // Duration in months
	StartDate        *string       `json:"start_date,omitempty"`    // Start date (YYYY-MM-DD)
	MaturityDate     *string       `json:"maturity_date,omitempty"` // Expected maturity date (YYYY-MM-DD)
	AutoRenew        bool          `json:"auto_renew"`              // Whether to auto-renew on maturity
	AccruedInterest  string        `json:"accrued_interest"`       // Calculated dynamically from transactions
	Status           AccountStatus `json:"status"`                 // active/matured/closed
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
}

