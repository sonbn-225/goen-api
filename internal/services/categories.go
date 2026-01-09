package services

import (
	"context"
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
	return s.repo.GetCategory(ctx, userID, categoryID)
}

func (s *categoryService) List(ctx context.Context, userID string) ([]domain.Category, error) {
	return s.repo.ListCategories(ctx, userID)
}
