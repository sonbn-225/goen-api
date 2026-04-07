package entity

import (
	"time"

	"github.com/google/uuid"
)

type SavingsStatus string

const (
	SavingsStatusActive  SavingsStatus = "active"
	SavingsStatusMatured SavingsStatus = "matured"
	SavingsStatusClosed  SavingsStatus = "closed"
)

// Savings represents a simple savings product like a Term Deposit or Goal.
type Savings struct {
	AuditEntity
	SavingsAccountID uuid.UUID     `json:"savings_account_id"`
	ParentAccountID  uuid.UUID     `json:"parent_account_id"`
	Principal        string        `json:"principal"`
	InterestRate     *string       `json:"interest_rate,omitempty"`
	TermMonths       *int          `json:"term_months,omitempty"`
	StartDate        *string       `json:"start_date,omitempty"`
	MaturityDate     *string       `json:"maturity_date,omitempty"`
	AutoRenew        bool          `json:"auto_renew"`
	AccruedInterest  string        `json:"accrued_interest"`
	Status           SavingsStatus `json:"status"`
	ClosedAt         *time.Time    `json:"closed_at,omitempty"`
}

