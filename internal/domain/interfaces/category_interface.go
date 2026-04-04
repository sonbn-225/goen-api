package interfaces

import (
	"context"

	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type CategoryRepository interface {
	GetCategory(ctx context.Context, userID string, categoryID string) (*entity.Category, error)
	ListCategories(ctx context.Context, userID string) ([]entity.Category, error)
}

type CategoryService interface {
	Get(ctx context.Context, userID, categoryID string) (*entity.Category, error)
	List(ctx context.Context, userID string, txType string) ([]entity.Category, error)
}
