// Package budget provides budget management functionality.
package budget

import (
	"github.com/sonbn-225/goen-api/internal/domain"
)

// Module represents the budget module.
type Module struct {
	Service *Service
	Handler *Handler
}

// ModuleDeps contains dependencies for the budget module.
type ModuleDeps struct {
	BudgetRepo   domain.BudgetRepository
	CategoryRepo domain.CategoryRepository
}

// NewModule creates a new budget module.
func NewModule(deps ModuleDeps) *Module {
	svc := NewService(deps.BudgetRepo, deps.CategoryRepo)
	h := NewHandler(svc)

	return &Module{
		Service: svc,
		Handler: h,
	}
}

