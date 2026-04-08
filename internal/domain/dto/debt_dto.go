package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// CreateDebtRequest is used when creating a new debt or loan.
// Used in: DebtHandler, DebtService, DebtInterface
// CreateDebtRequest is used when creating a new debt or loan.
// Used in: DebtHandler, DebtService, DebtInterface
type CreateDebtRequest struct {
	AccountID                string               `json:"account_id" binding:"required"`    // ID of the account linked to the debt (as string)
	OriginatingTransactionID *string              `json:"originating_transaction_id,omitempty"` // Optional ID of the transaction that created this debt
	Direction                entity.DebtDirection `json:"direction" binding:"required"` // Direction of money (lent/borrowed)
	Name         *string              `json:"name,omitempty"`                 // Optional name or purpose of the debt
	ContactID    *string              `json:"contact_id,omitempty"`           // Optional ID of an existing contact
	ContactName  *string              `json:"contact_name,omitempty"`         // Name of the contact if not an existing one
	Principal    string               `json:"principal" binding:"required"`    // Original principal amount (decimal string)
	StartDate    string               `json:"start_date" binding:"required"`    // Date the debt was initiated (YYYY-MM-DD)
	DueDate      string               `json:"due_date" binding:"required"`      // Target date for full repayment (YYYY-MM-DD)
	InterestRate *string              `json:"interest_rate,omitempty"`        // Annual interest rate (decimal string)
	InterestRule      *string              `json:"interest_rule,omitempty"`        // Description of how interest is calculated
	CreateTransaction bool                 `json:"create_transaction,omitempty"`   // Whether to automatically create a transaction for this debt
}

// UpdateDebtRequest is used when updating an existing debt.
// Used in: DebtHandler, DebtService, DebtInterface
type UpdateDebtRequest struct {
	Name                     *string            `json:"name,omitempty"`         // New name or purpose
	OriginatingTransactionID *string            `json:"originating_transaction_id,omitempty"` // New ID of the transaction that created this debt
	DueDate      *string            `json:"due_date,omitempty"`      // New repayment due date
	Status       *entity.DebtStatus `json:"status,omitempty"`        // New debt status (active/paid/etc.)
	InterestRate *string            `json:"interest_rate,omitempty"` // New annual interest rate
}

// DebtPaymentRequest is used when recording a payment for a debt.
// Used in: DebtHandler, DebtService, DebtInterface
type DebtPaymentRequest struct {
	TransactionID string  `json:"transaction_id" binding:"required"` // ID of the transaction representing the payment
	PrincipalPaid *string `json:"principal_paid,omitempty"`        // Amount of principal covered by this payment
	InterestPaid  *string `json:"interest_paid,omitempty"`         // Amount of interest covered by this payment
	AmountPaid    *string `json:"amount_paid,omitempty"`        // Total paid amount (decimal string)
}

// DebtRepayRequest is used for 1-step repayment (automatically creating a transaction)
type DebtRepayRequest struct {
	AccountID string  `json:"account_id" binding:"required"` // Account to pay from
	Amount    string  `json:"amount" binding:"required"`     // Amount to pay
	Note      *string `json:"note,omitempty"`              // Optional note
}

// DebtResponse represents the debt information sent back to the client.
// Used in: DebtHandler, DebtService, DebtInterface
type DebtResponse struct {
	ID                   uuid.UUID            `json:"id"`                             // Unique debt identifier
	UserID               uuid.UUID            `json:"user_id"`                        // ID of the user who owns the record
	AccountID                *uuid.UUID           `json:"account_id,omitempty"`           // ID of the linked account
	OriginatingTransactionID *uuid.UUID           `json:"originating_transaction_id,omitempty"` // ID of the transaction that created this debt
	Direction                entity.DebtDirection `json:"direction"`                      // Lent or Borrowed
	Name                     *string              `json:"name,omitempty"`                 // Name or purpose of the debt
	ContactID            *uuid.UUID           `json:"contact_id,omitempty"`           // ID of the associated contact
	ContactName          *string              `json:"contact_name,omitempty"`         // Name of the contact (enriched)
	ContactAvatarURL     *string              `json:"contact_avatar_url,omitempty"`   // Avatar URL of the contact (enriched)
	Principal            string               `json:"principal"`                      // Original principal amount
	Currency             *string              `json:"currency,omitempty"`             // Currency of the debt (e.g., "VND")
	StartDate            string               `json:"start_date"`                     // Initiation date
	DueDate              string               `json:"due_date"`                       // Repayment due date
	InterestRate         *string              `json:"interest_rate,omitempty"`        // Annual interest rate
	InterestRule         *string              `json:"interest_rule,omitempty"`        // Interest calculation details
	OutstandingPrincipal string               `json:"outstanding_principal"`           // Remaining principal to be paid
	AccruedInterest      string               `json:"accrued_interest"`               // Total interest accumulated so far
	Status               entity.DebtStatus    `json:"status"`                         // Current status of the debt
	CreatedAt            time.Time            `json:"created_at"`                     // Timestamp when the record was created
}

// DebtPaymentLinkResponse represents the link between a debt and a payment transaction.
// Used in: DebtHandler, DebtService, DebtInterface
type DebtPaymentLinkResponse struct {
	ID            uuid.UUID `json:"id"`                             // Unique link identifier
	DebtID        uuid.UUID `json:"debt_id"`                        // ID of the repaid debt
	TransactionID uuid.UUID `json:"transaction_id"`                 // ID of the payment transaction
	PrincipalPaid *string   `json:"principal_paid,omitempty"`        // Principal amount paid in this transaction
	InterestPaid  *string   `json:"interest_paid,omitempty"`         // Interest amount paid in this transaction
	CreatedAt     time.Time `json:"created_at"`                     // Timestamp when the link was created
}
