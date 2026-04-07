package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type CategoryRepository interface {
	GetCategory(ctx context.Context, userID uuid.UUID, categoryID uuid.UUID) (*entity.Category, error)
	ListCategories(ctx context.Context, userID uuid.UUID) ([]entity.Category, error)
}

type CategoryService interface {
	Get(ctx context.Context, userID, categoryID uuid.UUID) (*dto.CategoryResponse, error)
	List(ctx context.Context, userID uuid.UUID, txType string) ([]dto.CategoryResponse, error)
}

