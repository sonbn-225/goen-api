package dto
 
import (
	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)
 
// Savings Requests/Responses
// CreateSavingsRequest is the payload for creating a new savings goal or account.
// Used in: SavingsHandler, SavingsService, SavingsInterface
type CreateSavingsRequest struct {
	Name             string    `json:"name"`                 // Name of the savings goal
	SavingsAccountID uuid.UUID `json:"savings_account_id"`   // ID of the dedicated savings account
	ParentAccountID  uuid.UUID `json:"parent_account_id"`    // ID of the source account for funding
	Principal        string    `json:"principal"`           // Initial deposit amount (decimal string)
	InterestRate     *string   `json:"interest_rate,omitempty"` // Annual interest rate (decimal string)
	TermMonths       *int      `json:"term_months,omitempty"`   // Duration in months
	StartDate        *string   `json:"start_date,omitempty"`    // Start date (YYYY-MM-DD)
	MaturityDate     *string   `json:"maturity_date,omitempty"` // Expected maturity date (YYYY-MM-DD)
	AutoRenew        bool      `json:"auto_renew"`              // Whether to auto-renew on maturity
}
 
// PatchSavingsRequest is the payload for updating an existing savings goal.
// Used in: SavingsHandler, SavingsService, SavingsInterface
type PatchSavingsRequest struct {
	Name             *string    `json:"name,omitempty"`             // New goal name
	SavingsAccountID *uuid.UUID `json:"savings_account_id,omitempty"` // New savings account ID
	Principal        *string    `json:"principal,omitempty"`        // New principal amount
	InterestRate     *string    `json:"interest_rate,omitempty"`    // New interest rate
	TermMonths       *int       `json:"term_months,omitempty"`      // New duration
	MaturityDate     *string    `json:"maturity_date,omitempty"`    // New maturity date
	AutoRenew        *bool      `json:"auto_renew,omitempty"`       // New auto-renew setting
	Status           *entity.SavingsStatus `json:"status,omitempty"` // New status (active/matured/closed)
}
 
// SavingsResponse represents a savings goal and its current progress.
// Used in: SavingsHandler, SavingsService, SavingsInterface
type SavingsResponse struct {
	ID               uuid.UUID `json:"id"`                             // Unique savings identifier
	SavingsAccountID uuid.UUID `json:"savings_account_id"`             // ID of the dedicated account
	ParentAccountID  uuid.UUID `json:"parent_account_id"`              // ID of the source account
	Principal        string    `json:"principal"`                      // Initial deposit amount
	InterestRate     *string   `json:"interest_rate,omitempty"`         // Annual interest rate
	TermMonths       *int      `json:"term_months,omitempty"`           // Duration in months
	StartDate        *string   `json:"start_date,omitempty"`            // Start date
	MaturityDate     *string   `json:"maturity_date,omitempty"`         // Maturity date
	AutoRenew        bool      `json:"auto_renew"`                      // Auto-renew setting
	AccruedInterest  string    `json:"accrued_interest"`               // Total interest earned so far
	Status           entity.SavingsStatus `json:"status"`               // Current state of the savings goal
}
