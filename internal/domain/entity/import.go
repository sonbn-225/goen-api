package entity

import (
	"time"
)

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

type ImportMappingRule struct {
	ID         string    `json:"id"`
	UserID     string    `json:"-"`
	Kind       string    `json:"kind"` // account | category
	SourceName string    `json:"source_name"`
	MappedID   string    `json:"mapped_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
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

type ImportMappingRuleUpsert struct {
	Kind       string
	SourceName string
	MappedID   string
}

type ExportTransactionRow struct {
	TransactionID   string  `json:"transaction_id"`
	TransactionDate string  `json:"transaction_date"`
	Amount          string  `json:"amount"`
	Description     *string `json:"description,omitempty"`
	TransactionType string  `json:"transaction_type"`
	AccountName     string  `json:"account_name"`
	AccountID       string  `json:"account_id"`
	CategoryName    *string `json:"category_name,omitempty"`
	CategoryID      *string `json:"category_id,omitempty"`
	FromAccountName *string `json:"from_account_name,omitempty"`
	FromAccountID   *string `json:"from_account_id,omitempty"`
	ToAccountName   *string `json:"to_account_name,omitempty"`
	ToAccountID     *string `json:"to_account_id,omitempty"`
}
