// Package category provides category management functionality.
package category

import (
	"context"

	"github.com/sonbn-225/goen-api/internal/domain"
)

// Module represents the category module.
type Module struct {
	Service *Service
	Handler *Handler
}

// ModuleDeps contains dependencies for the category module.
type ModuleDeps struct {
	Repo domain.CategoryRepository
}

// NewModule creates a new category module.
func NewModule(deps ModuleDeps) *Module {
	svc := NewService(deps.Repo)
	h := NewHandler(svc)

	return &Module{
		Service: svc,
		Handler: h,
	}
}

// ServiceInterface defines the category service contract.
type ServiceInterface interface {
	Get(ctx context.Context, userID, categoryID string) (*domain.Category, error)
	List(ctx context.Context, userID string) ([]domain.Category, error)
}
