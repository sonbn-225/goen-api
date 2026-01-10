// Package marketdata provides market data sync and status endpoints.
package marketdata

import (
	"github.com/sonbn-225/goen-api/internal/config"
	"github.com/sonbn-225/goen-api/internal/storage"
)

// InvestmentServiceInterface defines investment service contract.
type InvestmentServiceInterface interface {
	GetSecurity(ctx interface{}, userID, securityID string) (interface{}, error)
}

// Module represents the market data module.
type Module struct {
	Service *Service
	Handler *Handler
}

// ModuleDeps contains dependencies for the market data module.
type ModuleDeps struct {
	Cfg       *config.Config
	DB        *storage.Postgres
	Redis     *storage.Redis
	InvestSvc InvestmentServiceInterface
}

// NewModule creates a new market data module.
func NewModule(deps ModuleDeps) *Module {
	svc := NewService(deps.Cfg, deps.DB, deps.Redis, deps.InvestSvc)
	h := NewHandler(svc)

	return &Module{
		Service: svc,
		Handler: h,
	}
}
