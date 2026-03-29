package domain

import (
	"context"
	"time"
)

// Contact represents a person or entity linked to a user.
type Contact struct {
	ID           string     `json:"id"`
	UserID       string     `json:"user_id"`
	Name         string     `json:"name"`
	Email        *string    `json:"email,omitempty"`
	Phone        *string    `json:"phone,omitempty"`
	AvatarURL    *string    `json:"avatar_url,omitempty"`
	LinkedUserID *string    `json:"linked_user_id,omitempty"`
	Notes        *string    `json:"notes,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`

	// Linked info (synced from Users table)
	LinkedDisplayName *string `json:"linked_display_name,omitempty"`
	LinkedAvatarURL   *string `json:"linked_avatar_url,omitempty"`
}

// ContactRepository defines the storage interface for contacts.
type ContactRepository interface {
	CreateContact(ctx context.Context, c Contact) error
	GetContact(ctx context.Context, userID, contactID string) (*Contact, error)
	ListContacts(ctx context.Context, userID string) ([]Contact, error)
	UpdateContact(ctx context.Context, userID string, c Contact) error
	DeleteContact(ctx context.Context, userID, contactID string) error
	
	// Finding users for linking
	FindUserByEmail(ctx context.Context, email string) (*User, error)
	FindUserByPhone(ctx context.Context, phone string) (*User, error)
}
