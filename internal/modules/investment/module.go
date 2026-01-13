// Package investment provides investment-related functionality including
// investment accounts, securities, trades, holdings, and market data.
//
// This module follows the feature-based structure pattern where each module
// contains its own handler, service, and repository layers.
//
// Usage:
//
//	deps := investment.ModuleDeps{
//	    DB:                 postgresDB,
//	    AccountService:     accountSvc,
//	    TransactionService: txSvc,
//	    Config:             cfg,
//	    Redis:              redis,
//	}
//	mod := investment.NewModule(deps)
//	// Use mod.Handler for HTTP routes
//	// Use mod.Service for business logic
package investment

import (
	"context"

	"github.com/sonbn-225/goen-api/internal/config"
	"github.com/sonbn-225/goen-api/internal/domain"
	"github.com/sonbn-225/goen-api/internal/modules/transaction"
	"github.com/sonbn-225/goen-api/internal/storage"
)

// Module represents the investment module with all its dependencies.
type Module struct {
	Service *Service
	Handler *Handler
}

// ModuleDeps contains external dependencies required by the investment module.
type ModuleDeps struct {
	Repo               domain.InvestmentRepository
	Redis              *storage.Redis
	Config             *config.Config
	AccountService     AccountServiceDep
	TransactionService TransactionServiceDep
}

// AccountServiceDep defines the account service methods needed by this module.
type AccountServiceDep interface {
	GetAccountByID(ctx context.Context, userID, accountID string) (*domain.Account, error)
}

// TransactionServiceDep defines the transaction service methods needed by this module.
type TransactionServiceDep interface {
	Create(ctx context.Context, userID string, req transaction.CreateRequest) (*domain.Transaction, error)
}

// NewModule creates a new investment module with all dependencies wired.
func NewModule(deps ModuleDeps) *Module {
	svc := NewService(deps.Repo, deps.AccountService, deps.TransactionService, deps.Config, deps.Redis)
	h := NewHandler(svc)

	return &Module{
		Service: svc,
		Handler: h,
	}
}
