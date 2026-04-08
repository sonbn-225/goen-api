package entity

import (
	"time"

	"github.com/google/uuid"
)

// TransactionLineItem represents a single categorized part of a transaction.
type TransactionLineItem struct {
	BaseEntity
	CategoryID    *uuid.UUID  `json:"category_id,omitempty"` // ID of the category for this line item
	TagIDs        []uuid.UUID `json:"tag_ids,omitempty"`     // IDs of tags associated with this line item
	Amount        string      `json:"amount"`               // Amount for this specific line item (decimal string)
	Note          *string     `json:"note,omitempty"`        // Optional memo for this line item
	TransactionID uuid.UUID   `json:"-"`                    // ID of the parent transaction
}

// Transaction represents a financial record of income, expense, or transfer.
type Transaction struct {
	AuditEntity
	ExternalRef  *string   `json:"external_ref,omitempty"` // Reference ID from an external system (e.g., bank ref)
	Type         TransactionType    `json:"type"`          // Transaction type (income/expense/transfer)
	OccurredAt   time.Time `json:"occurred_at"`            // Exact timestamp of the transaction
	OccurredDate string    `json:"occurred_date"`          // Date of the transaction (YYYY-MM-DD)
	Amount       string    `json:"amount"`                 // Total transaction amount (decimal string)
	// For FX transfers, amounts can differ per side.
	FromAmount *string `json:"from_amount,omitempty"` // Amount deducted from source account (for transfers)
	ToAmount   *string `json:"to_amount,omitempty"`   // Amount added to destination account (for transfers)
	// Currencies are derived from linked accounts (not stored on transactions).
	AccountCurrency *string               `json:"account_currency,omitempty"` // Currency of the primary account (enriched)
	FromCurrency    *string               `json:"from_currency,omitempty"`    // Currency of the source account (enriched)
	ToCurrency      *string               `json:"to_currency,omitempty"`      // Currency of the destination account (enriched)
	Description     *string               `json:"description,omitempty"`      // Description or memo of the transaction
	AccountID       *uuid.UUID            `json:"account_id,omitempty"`       // ID of the primary account (for income/expense)
	FromAccountID   *uuid.UUID            `json:"from_account_id,omitempty"`   // ID of the source account (for transfers)
	ToAccountID     *uuid.UUID            `json:"to_account_id,omitempty"`     // ID of the destination account (for transfers)
	ExchangeRate    *string               `json:"exchange_rate,omitempty"`    // Exchange rate used for cross-currency transfers
	Status          TransactionStatus     `json:"status"`                    // Current state (pending/posted/cancelled)
	LineItems       []TransactionLineItem `json:"line_items,omitempty"`      // Breakdown of the transaction into categories/tags
	TagIDs          []uuid.UUID           `json:"tag_ids,omitempty"`          // Merged tag IDs from all line items (enriched)
	CategoryIDs     []uuid.UUID           `json:"category_ids,omitempty"`     // Merged category IDs from all line items (enriched)

	// Wrapper fields (enriched)
	AccountName    *string  `json:"account_name,omitempty"`    // Name of the primary account
	CategoryNames  []string `json:"category_names,omitempty"`  // Merged category names for display
	TagNames       []string `json:"tag_names,omitempty"`       // Merged tag names for display
	CategoryColors []string `json:"category_colors,omitempty"` // UI colors for categories
	TagColors      []string `json:"tag_colors,omitempty"`      // UI colors for tags
}

// TransactionListFilter defines the criteria for querying transactions.
type TransactionListFilter struct {
	AccountID         *uuid.UUID // Filter by specific account
	CategoryID        *uuid.UUID // Filter by specific category
	Type              *TransactionType // Filter by transaction type
	Search            *string    // Search text in description/notes
	ExternalRefFamily *string    // Filter by group of external references
	From              *time.Time // Filter by minimum date
	To                *time.Time // Filter by maximum date
	Cursor            *string    // Pagination cursor
	Page              int        // Page number (if using offset pagination)
	Limit             int        // Maximum number of records to return
}

type TransactionPatch struct {
	Description *string            // New main description
	CategoryIDs []uuid.UUID        // New primary categories (atomic replacement)
	TagIDs      []uuid.UUID        // New primary tags (atomic replacement)
	Amount      *string            // New total amount
	Status      *TransactionStatus     // New status (posted/cancelled)
	OccurredAt  *time.Time             // New occurrence timestamp
	LineItems   *[]TransactionLineItem // Complete replacement of line items
}

