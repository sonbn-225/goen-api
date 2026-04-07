package entity

import (
	"github.com/google/uuid"
)

type ImportedTransaction struct {
	AuditEntity
	UserID               uuid.UUID      `json:"-"`
	Source               string         `json:"source"`
	TransactionDate      string         `json:"transaction_date"`
	Amount               string         `json:"amount"`
	Description          *string        `json:"description,omitempty"`
	TransactionType      *string        `json:"transaction_type,omitempty"`
	ImportedAccountName  *string        `json:"imported_account_name,omitempty"`
	ImportedCategoryName *string        `json:"imported_category_name,omitempty"`
	MappedAccountID      *uuid.UUID     `json:"mapped_account_id,omitempty"`
	MappedCategoryID     *uuid.UUID     `json:"mapped_category_id,omitempty"`
	RawPayload           map[string]any `json:"raw_payload,omitempty"`
}

type ImportMappingRuleKind string

const (
	ImportMappingRuleKindAccount  ImportMappingRuleKind = "account"
	ImportMappingRuleKindCategory ImportMappingRuleKind = "category"
)

type ImportMappingRule struct {
	AuditEntity
	UserID     uuid.UUID             `json:"-"`
	Kind       ImportMappingRuleKind `json:"kind"`
	SourceName string                `json:"source_name"`
	MappedID   uuid.UUID             `json:"mapped_id"`
}

type ImportedTransactionCreate struct {
	Source               string
	TransactionDate      string
	Amount               string
	Description          *string
	TransactionType      *string
	ImportedAccountName  *string
	ImportedCategoryName *string
	MappedAccountID      *uuid.UUID
	MappedCategoryID     *uuid.UUID
	RawPayload           map[string]any
}

type ImportedTransactionPatch struct {
	MappedAccountID  *uuid.UUID
	MappedCategoryID *uuid.UUID
}

type ImportMappingRuleUpsert struct {
	Kind       ImportMappingRuleKind
	SourceName string
	MappedID   uuid.UUID
}

type ExportTransactionRow struct {
	TransactionID   uuid.UUID  `json:"transaction_id"`
	TransactionDate string     `json:"transaction_date"`
	Amount          string     `json:"amount"`
	Description     *string    `json:"description,omitempty"`
	TransactionType string     `json:"transaction_type"`
	AccountName     string     `json:"account_name"`
	AccountID       uuid.UUID  `json:"account_id"`
	CategoryName    *string    `json:"category_name,omitempty"`
	CategoryID      *uuid.UUID `json:"category_id,omitempty"`
	FromAccountName *string    `json:"from_account_name,omitempty"`
	FromAccountID   *uuid.UUID `json:"from_account_id,omitempty"`
	ToAccountName   *string    `json:"to_account_name,omitempty"`
	ToAccountID     *uuid.UUID `json:"to_account_id,omitempty"`
}

