package interfaces

import (
	"context"
	"mime/multipart"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user entity.UserWithPassword) error
	FindUserByEmail(ctx context.Context, email string) (*entity.UserWithPassword, error)
	FindUserByPhone(ctx context.Context, phone string) (*entity.UserWithPassword, error)
	FindUserByUsername(ctx context.Context, username string) (*entity.UserWithPassword, error)
	FindUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	UpdateUserSettings(ctx context.Context, userID uuid.UUID, patch map[string]any) (*entity.User, error)
	UpdateUserProfile(ctx context.Context, userID uuid.UUID, params entity.UpdateUserParams) (*entity.User, error)
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *entity.RefreshToken) error
	GetByToken(ctx context.Context, token string) (*entity.RefreshToken, error)
	DeleteByToken(ctx context.Context, token string) error
	DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error
}

type AuthService interface {
	Signup(ctx context.Context, req dto.SignupRequest) (*dto.AuthResponse, error)
	Signin(ctx context.Context, req dto.SigninRequest) (*dto.AuthResponse, error)
	Refresh(ctx context.Context, refreshToken string) (*dto.AuthResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	GetMe(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error)
	UpdateMySettings(ctx context.Context, userID uuid.UUID, patch map[string]any) (*dto.UserResponse, error)
	UploadAvatar(ctx context.Context, userID uuid.UUID, file *multipart.FileHeader) (*dto.UserResponse, error)
	GetMyAvatars(ctx context.Context, userID uuid.UUID) ([]dto.MediaResponse, error)
	UpdateMyProfile(ctx context.Context, userID uuid.UUID, displayName, email, phone, username *string) (*dto.UserResponse, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error
}

