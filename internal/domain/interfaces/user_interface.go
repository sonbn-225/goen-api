package interfaces

import (
	"context"
	"mime/multipart"

	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user entity.UserWithPassword) error
	FindUserByEmail(ctx context.Context, email string) (*entity.UserWithPassword, error)
	FindUserByPhone(ctx context.Context, phone string) (*entity.UserWithPassword, error)
	FindUserByUsername(ctx context.Context, username string) (*entity.UserWithPassword, error)
	FindUserByID(ctx context.Context, id string) (*entity.User, error)
	UpdateUserSettings(ctx context.Context, userID string, patch map[string]any) (*entity.User, error)
	UpdateUserProfile(ctx context.Context, userID string, params entity.UpdateUserParams) (*entity.User, error)
}

type AuthService interface {
	Signup(ctx context.Context, req dto.SignupRequest) (*dto.AuthResponse, error)
	Signin(ctx context.Context, req dto.SigninRequest) (*dto.AuthResponse, error)
	Refresh(ctx context.Context, userID string) (*dto.AuthResponse, error)
	GetMe(ctx context.Context, userID string) (*entity.User, error)
	UpdateMySettings(ctx context.Context, userID string, patch map[string]any) (*entity.User, error)
	UploadAvatar(ctx context.Context, userID string, file *multipart.FileHeader) (*entity.User, error)
	UpdateMyProfile(ctx context.Context, userID string, displayName, email, phone, username *string) (*entity.User, error)
	ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error
}
