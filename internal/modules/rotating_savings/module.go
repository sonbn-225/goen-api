// Package rotatingsavings provides rotating savings group (hụi/họ) management.
package rotatingsavings

import (
	"context"

	"github.com/sonbn-225/goen-api/internal/domain"
	"github.com/sonbn-225/goen-api/internal/storage"
)

// Module represents the rotating savings module.
type Module struct {
	Service *Service
	Handler *Handler
}

// TransactionServiceInterface defines the transaction service contract.
type TransactionServiceInterface interface {
	Create(ctx context.Context, userID string, req interface{}) (*domain.Transaction, error)
}

// ModuleDeps contains dependencies for the rotating savings module.
type ModuleDeps struct {
	DB          *storage.Postgres
	Repo        domain.RotatingSavingsRepository
	AccountRepo domain.AccountRepository
	TxSvc       TransactionServiceInterface
}

// NewModule creates a new rotating savings module.
func NewModule(deps ModuleDeps) *Module {
	svc := NewService(deps.AccountRepo, deps.TxSvc, deps.Repo)
	h := NewHandler(svc)

	return &Module{
		Service: svc,
		Handler: h,
	}
}
