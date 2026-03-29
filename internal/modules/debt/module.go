// Package debt provides debt management functionality.
package debt

import (
	"context"

	"github.com/sonbn-225/goen-api/internal/domain"
	"github.com/sonbn-225/goen-api/internal/modules/contact"
)

// Module represents the debt module.
type Module struct {
	Service *Service
	Handler *Handler
}

// TransactionServiceInterface defines the transaction service contract needed by debt.
type TransactionServiceInterface interface {
	Get(ctx context.Context, userID, transactionID string) (*domain.Transaction, error)
}

// ModuleDeps contains dependencies for the debt module.
type ModuleDeps struct {
	Repo       domain.DebtRepository
	TxSvc      TransactionServiceInterface
	ContactSvc *contact.Service
}

// NewModule creates a new debt module.
func NewModule(deps ModuleDeps) *Module {
	svc := NewService(deps.TxSvc, deps.Repo, deps.ContactSvc)
	h := NewHandler(svc)

	return &Module{
		Service: svc,
		Handler: h,
	}
}

