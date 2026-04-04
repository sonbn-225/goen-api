package tag

import (
	"context"
	"time"
)

type Tag struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	NameVI    *string   `json:"name_vi,omitempty"`
	NameEN    *string   `json:"name_en,omitempty"`
	Color     *string   `json:"color,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateInput struct {
	NameVI *string `json:"name_vi,omitempty"`
	NameEN *string `json:"name_en,omitempty"`
	Color  *string `json:"color,omitempty"`
}

type Repository interface {
	Create(ctx context.Context, userID string, input Tag) error
	GetByID(ctx context.Context, userID, tagID string) (*Tag, error)
	ListByUser(ctx context.Context, userID string) ([]Tag, error)
}

type Service interface {
	Create(ctx context.Context, userID string, input CreateInput) (*Tag, error)
	Get(ctx context.Context, userID, tagID string) (*Tag, error)
	List(ctx context.Context, userID string) ([]Tag, error)
	GetOrCreateByName(ctx context.Context, userID, name, langHint string) (string, error)
}

type ModuleDeps struct {
	Repo    Repository
	Service Service
}

type Module struct {
	Service Service
	Handler *Handler
}
