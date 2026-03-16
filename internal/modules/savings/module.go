// Package savings provides savings instrument management functionality.
package savings

import (
	"context"

	"github.com/sonbn-225/goen-api/internal/domain"
	"github.com/sonbn-225/goen-api/internal/modules/account"
	"github.com/sonbn-225/goen-api/internal/modules/transaction"
)

// Module represents the savings module.
type Module struct {
	Service *Service
	Handler *Handler
}

// AccountServiceInterface defines the account service contract needed by savings.
type AccountServiceInterface interface {
	Get(ctx context.Context, userID, accountID string) (*domain.Account, error)
	Create(ctx context.Context, userID string, req account.CreateAccountRequest) (*domain.Account, error)
	Delete(ctx context.Context, userID, accountID string) error
}

// TransactionServiceInterface defines the transaction service contract needed by savings.
type TransactionServiceInterface interface {
	Create(ctx context.Context, userID string, req transaction.CreateRequest) (*domain.Transaction, error)
}

// ModuleDeps contains dependencies for the savings module.
type ModuleDeps struct {
	Repo       domain.SavingsRepository
	AccountSvc AccountServiceInterface
	TxSvc      TransactionServiceInterface
}

// NewModule creates a new savings module.
func NewModule(deps ModuleDeps) *Module {
	svc := NewService(deps.AccountSvc, deps.TxSvc, deps.Repo)
	h := NewHandler(svc)

	return &Module{
		Service: svc,
		Handler: h,
	}
}

