package contact

import (
	"context"
	"time"
)

type Contact struct {
	ID                string     `json:"id"`
	UserID            string     `json:"user_id"`
	Name              string     `json:"name"`
	Email             *string    `json:"email,omitempty"`
	Phone             *string    `json:"phone,omitempty"`
	AvatarURL         *string    `json:"avatar_url,omitempty"`
	LinkedUserID      *string    `json:"linked_user_id,omitempty"`
	Notes             *string    `json:"notes,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	DeletedAt         *time.Time `json:"deleted_at,omitempty"`
	LinkedDisplayName *string    `json:"linked_display_name,omitempty"`
	LinkedAvatarURL   *string    `json:"linked_avatar_url,omitempty"`
}

type LinkedUser struct {
	ID          string
	DisplayName *string
	AvatarURL   *string
}

type CreateInput struct {
	Name      string  `json:"name"`
	Email     *string `json:"email,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
	Notes     *string `json:"notes,omitempty"`
}

type UpdateInput struct {
	Name      string  `json:"name"`
	Email     *string `json:"email,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
	Notes     *string `json:"notes,omitempty"`
}

type Repository interface {
	Create(ctx context.Context, userID string, input Contact) error
	GetByID(ctx context.Context, userID, contactID string) (*Contact, error)
	ListByUser(ctx context.Context, userID string) ([]Contact, error)
	Update(ctx context.Context, userID string, input Contact) error
	Delete(ctx context.Context, userID, contactID string) error
	FindLinkedUserByEmail(ctx context.Context, email string) (*LinkedUser, error)
	FindLinkedUserByPhone(ctx context.Context, phone string) (*LinkedUser, error)
}

type Service interface {
	Create(ctx context.Context, userID string, input CreateInput) (*Contact, error)
	Get(ctx context.Context, userID, contactID string) (*Contact, error)
	List(ctx context.Context, userID string) ([]Contact, error)
	Update(ctx context.Context, userID, contactID string, input UpdateInput) (*Contact, error)
	Delete(ctx context.Context, userID, contactID string) error
}

type ModuleDeps struct {
	Repo    Repository
	Service Service
}

type Module struct {
	Service Service
	Handler *Handler
}
