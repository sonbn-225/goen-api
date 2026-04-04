package account

import (
	"context"
	"time"

	"github.com/sonbn-225/goen-api-v2/internal/core/money"
)

type Account struct {
	ID              string       `json:"id"`
	UserID          string       `json:"user_id"`
	Name            string       `json:"name"`
	Type            string       `json:"account_type"`
	Currency        string       `json:"currency"`
	ParentAccountID *string      `json:"parent_account_id,omitempty"`
	AccountNumber   *string      `json:"account_number,omitempty"`
	Color           *string      `json:"color,omitempty"`
	Status          string       `json:"status"`
	ClosedAt        *time.Time   `json:"closed_at,omitempty"`
	Balance         money.Amount `json:"balance"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

type CreateInput struct {
	Name            string `json:"name"`
	Type            string `json:"type"`
	Currency        string `json:"currency"`
	ParentAccountID string `json:"parent_account_id"`
	AccountNumber   string `json:"account_number"`
	Color           string `json:"color"`
}

type Repository interface {
	Create(ctx context.Context, account *Account) error
	ListByUser(ctx context.Context, userID string) ([]Account, error)
	GetByID(ctx context.Context, userID, accountID string) (*Account, error)
	GetDefaultCurrency(ctx context.Context, userID string) (string, error)
	IsOwner(ctx context.Context, userID, accountID string) (bool, error)
	HasRelatedTransferTransactionsForAccount(ctx context.Context, accountID string) (bool, error)
	Delete(ctx context.Context, userID, accountID string) (bool, error)
}

type Service interface {
	Create(ctx context.Context, userID string, input CreateInput) (*Account, error)
	List(ctx context.Context, userID string) ([]Account, error)
	Get(ctx context.Context, userID, accountID string) (*Account, error)
	Delete(ctx context.Context, userID, accountID string) error
}

type ModuleDeps struct {
	Repo    Repository
	Service Service
}

type Module struct {
	Service Service
	Handler *Handler
}
