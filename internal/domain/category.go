package domain

import (
	"context"
	"time"
)

type Category struct {
	ID               string     `json:"id"`
	ParentCategoryID *string    `json:"parent_category_id,omitempty"`
	Type             *string    `json:"type,omitempty"`
	SortOrder        *int       `json:"sort_order,omitempty"`
	IsActive         bool       `json:"is_active"`
	Icon             *string    `json:"icon,omitempty"`
	Color            *string    `json:"color,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"`
}

type CategoryRepository interface {
	GetCategory(ctx context.Context, userID string, categoryID string) (*Category, error)
	ListCategories(ctx context.Context, userID string) ([]Category, error)
	FindCategoryByName(ctx context.Context, userID string, name string) (*Category, error)
}
