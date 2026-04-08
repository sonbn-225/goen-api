package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// CreateLineItemRequest is the payload for a single line item within a transaction.
// Used in: TransactionHandler, TransactionService, TransactionInterface
type CreateLineItemRequest struct {
	CategoryID *uuid.UUID  `json:"category_id,omitempty"` // ID of the category for this line item
	TagIDs     []uuid.UUID `json:"tag_ids,omitempty"`     // IDs of tags for this line item
	Amount     string      `json:"amount"`               // Amount allocated to this category (decimal string)
	Note       *string     `json:"note,omitempty"`        // Optional memo for this line item
}

// GroupParticipantInput represents a participant in a group expense transaction.
// Used in: TransactionHandler, TransactionService, TransactionInterface
type GroupParticipantInput struct {
	ParticipantName string `json:"participant_name"` // Name of the person sharing the expense
	OriginalAmount  string `json:"original_amount"`  // The amount they were originally responsible for
	ShareAmount     string `json:"share_amount"`     // Their final calculated share after adjustments
}

// CreateTransactionRequest is the main payload for creating any type of transaction.
// Used in: TransactionHandler, TransactionService, TransactionInterface
type CreateTransactionRequest struct {
	ExternalRef         *string                 `json:"external_ref,omitempty"`         // Optional reference from external system
	Type                entity.TransactionType  `json:"type"`                          // Income, Expense, or Transfer
	OccurredAt          *string                 `json:"occurred_at,omitempty"`          // Full timestamp string
	OccurredDate        *string                 `json:"occurred_date,omitempty"`        // Date (YYYY-MM-DD)
	OccurredTime        *string                 `json:"occurred_time,omitempty"`        // Time (HH:MM:SS)
	Amount              string                  `json:"amount"`                         // Total transaction amount
	FromAmount          *string                 `json:"from_amount,omitempty"`          // Amount deducted (for transfers)
	ToAmount            *string                 `json:"to_amount,omitempty"`            // Amount added (for transfers)
	Description         *string                 `json:"description,omitempty"`          // Main transaction description
	AccountID           *uuid.UUID              `json:"account_id,omitempty"`           // Primary account (for income/expense)
	FromAccountID       *uuid.UUID              `json:"from_account_id,omitempty"`       // Source account (for transfers)
	ToAccountID         *uuid.UUID              `json:"to_account_id,omitempty"`         // Destination account (for transfers)
	ExchangeRate        *string                 `json:"exchange_rate,omitempty"`        // Manual exchange rate for transfers
	CategoryID          *uuid.UUID              `json:"category_id,omitempty"`          // Primary category (if no line items)
	TagIDs              []uuid.UUID             `json:"tag_ids,omitempty"`              // Primary tags (if no line items)
	LineItems           []CreateLineItemRequest `json:"line_items,omitempty"`          // Detailed breakdown into categories
	GroupParticipants   []GroupParticipantInput `json:"group_participants,omitempty"`   // Shared expense participants
	OwnerOriginalAmount *string                 `json:"owner_original_amount,omitempty"` // The creator's share of a group expense
	Lang                string                  `json:"lang,omitempty"`                 // Preferred language for category names (enriched)
	Source              *string                 `json:"source,omitempty"`              // Source of the transaction (e.g., 'manual', 'import')
}

// TransactionPatchRequest is the payload for updating an existing transaction.
// Used in: TransactionHandler, TransactionService, TransactionInterface
type TransactionPatchRequest struct {
	Description       *string                  `json:"description,omitempty"`       // Updated description
	CategoryIDs       []uuid.UUID              `json:"category_ids,omitempty"`      // New set of category IDs
	TagIDs            []uuid.UUID              `json:"tag_ids,omitempty"`           // New set of tag IDs
	Amount            *string                  `json:"amount,omitempty"`            // New total amount
	Status            *entity.TransactionStatus `json:"status,omitempty"`            // New status (posted/cancelled)
	OccurredAt        *string                  `json:"occurred_at,omitempty"`       // Updated occurrence timestamp
	LineItems         *[]CreateLineItemRequest `json:"line_items,omitempty"`       // Complete replacement of line items
	Lang              string                   `json:"lang,omitempty"`              // Language for enriched data
}

// BatchPatchRequest is the payload for updating multiple transactions at once.
// Used in: TransactionHandler, TransactionService, TransactionInterface
type BatchPatchRequest struct {
	TransactionIDs []uuid.UUID             `json:"transaction_ids"` // List of transactions to update
	Patch          TransactionPatchRequest `json:"patch"`           // Changes to apply to all selected transactions
	Mode           *string                 `json:"mode,omitempty"`  // Update mode (e.g., "merge" or "replace" tags)
}

// BatchPatchResult represents the outcome of a batch update operation.
// Used in: TransactionHandler, TransactionService, TransactionInterface
type BatchPatchResult struct {
	Mode         string      `json:"mode"`                   // The update mode used
	UpdatedCount int         `json:"updated_count"`          // Number of successful updates
	FailedCount  int         `json:"failed_count"`           // Number of failed updates
	UpdatedIDs   []uuid.UUID `json:"updated_ids,omitempty"`  // IDs of successfully updated records
	FailedIDs    []uuid.UUID `json:"failed_ids,omitempty"`   // IDs of records that failed to update
}



// TransactionLineItemResponse represents a single line item of a processed transaction.
// Used in: TransactionHandler, TransactionService, TransactionInterface
type TransactionLineItemResponse struct {
	ID         uuid.UUID   `json:"id"`                    // Unique line item identifier
	CategoryID *uuid.UUID  `json:"category_id,omitempty"` // ID of the assigned category
	TagIDs     []uuid.UUID `json:"tag_ids,omitempty"`     // IDs of assigned tags
	Amount     string      `json:"amount"`               // Allocated amount
	Note       *string     `json:"note,omitempty"`        // Line item memo
}

// TransactionResponse represents a complete transaction record sent to the client.
// Used in: TransactionHandler, TransactionService, TransactionInterface
type TransactionResponse struct {
	ID              uuid.UUID                     `json:"id"`                             // Unique transaction identifier
	ExternalRef     *string                       `json:"external_ref,omitempty"`         // External system reference
	Type            entity.TransactionType        `json:"type"`                          // Income, Expense, or Transfer
	OccurredAt      time.Time                     `json:"occurred_at"`                    // Exact time of occurrence
	OccurredDate    string                        `json:"occurred_date"`                  // Date (YYYY-MM-DD)
	Amount          string                        `json:"amount"`                         // Transaction amount
	FromAmount      *string                       `json:"from_amount,omitempty"`          // Source amount (for transfers)
	ToAmount        *string                       `json:"to_amount,omitempty"`            // Destination amount (for transfers)
	AccountCurrency *string                       `json:"account_currency,omitempty"`     // Currency of the primary account
	FromCurrency    *string                       `json:"from_currency,omitempty"`        // Currency of source account
	ToCurrency      *string                       `json:"to_currency,omitempty"`          // Currency of destination account
	Description     *string                       `json:"description,omitempty"`          // Transaction description
	AccountID       *uuid.UUID                    `json:"account_id,omitempty"`           // Primary account (income/expense)
	FromAccountID   *uuid.UUID                    `json:"from_account_id,omitempty"`       // Source account (transfer)
	ToAccountID     *uuid.UUID                    `json:"to_account_id,omitempty"`         // Destination account (transfer)
	ExchangeRate    *string                       `json:"exchange_rate,omitempty"`        // Used exchange rate
	Status          entity.TransactionStatus      `json:"status"`                         // Current status
	LineItems       []TransactionLineItemResponse `json:"line_items,omitempty"`          // Detailed breakdown
	TagIDs          []uuid.UUID                   `json:"tag_ids,omitempty"`              // List of all tag IDs involved
	CategoryIDs     []uuid.UUID                   `json:"category_ids,omitempty"`          // List of all category IDs involved
	CategoryNames   []string                      `json:"category_names,omitempty"`        // Enriched category names
	TagNames        []string                      `json:"tag_names,omitempty"`            // Enriched tag names
	CategoryColors  []string                      `json:"category_colors,omitempty"`       // Enriched category UI colors
	TagColors       []string                      `json:"tag_colors,omitempty"`           // Enriched tag UI colors
}



// ListTransactionsRequest defines the filters and pagination for listing transactions.
// Used in: TransactionHandler, TransactionService, TransactionInterface
type ListTransactionsRequest struct {
	AccountID  *uuid.UUID              `json:"account_id"`  // Filter by primary account ID
	CategoryID *uuid.UUID              `json:"category_id"` // Filter by primary category ID
	Type       *entity.TransactionType `json:"type"`        // Filter by transaction type
	Search     *string                 `json:"search"`      // Full-text search on description
	From       *string                 `json:"from"`        // Start date filter (YYYY-MM-DD)
	To         *string                 `json:"to"`          // End date filter (YYYY-MM-DD)
	Page       int                     `json:"page"`        // Page number for pagination
	Limit      int                     `json:"limit"`       // Items per page
}

// ListTransactionsResponse represents a paginated list of transactions.
// Used in: TransactionHandler, TransactionService, TransactionInterface
type ListTransactionsResponse struct {
	Data       []TransactionResponse `json:"data"`                  // List of transactions for current page
	NextCursor *string               `json:"next_cursor,omitempty"`  // Cursor for key-based pagination (if used)
	TotalCount int                   `json:"total_count"`           // Overall total matching records
	TotalPages int                   `json:"total_pages,omitempty"`  // Total pages available
	Page       int                   `json:"page,omitempty"`         // Current page number
	Limit      int                   `json:"limit,omitempty"`        // Page size used
}
