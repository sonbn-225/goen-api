package category

import (
	"context"
	"errors"
	"strings"

	"github.com/sonbn-225/goen-api/internal/apperrors"
	"github.com/sonbn-225/goen-api/internal/domain"
)

// Service handles category business logic.
type Service struct {
	repo domain.CategoryRepository
}

// NewService creates a new category service.
func NewService(repo domain.CategoryRepository) *Service {
	return &Service{repo: repo}
}

// Get retrieves a single category.
func (s *Service) Get(ctx context.Context, userID, categoryID string) (*domain.Category, error) {
	id := strings.TrimSpace(categoryID)
	if id == "" {
		return nil, apperrors.Validation("categoryId is required", map[string]any{"field": "categoryId"})
	}

	item, err := s.repo.GetCategory(ctx, userID, id)
	if err != nil {
		if errors.Is(err, apperrors.ErrCategoryNotFound) {
			return nil, apperrors.Wrap(apperrors.KindNotFound, "category not found", err)
		}
		return nil, err
	}
	return item, nil
}

// List returns all categories for a user.
func (s *Service) List(ctx context.Context, userID string) ([]domain.Category, error) {
	return s.repo.ListCategories(ctx, userID)
}

