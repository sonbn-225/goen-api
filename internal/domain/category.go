package domain

import (
	"context"
	"time"
)

type Category struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	ParentCategoryID *string    `json:"parent_category_id,omitempty"`
	Type             *string    `json:"type,omitempty"`
	SortOrder        *int       `json:"sort_order,omitempty"`
	IsActive         bool       `json:"is_active"`
	IsSystem         bool       `json:"is_system"`
	Icon             *string    `json:"icon,omitempty"`
	Color            *string    `json:"color,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"`
}

type CategoryRepository interface {
	GetCategory(ctx context.Context, userID string, categoryID string) (*Category, error)
	ListCategories(ctx context.Context, userID string) ([]Category, error)
}
