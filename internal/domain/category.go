package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrCategoryNotFound = errors.New("category not found")
)

type Category struct {
	ID              string    `json:"id"`
	UserID          *string   `json:"user_id,omitempty"`
	Name            string    `json:"name"`
	ParentCategoryID *string   `json:"parent_category_id,omitempty"`
	Type            *string   `json:"type,omitempty"`
	SortOrder       *int      `json:"sort_order,omitempty"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
}

type CategoryRepository interface {
	CreateCategory(ctx context.Context, userID string, c Category) error
	GetCategory(ctx context.Context, userID string, categoryID string) (*Category, error)
	ListCategories(ctx context.Context, userID string) ([]Category, error)
}
