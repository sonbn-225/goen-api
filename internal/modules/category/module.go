// Package category provides category management functionality.
package category

import (
	"github.com/sonbn-225/goen-api/internal/domain"
	"github.com/sonbn-225/goen-api/internal/storage"
)

// Module represents the category module.
type Module struct {
	Service *Service
	Handler *Handler
}

// ModuleDeps contains dependencies for the category module.
type ModuleDeps struct {
	DB *storage.Postgres
}

// NewModule creates a new category module.
func NewModule(deps ModuleDeps) *Module {
	repo := storage.NewCategoryRepo(deps.DB)
	svc := NewService(repo)
	h := NewHandler(svc)

	return &Module{
		Service: svc,
		Handler: h,
	}
}

// ServiceInterface defines the category service contract.
type ServiceInterface interface {
	Get(ctx interface{}, userID, categoryID string) (*domain.Category, error)
	List(ctx interface{}, userID string) ([]domain.Category, error)
}
