package domain

import (
	"context"
	"time"
)

type TransactionLineItem struct {
	ID            string   `json:"id"`
	CategoryID    *string  `json:"category_id,omitempty"`
	TagIDs        []string `json:"tag_ids,omitempty"`
	Amount        string   `json:"amount"`
	Note          *string  `json:"note,omitempty"`
	TransactionID string   `json:"-"`
}

type Transaction struct {
	ID           string    `json:"id"`
	ClientID     *string   `json:"client_id,omitempty"`
	ExternalRef  *string   `json:"external_ref,omitempty"`
	Type         string    `json:"type"`
	OccurredAt   time.Time `json:"occurred_at"`
	OccurredDate string    `json:"occurred_date"`
	Amount       string    `json:"amount"`
	// For FX transfers, amounts can differ per side.
	FromAmount *string `json:"from_amount,omitempty"`
	ToAmount   *string `json:"to_amount,omitempty"`
	// Currencies are derived from linked accounts (not stored on transactions).
	AccountCurrency *string               `json:"account_currency,omitempty"`
	FromCurrency    *string               `json:"from_currency,omitempty"`
	ToCurrency      *string               `json:"to_currency,omitempty"`
	Description     *string               `json:"description,omitempty"`
	AccountID       *string               `json:"account_id,omitempty"`
	FromAccountID   *string               `json:"from_account_id,omitempty"`
	ToAccountID     *string               `json:"to_account_id,omitempty"`
	ExchangeRate    *string               `json:"exchange_rate,omitempty"`
	Status          string                `json:"status"`
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
	CreatedBy       *string               `json:"created_by,omitempty"`
	UpdatedBy       *string               `json:"updated_by,omitempty"`
	DeletedAt       *time.Time            `json:"deleted_at,omitempty"`
	LineItems       []TransactionLineItem `json:"line_items,omitempty"`
	TagIDs          []string              `json:"tag_ids,omitempty"`
	CategoryIDs     []string              `json:"category_ids,omitempty"`
}

type TransactionListFilter struct {
	AccountID         *string
	CategoryID        *string
	Type              *string
	Search            *string
	ExternalRefFamily *string
	From              *time.Time
	To                *time.Time
	Cursor            *string
	Page              int
	Limit             int
}

type ImportedTransaction struct {
	ID                   string         `json:"id"`
	UserID               string         `json:"-"`
	Source               string         `json:"source"`
	TransactionDate      string         `json:"transaction_date"`
	Amount               string         `json:"amount"`
	Description          *string        `json:"description,omitempty"`
	TransactionType      *string        `json:"transaction_type,omitempty"`
	ImportedAccountName  *string        `json:"imported_account_name,omitempty"`
	ImportedCategoryName *string        `json:"imported_category_name,omitempty"`
	MappedAccountID      *string        `json:"mapped_account_id,omitempty"`
	MappedCategoryID     *string        `json:"mapped_category_id,omitempty"`
	RawPayload           map[string]any `json:"raw_payload,omitempty"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
}

type ImportedTransactionCreate struct {
	Source               string
	TransactionDate      string
	Amount               string
	Description          *string
	TransactionType      *string
	ImportedAccountName  *string
	ImportedCategoryName *string
	MappedAccountID      *string
	MappedCategoryID     *string
	RawPayload           map[string]any
}

type ImportedTransactionPatch struct {
	MappedAccountID  *string
	MappedCategoryID *string
}

type ImportMappingRule struct {
	ID         string    `json:"id"`
	UserID     string    `json:"-"`
	Kind       string    `json:"kind"` // account | category
	SourceName string    `json:"source_name"`
	MappedID   string    `json:"mapped_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ImportMappingRuleUpsert struct {
	Kind       string
	SourceName string
	MappedID   string
}

// ExportTransactionRow represents a transaction row for CSV export
// (compatible with generic import format)
type ExportTransactionRow struct {
	TransactionID   string  `json:"transaction_id"`
	TransactionDate string  `json:"transaction_date"`
	Amount          string  `json:"amount"`
	Description     *string `json:"description,omitempty"`
	TransactionType string  `json:"transaction_type"` // "expense", "income", "transfer", etc.
	AccountName     string  `json:"account_name"`
	AccountID       string  `json:"account_id"`
	CategoryName    *string `json:"category_name,omitempty"`
	CategoryID      *string `json:"category_id,omitempty"`
	// For transfers
	FromAccountName *string `json:"from_account_name,omitempty"`
	FromAccountID   *string `json:"from_account_id,omitempty"`
	ToAccountName   *string `json:"to_account_name,omitempty"`
	ToAccountID     *string `json:"to_account_id,omitempty"`
}

type TransactionRepository interface {
	CreateTransaction(ctx context.Context, userID string, tx Transaction, lineItems []TransactionLineItem, tagIDs []string, participants []GroupExpenseParticipant) error
	GetTransaction(ctx context.Context, userID string, transactionID string) (*Transaction, error)
	GetImportedTransaction(ctx context.Context, userID string, importID string) (*ImportedTransaction, error)
	ListTransactions(ctx context.Context, userID string, filter TransactionListFilter) ([]Transaction, *string, int, error)
	PatchTransaction(ctx context.Context, userID string, transactionID string, patch TransactionPatch) (*Transaction, error)
	BatchPatchTransactions(ctx context.Context, userID string, transactionIDs []string, patches map[string]TransactionPatch, mode string) ([]string, []string, error)
	DeleteTransaction(ctx context.Context, userID string, transactionID string) error
	CreateImportedTransactions(ctx context.Context, userID string, items []ImportedTransactionCreate) ([]ImportedTransaction, error)
	ListImportedTransactions(ctx context.Context, userID string) ([]ImportedTransaction, error)
	PatchImportedTransaction(ctx context.Context, userID string, importID string, patch ImportedTransactionPatch) (*ImportedTransaction, error)
	DeleteImportedTransaction(ctx context.Context, userID string, importID string) error
	DeleteAllImportedTransactions(ctx context.Context, userID string) (int64, error)
	UpsertImportMappingRules(ctx context.Context, userID string, rules []ImportMappingRuleUpsert) ([]ImportMappingRule, error)
	ListImportMappingRules(ctx context.Context, userID string) ([]ImportMappingRule, error)
	DeleteImportMappingRule(ctx context.Context, userID string, ruleID string) error
}

type TransactionPatch struct {
	Description       *string
	CategoryIDs       []string
	TagIDs            []string
	Amount            *string
	Status            *string
	OccurredAt        *time.Time
	LineItems         *[]TransactionLineItem     // nil = no change, non-nil = replace all
	GroupParticipants *[]GroupExpenseParticipant // nil = no change, non-nil = replace all (only unsettled)
}
