package category

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

type Repository interface {
	GetByID(ctx context.Context, userID, categoryID string) (*Category, error)
	ListByUser(ctx context.Context, userID string) ([]Category, error)
}

type Service interface {
	Get(ctx context.Context, userID, categoryID string) (*Category, error)
	List(ctx context.Context, userID string, txType string) ([]Category, error)
}

type ModuleDeps struct {
	Repo    Repository
	Service Service
}

type Module struct {
	Service Service
	Handler *Handler
}
