package auth

import (
	"context"
	"time"
)

type User struct {
	ID          string         `json:"id"`
	Username    string         `json:"username"`
	Email       *string        `json:"email,omitempty"`
	Phone       *string        `json:"phone,omitempty"`
	DisplayName *string        `json:"display_name,omitempty"`
	AvatarURL   *string        `json:"avatar_url,omitempty"`
	Settings    map[string]any `json:"settings,omitempty"`
	CreatedAt   time.Time      `json:"created_at,omitempty"`
	UpdatedAt   time.Time      `json:"updated_at,omitempty"`
}

type UserWithPassword struct {
	User
	PasswordHash string
}

type SignupRequest struct {
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	DisplayName string `json:"display_name"`
	Username    string `json:"username"`
	Password    string `json:"password"`
}

type SigninRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	User        User   `json:"user"`
}

type UpdateProfileInput struct {
	DisplayName *string
	Email       *string
	Phone       *string
	Username    *string
}

type UserRepository interface {
	CreateUser(ctx context.Context, user UserWithPassword) error
	FindUserByID(ctx context.Context, userID string) (*User, error)
	FindUserByEmail(ctx context.Context, email string) (*UserWithPassword, error)
	FindUserByPhone(ctx context.Context, phone string) (*UserWithPassword, error)
	FindUserByUsername(ctx context.Context, username string) (*UserWithPassword, error)
	UpdateUserProfile(ctx context.Context, userID string, input UpdateProfileInput) (*User, error)
	UpdateAvatarURL(ctx context.Context, userID, avatarURL string) (*User, error)
	UpdateUserSettings(ctx context.Context, userID string, patch map[string]any) (*User, error)
	UpdatePasswordHash(ctx context.Context, userID, passwordHash string) error
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

type TokenIssuer interface {
	Issue(userID string) (string, error)
}

type Service interface {
	Signup(ctx context.Context, req SignupRequest) (*AuthResponse, error)
	Signin(ctx context.Context, req SigninRequest) (*AuthResponse, error)
	Refresh(ctx context.Context, userID string) (*AuthResponse, error)
	GetMe(ctx context.Context, userID string) (*User, error)
	UpdateMyProfile(ctx context.Context, userID string, input UpdateProfileInput) (*User, error)
	UpdateMySettings(ctx context.Context, userID string, patch map[string]any) (*User, error)
	UpdateMyAvatar(ctx context.Context, userID, avatarURL string) (*User, error)
	ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error
}

type ModuleDeps struct {
	UserRepo         UserRepository
	Hasher           PasswordHasher
	Issuer           TokenIssuer
	AccessTTLMinutes int
	Service          Service
}

type Module struct {
	Service Service
	Handler *Handler
}
