package setting

import (
	"context"

	"github.com/sonbn-225/goen-api-v2/internal/domains/auth"
)

type Service interface {
	UpdateMySettings(ctx context.Context, userID string, patch map[string]any) (*auth.User, error)
}

type ModuleDeps struct {
	Service Service
}

type Module struct {
	Handler *Handler
}
