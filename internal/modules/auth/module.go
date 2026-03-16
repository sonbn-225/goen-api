// Package auth provides authentication and user management functionality.
package auth

import (
	"context"

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
	UserRepo domain.UserRepository
	Config   *config.Config
	S3Client *storage.S3Client
}

// NewModule creates a new auth module with all dependencies wired.
func NewModule(deps ModuleDeps) *Module {
	svc := NewService(deps.UserRepo, deps.Config, deps.S3Client)
	h := NewHandler(svc)

	return &Module{
		Service: svc,
		Handler: h,
	}
}

// ServiceInterface defines the auth service contract for external modules.
type ServiceInterface interface {
	Signup(ctx context.Context, req SignupRequest) (*AuthResponse, error)
	Signin(ctx context.Context, req SigninRequest) (*AuthResponse, error)
	Refresh(ctx context.Context, userID string) (*AuthResponse, error)
	GetMe(ctx context.Context, userID string) (*domain.User, error)
	UpdateMySettings(ctx context.Context, userID string, patch map[string]any) (*domain.User, error)
}

