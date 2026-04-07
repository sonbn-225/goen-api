package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
)

const (
	CategoryCacheKey = "goen:categories:all5"
	CategoryCacheTTL = 24 * time.Hour
)

type CategoryService struct {
	repo interfaces.CategoryRepository
	rds  *database.Redis
}

func NewCategoryService(repo interfaces.CategoryRepository, rds *database.Redis) *CategoryService {
	return &CategoryService{
		repo: repo,
		rds:  rds,
	}
}

func (s *CategoryService) Get(ctx context.Context, userID, categoryID uuid.UUID) (*dto.CategoryResponse, error) {
	// For Get, we can just fetch all from cache and find the specific one
	// as the list is small and global.
	cats, err := s.List(ctx, userID, "")
	if err != nil {
		return nil, err
	}

	for _, c := range cats {
		if c.ID == categoryID {
			return &c, nil
		}
	}

	return nil, nil
}

func (s *CategoryService) List(ctx context.Context, userID uuid.UUID, txType string) ([]dto.CategoryResponse, error) {
	var allCats []dto.CategoryResponse

	// Try cache first
	if s.rds != nil {
		cached, err := s.rds.Get(ctx, CategoryCacheKey)
		if err == nil {
			if err := json.Unmarshal([]byte(cached), &allCats); err == nil {
				return s.filterByType(allCats, txType), nil
			}
		}
	}

	// Cache miss or Redis not configured
	cats, err := s.repo.ListCategories(ctx, userID)
	if err != nil {
		return nil, err
	}

	allCats = dto.NewCategoryResponses(cats)

	// Save to cache (best effort)
	if s.rds != nil {
		if data, err := json.Marshal(allCats); err == nil {
			_ = s.rds.Set(ctx, CategoryCacheKey, string(data), CategoryCacheTTL)
		}
	}

	return s.filterByType(allCats, txType), nil
}

func (s *CategoryService) filterByType(cats []dto.CategoryResponse, txType string) []dto.CategoryResponse {
	if txType == "" {
		return cats
	}

	filtered := make([]dto.CategoryResponse, 0, len(cats))
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
	return filtered
}
