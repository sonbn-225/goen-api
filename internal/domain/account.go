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

type AccountPatch struct {
	Name   *string `json:"name,omitempty"`
	Status *string `json:"status,omitempty"`
}

type AccountBalance struct {
	AccountID string `json:"account_id"`
	Currency  string `json:"currency"`
	Balance   string `json:"balance"`
}

type AccountRepository interface {
	CreateAccountWithOwner(ctx context.Context, account Account, ownerUserID string) error
	ListAccountsForUser(ctx context.Context, userID string) ([]Account, error)
	GetAccountForUser(ctx context.Context, userID string, accountID string) (*Account, error)
	PatchAccount(ctx context.Context, actorUserID string, accountID string, patch AccountPatch) (*Account, error)
	DeleteAccount(ctx context.Context, actorUserID string, accountID string) error
	ListAccountBalancesForUser(ctx context.Context, userID string) ([]AccountBalance, error)

	// UC-007 Shared Account
	AccountShareRepository
}
