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

// List returns all categories for a user, optionally filtered by transaction type.
// If txType is provided (income, expense), only returns categories that support that type.
func (s *Service) List(ctx context.Context, userID string, txType string) ([]domain.Category, error) {
	cats, err := s.repo.ListCategories(ctx, userID)
	if err != nil {
		return nil, err
	}

	// If no type filter specified, return all categories
	if txType == "" {
		return cats, nil
	}

	// Filter by transaction type
	// For income: include categories with type="income" or type="both"
	// For expense: include categories with type="expense" or type="both"
	filtered := make([]domain.Category, 0, len(cats))
	for _, cat := range cats {
		if cat.Type == nil {
			// Include categories with no type restriction
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
