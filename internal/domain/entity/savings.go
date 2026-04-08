package entity

import (
	"time"

	"github.com/google/uuid"
)

// Savings represents a fixed-term deposit or a specific savings goal.
type Savings struct {
	AuditEntity
	SavingsAccountID uuid.UUID     `json:"savings_account_id"` // ID of the dedicated savings account
	ParentAccountID  uuid.UUID     `json:"parent_account_id"`  // ID of the source account for funding
	Principal        string        `json:"principal"`          // Initial deposit amount (decimal string)
	InterestRate     *string       `json:"interest_rate,omitempty"` // Annual interest rate (decimal string)
	TermMonths       *int          `json:"term_months,omitempty"`   // Duration of the savings term in months
	StartDate        *string       `json:"start_date,omitempty"`    // Start date of the savings term (YYYY-MM-DD)
	MaturityDate     *string       `json:"maturity_date,omitempty"` // Expected maturity date (YYYY-MM-DD)
	AutoRenew        bool          `json:"auto_renew"`              // Whether to automatically renew the term on maturity
	AccruedInterest  string        `json:"accrued_interest"`       // Current accumulated interest (decimal string)
	Status           SavingsStatus `json:"status"`                 // Current state (active/matured/closed)
	ClosedAt         *time.Time    `json:"closed_at,omitempty"`     // Timestamp when the savings goal was reached or closed
}

