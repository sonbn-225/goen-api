package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrAccountNotFound     = errors.New("account not found")
	ErrAccountForbidden    = errors.New("account forbidden")
	ErrAccountInvalidInput = errors.New("invalid account input")
)

type Account struct {
	ID              string     `json:"id"`
	ClientID        *string    `json:"client_id,omitempty"`
	Name            string     `json:"name"`
	AccountType     string     `json:"account_type"`
	Currency        string     `json:"currency"`
	ParentAccountID *string    `json:"parent_account_id,omitempty"`
	Status          string     `json:"status"`
	ClosedAt        *time.Time `json:"closed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	CreatedBy       *string    `json:"created_by,omitempty"`
	UpdatedBy       *string    `json:"updated_by,omitempty"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
}

type AccountRepository interface {
	CreateAccountWithOwner(ctx context.Context, account Account, ownerUserID string) error
	ListAccountsForUser(ctx context.Context, userID string) ([]Account, error)
	GetAccountForUser(ctx context.Context, userID string, accountID string) (*Account, error)
}
