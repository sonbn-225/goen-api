package entity

import (
	"time"

	"github.com/google/uuid"
)

// Debt represents a loan or debt between the user and a contact.
// Debt represents a loan or debt between the user and a contact.
type Debt struct {
	AuditEntity
	UserID               uuid.UUID      `json:"user_id"`                        // ID of the user who owns the debt record
	AccountID                *uuid.UUID     `json:"account_id,omitempty"`           // Optional ID of the account linked to this debt
	OriginatingTransactionID *uuid.UUID     `json:"originating_transaction_id,omitempty"` // ID of the transaction that created this debt
	Direction                DebtDirection  `json:"direction"`                      // Whether money was lent or borrowed
	Name                 *string        `json:"name,omitempty"`                 // Optional name/label for the debt
	ContactID            *uuid.UUID     `json:"contact_id,omitempty"`           // Optional ID of the contact involved
	ContactName          *string        `json:"contact_name,omitempty"`         // Name of the contact (enriched)
	ContactAvatarURL     *string        `json:"contact_avatar_url,omitempty"`   // Avatar URL of the contact (enriched)
	Principal            string         `json:"principal"`                      // Initial principal amount (decimal string)
	Currency             *string        `json:"currency,omitempty"`             // Currency of the debt (e.g., "VND")
	StartDate            string         `json:"start_date"`                     // Date the debt started (YYYY-MM-DD)
	DueDate              string         `json:"due_date"`                       // Final due date for the debt (YYYY-MM-DD)
	InterestRate         *string        `json:"interest_rate,omitempty"`        // Annual interest rate (decimal string)
	InterestRule         *string        `json:"interest_rule,omitempty"`        // Description of interest calculation rules
	OutstandingPrincipal string         `json:"outstanding_principal"`           // Current remaining principal (decimal string)
	AccruedInterest      string         `json:"accrued_interest"`               // Current accumulated interest (decimal string)
	Status               DebtStatus     `json:"status"`                         // Current debt status (active/paid/cancelled)
	ClosedAt             *time.Time     `json:"closed_at,omitempty"`           // Timestamp when the debt was closed
}

// DebtPaymentLink connects a transaction to a specific debt repayment.
type DebtPaymentLink struct {
	BaseEntity
	DebtID        uuid.UUID `json:"debt_id"`                 // ID of the debt being repaid
	TransactionID uuid.UUID `json:"transaction_id"`          // ID of the repayment transaction
	PrincipalPaid *string   `json:"principal_paid,omitempty"` // Amount of principal repaid (decimal string)
	InterestPaid  *string   `json:"interest_paid,omitempty"`  // Amount of interest paid (decimal string)
	CreatedAt     time.Time `json:"created_at"`              // Creation timestamp
}

// DebtInstallment represents a single scheduled payment for a debt.
type DebtInstallment struct {
	BaseEntity
	DebtID        uuid.UUID `json:"debt_id"`        // ID of the debt this installment belongs to
	InstallmentNo int       `json:"installment_no"` // Sequence number in the payment schedule
	DueDate       string    `json:"due_date"`       // Due date for this installment (YYYY-MM-DD)
	AmountDue     string    `json:"amount_due"`     // Total amount due (decimal string)
	AmountPaid    string    `json:"amount_paid"`    // Amount already paid (decimal string)
	Status        string    `json:"status"`         // Current status (e.g., "pending", "paid")
}
