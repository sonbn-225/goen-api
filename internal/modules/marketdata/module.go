// Package marketdata provides market data sync and status endpoints.
package marketdata

import (
	"context"

	"github.com/sonbn-225/goen-api/internal/config"
	"github.com/sonbn-225/goen-api/internal/domain"
	"github.com/sonbn-225/goen-api/internal/storage"
)

// InvestmentServiceInterface defines investment service contract.
type InvestmentServiceInterface interface {
	GetSecurity(ctx context.Context, securityID string) (*domain.Security, error)
}

// Repository defines the market data persistence contract.
// It is used by the service to avoid depending on concrete DB types.
type Repository interface {
	LoadSecurityIDsBySymbols(ctx context.Context, symbols []string) (map[string]string, error)
	LoadSyncState(ctx context.Context, syncKey string) (*SyncState, error)
}

// Module represents the market data module.
type Module struct {
	Service *Service
	Handler *Handler
}

// ModuleDeps contains dependencies for the market data module.
type ModuleDeps struct {
	Cfg       *config.Config
	Redis     *storage.Redis
	Repo      Repository
	InvestSvc InvestmentServiceInterface
}

// NewModule creates a new market data module.
func NewModule(deps ModuleDeps) *Module {
	svc := NewService(deps.Cfg, deps.Repo, deps.Redis, deps.InvestSvc)
	h := NewHandler(svc)

	return &Module{
		Service: svc,
		Handler: h,
	}
}
