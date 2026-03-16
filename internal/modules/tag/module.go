// Package tag provides tag management functionality.
package tag

import (
	"context"

	"github.com/sonbn-225/goen-api/internal/domain"
)

// Module represents the tag module.
type Module struct {
	Service *Service
	Handler *Handler
}

// ModuleDeps contains dependencies for the tag module.
type ModuleDeps struct {
	Repo domain.TagRepository
}

// NewModule creates a new tag module.
func NewModule(deps ModuleDeps) *Module {
	svc := NewService(deps.Repo)
	h := NewHandler(svc)

	return &Module{
		Service: svc,
		Handler: h,
	}
}

// ServiceInterface defines the tag service contract.
type ServiceInterface interface {
	Create(ctx context.Context, userID string, req CreateTagRequest) (*domain.Tag, error)
	Get(ctx context.Context, userID, tagID string) (*domain.Tag, error)
	List(ctx context.Context, userID string) ([]domain.Tag, error)
}

