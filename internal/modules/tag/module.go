// Package tag provides tag management functionality.
package tag

import (
	"github.com/sonbn-225/goen-api/internal/domain"
	"github.com/sonbn-225/goen-api/internal/storage"
)

// Module represents the tag module.
type Module struct {
	Service *Service
	Handler *Handler
}

// ModuleDeps contains dependencies for the tag module.
type ModuleDeps struct {
	DB *storage.Postgres
}

// NewModule creates a new tag module.
func NewModule(deps ModuleDeps) *Module {
	repo := storage.NewTagRepo(deps.DB)
	svc := NewService(repo)
	h := NewHandler(svc)

	return &Module{
		Service: svc,
		Handler: h,
	}
}

// ServiceInterface defines the tag service contract.
type ServiceInterface interface {
	Create(ctx interface{}, userID string, req CreateTagRequest) (*domain.Tag, error)
	Get(ctx interface{}, userID, tagID string) (*domain.Tag, error)
	List(ctx interface{}, userID string) ([]domain.Tag, error)
}
