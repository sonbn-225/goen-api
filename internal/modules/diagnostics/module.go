// Package diagnostics provides health and connectivity endpoints.
package diagnostics

import (
	"github.com/sonbn-225/goen-api/internal/config"
	"github.com/sonbn-225/goen-api/internal/storage"
)

// Module represents the diagnostics module.
type Module struct {
	Service *Service
	Handler *Handler
}

// ModuleDeps contains dependencies for the diagnostics module.
type ModuleDeps struct {
	Cfg   *config.Config
	DB    *storage.Postgres
	Redis *storage.Redis
}

// NewModule creates a new diagnostics module.
func NewModule(deps ModuleDeps) *Module {
	svc := NewService(deps.DB, deps.Redis)
	h := NewHandler(svc, deps.Cfg)

	return &Module{
		Service: svc,
		Handler: h,
	}
}
