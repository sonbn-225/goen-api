// Package transaction provides transaction management functionality.
package transaction

import (
	"github.com/sonbn-225/goen-api/internal/domain"
)

// Module represents the transaction module.
type Module struct {
	Service *Service
	Handler *Handler
}

// ModuleDeps contains dependencies for the transaction module.
type ModuleDeps struct {
	Repo domain.TransactionRepository
}

// NewModule creates a new transaction module.
func NewModule(deps ModuleDeps) *Module {
	svc := NewService(deps.Repo)
	h := NewHandler(svc)

	return &Module{
		Service: svc,
		Handler: h,
	}
}
