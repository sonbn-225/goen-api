package profile

import (
	"context"

	"github.com/sonbn-225/goen-api-v2/internal/domains/auth"
)

type Service interface {
	GetMe(ctx context.Context, userID string) (*auth.User, error)
	UpdateMyProfile(ctx context.Context, userID string, input auth.UpdateProfileInput) (*auth.User, error)
	UploadAvatar(ctx context.Context, userID, fileName, contentType string, raw []byte) (*auth.User, error)
	ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error
}

type AvatarStorage interface {
	UploadAvatar(ctx context.Context, userID, fileName, contentType string, data []byte) (string, error)
}

type ModuleDeps struct {
	Service Service
}

type Module struct {
	Handler *Handler
}
