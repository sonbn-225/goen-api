package service

import (
	"context"
	"errors"
	"strings"

	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
)

type CategoryService struct {
	repo interfaces.CategoryRepository
}

func NewCategoryService(repo interfaces.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) Get(ctx context.Context, userID, categoryID string) (*entity.Category, error) {
	id := strings.TrimSpace(categoryID)
	if id == "" {
		return nil, errors.New("category ID is required")
	}

	return s.repo.GetCategory(ctx, userID, id)
}

func (s *CategoryService) List(ctx context.Context, userID string, txType string) ([]entity.Category, error) {
	cats, err := s.repo.ListCategories(ctx, userID)
	if err != nil {
		return nil, err
	}

	if txType == "" {
		return cats, nil
	}

	filtered := make([]entity.Category, 0, len(cats))
	for _, cat := range cats {
		if cat.Type == nil {
			filtered = append(filtered, cat)
			continue
		}
		catType := *cat.Type
		if txType == "income" && (catType == "income" || catType == "both") {
			filtered = append(filtered, cat)
		} else if txType == "expense" && (catType == "expense" || catType == "both") {
			filtered = append(filtered, cat)
		}
	}
	return filtered, nil
}
