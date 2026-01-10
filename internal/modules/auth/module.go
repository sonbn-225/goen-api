// Package auth provides authentication and user management functionality.
package auth

import (
	"github.com/sonbn-225/goen-api/internal/config"
	"github.com/sonbn-225/goen-api/internal/domain"
	"github.com/sonbn-225/goen-api/internal/storage"
)

// Module represents the auth module with all its dependencies.
type Module struct {
	Service *Service
	Handler *Handler
}

// ModuleDeps contains external dependencies required by the auth module.
type ModuleDeps struct {
	DB     *storage.Postgres
	Config *config.Config
}

// NewModule creates a new auth module with all dependencies wired.
func NewModule(deps ModuleDeps) *Module {
	repo := storage.NewUserRepo(deps.DB)
	svc := NewService(repo, deps.Config)
	h := NewHandler(svc)

	return &Module{
		Service: svc,
		Handler: h,
	}
}

// ServiceInterface defines the auth service contract for external modules.
type ServiceInterface interface {
	Signup(ctx interface{}, req SignupRequest) (*AuthResponse, error)
	Signin(ctx interface{}, req SigninRequest) (*AuthResponse, error)
	GetMe(ctx interface{}, userID string) (*domain.User, error)
	UpdateMySettings(ctx interface{}, userID string, patch map[string]any) (*domain.User, error)
}
