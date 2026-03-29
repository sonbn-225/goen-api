// Package contact provides contact management functionality.
package contact

import (
	"github.com/sonbn-225/goen-api/internal/domain"
)

// Module represents the contact module.
type Module struct {
	Service *Service
	Handler *Handler
}

// ModuleDeps contains dependencies for the contact module.
type ModuleDeps struct {
	Repo domain.ContactRepository
}

// NewModule creates a new contact module.
func NewModule(deps ModuleDeps) *Module {
	svc := NewService(deps.Repo)
	h := NewHandler(svc)

	return &Module{
		Service: svc,
		Handler: h,
	}
}
