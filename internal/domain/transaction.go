package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrTransactionNotFound  = errors.New("transaction not found")
	ErrTransactionForbidden = errors.New("transaction forbidden")
)

type TransactionLineItem struct {
	ID            string  `json:"id"`
	CategoryID    *string `json:"category_id,omitempty"`
	Amount        string  `json:"amount"`
	Note          *string `json:"note,omitempty"`
	TransactionID string  `json:"-"`
}

type Transaction struct {
	ID           string  `json:"id"`
	ClientID     *string `json:"client_id,omitempty"`
	ExternalRef  *string `json:"external_ref,omitempty"`
	Type         string  `json:"type"`
	OccurredAt   time.Time `json:"occurred_at"`
	OccurredDate string  `json:"occurred_date"`
	Amount       string  `json:"amount"`
	// For FX transfers, amounts can differ per side.
	FromAmount   *string `json:"from_amount,omitempty"`
	ToAmount     *string `json:"to_amount,omitempty"`
	// Currencies are derived from linked accounts (not stored on transactions).
	AccountCurrency *string `json:"account_currency,omitempty"`
	FromCurrency    *string `json:"from_currency,omitempty"`
	ToCurrency      *string `json:"to_currency,omitempty"`
	Description  *string `json:"description,omitempty"`
	AccountID    *string `json:"account_id,omitempty"`
	FromAccountID *string `json:"from_account_id,omitempty"`
	ToAccountID   *string `json:"to_account_id,omitempty"`
	ExchangeRate *string `json:"exchange_rate,omitempty"`
	Counterparty *string `json:"counterparty,omitempty"`
	Notes        *string `json:"notes,omitempty"`
	Status       string  `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	CreatedBy    *string `json:"created_by,omitempty"`
	UpdatedBy    *string `json:"updated_by,omitempty"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
	LineItems    []TransactionLineItem `json:"line_items,omitempty"`
	TagIDs       []string `json:"tag_ids,omitempty"`
	CategoryIDs  []string `json:"category_ids,omitempty"`
}

type TransactionListFilter struct {
	AccountID *string
	From      *time.Time
	To        *time.Time
	Cursor    *string
	Limit     int
}

type TransactionRepository interface {
	CreateTransaction(ctx context.Context, userID string, tx Transaction, lineItems []TransactionLineItem, tagIDs []string) error
	GetTransaction(ctx context.Context, userID string, transactionID string) (*Transaction, error)
	ListTransactions(ctx context.Context, userID string, filter TransactionListFilter) ([]Transaction, *string, error)
	PatchTransaction(ctx context.Context, userID string, transactionID string, patch TransactionPatch) (*Transaction, error)
	DeleteTransaction(ctx context.Context, userID string, transactionID string) error
}

type TransactionPatch struct {
	Description  *string
	Notes        *string
	Counterparty *string
}
