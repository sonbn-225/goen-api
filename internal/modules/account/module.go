// Package account provides account management functionality.
package account

import (
	"github.com/sonbn-225/goen-api/internal/domain"
	"github.com/sonbn-225/goen-api/internal/storage"
)

// Module represents the account module.
type Module struct {
	Service *Service
	Handler *Handler
}

// ModuleDeps contains dependencies for the account module.
type ModuleDeps struct {
	DB           *storage.Postgres
	AccountRepo  domain.AccountRepository
	UserRepo     domain.UserRepository
	AuditService AuditServiceInterface
}

// AuditServiceInterface defines the audit service contract needed by account handlers.
type AuditServiceInterface interface {
	ListAuditEvents(ctx interface{}, userID, accountID string, limit int) ([]domain.AuditEvent, error)
}

// NewModule creates a new account module.
func NewModule(deps ModuleDeps) *Module {
	svc := NewService(deps.AccountRepo, deps.UserRepo)
	h := NewHandler(svc, deps.AuditService)

	return &Module{
		Service: svc,
		Handler: h,
	}
}
