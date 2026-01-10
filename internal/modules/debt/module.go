// Package debt provides debt management functionality.
package debt

import (
	"github.com/sonbn-225/goen-api/internal/domain"
	"github.com/sonbn-225/goen-api/internal/storage"
)

// Module represents the debt module.
type Module struct {
	Service *Service
	Handler *Handler
}

// TransactionServiceInterface defines the transaction service contract needed by debt.
type TransactionServiceInterface interface {
	Get(ctx interface{}, userID, transactionID string) (*domain.Transaction, error)
}

// ModuleDeps contains dependencies for the debt module.
type ModuleDeps struct {
	DB    *storage.Postgres
	Repo  domain.DebtRepository
	TxSvc TransactionServiceInterface
}

// NewModule creates a new debt module.
func NewModule(deps ModuleDeps) *Module {
	svc := NewService(deps.TxSvc, deps.Repo)
	h := NewHandler(svc)

	return &Module{
		Service: svc,
		Handler: h,
	}
}
