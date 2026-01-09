package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
)

type User struct {
	ID          string    `json:"id"`
	Email       *string   `json:"email,omitempty"`
	Phone       *string   `json:"phone,omitempty"`
	DisplayName *string   `json:"display_name,omitempty"`
	Settings    any       `json:"settings,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UserWithPassword struct {
	User
	PasswordHash string
}

type UserRepository interface {
	CreateUser(ctx context.Context, user UserWithPassword) error
	FindUserByEmail(ctx context.Context, email string) (*UserWithPassword, error)
	FindUserByPhone(ctx context.Context, phone string) (*UserWithPassword, error)
	FindUserByID(ctx context.Context, id string) (*User, error)
	UpdateUserSettings(ctx context.Context, userID string, patch map[string]any) (*User, error)
}
