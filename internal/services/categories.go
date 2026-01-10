package services

import (
	"context"
	"errors"
	"strings"

	"github.com/sonbn-225/goen-api/internal/domain"
)

type CategoryService interface {
	Get(ctx context.Context, userID string, categoryID string) (*domain.Category, error)
	List(ctx context.Context, userID string) ([]domain.Category, error)
}

type categoryService struct {
	repo domain.CategoryRepository
}

func NewCategoryService(repo domain.CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) Get(ctx context.Context, userID string, categoryID string) (*domain.Category, error) {
	id := strings.TrimSpace(categoryID)
	if id == "" {
		return nil, ValidationError("categoryId is required", map[string]any{"field": "categoryId"})
	}
	item, err := s.repo.GetCategory(ctx, userID, id)
	if err != nil {
		if errors.Is(err, domain.ErrCategoryNotFound) {
			return nil, NotFoundErrorWithCause("category not found", nil, err)
		}
		return nil, err
	}
	return item, nil
}

func (s *categoryService) List(ctx context.Context, userID string) ([]domain.Category, error) {
	return s.repo.ListCategories(ctx, userID)
}
