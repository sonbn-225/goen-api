package domain

import (
	"context"
	"time"
)

type Tag struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Color     *string   `json:"color,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TagRepository interface {
	CreateTag(ctx context.Context, userID string, t Tag) error
	GetTag(ctx context.Context, userID string, tagID string) (*Tag, error)
	ListTags(ctx context.Context, userID string) ([]Tag, error)
}

