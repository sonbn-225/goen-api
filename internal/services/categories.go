package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain"
)

type CategoryService interface {
	Create(ctx context.Context, userID string, req CreateCategoryRequest) (*domain.Category, error)
	Get(ctx context.Context, userID string, categoryID string) (*domain.Category, error)
	List(ctx context.Context, userID string) ([]domain.Category, error)
}

type CreateCategoryRequest struct {
	Name            string  `json:"name"`
	ParentCategoryID *string `json:"parent_category_id,omitempty"`
	Type            *string `json:"type,omitempty"`
	SortOrder       *int    `json:"sort_order,omitempty"`
	IsActive        *bool   `json:"is_active,omitempty"`
}

type categoryService struct {
	repo domain.CategoryRepository
}

func NewCategoryService(repo domain.CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) Create(ctx context.Context, userID string, req CreateCategoryRequest) (*domain.Category, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("name is required")
	}

	var kind *string
	if req.Type != nil {
		v := strings.TrimSpace(*req.Type)
		if v != "" {
			if v != "expense" && v != "income" && v != "both" {
				return nil, errors.New("type is invalid")
			}
			kind = &v
		}
	}

	parentID := normalizeOptionalString(req.ParentCategoryID)
	if parentID != nil {
		// Ensure parent belongs to user
		if _, err := s.repo.GetCategory(ctx, userID, *parentID); err != nil {
			return nil, err
		}
	}

	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	}

	now := time.Now().UTC()
	uid := userID
	c := domain.Category{
		ID:               uuid.NewString(),
		UserID:           &uid,
		Name:             name,
		ParentCategoryID: parentID,
		Type:             kind,
		SortOrder:        req.SortOrder,
		IsActive:         active,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.repo.CreateCategory(ctx, userID, c); err != nil {
		return nil, err
	}

	created, err := s.repo.GetCategory(ctx, userID, c.ID)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *categoryService) Get(ctx context.Context, userID string, categoryID string) (*domain.Category, error) {
	return s.repo.GetCategory(ctx, userID, categoryID)
}

func (s *categoryService) List(ctx context.Context, userID string) ([]domain.Category, error) {
	items, err := s.repo.ListCategories(ctx, userID)
	if err != nil {
		return nil, err
	}

	type key struct {
		name   string
		parent string
		typev  string
	}

	normalizeType := func(v *string) string {
		if v == nil {
			return ""
		}
		return strings.TrimSpace(*v)
	}

	normalizeParent := func(v *string) string {
		if v == nil {
			return ""
		}
		return strings.TrimSpace(*v)
	}

	isUserCat := func(c domain.Category) bool {
		if c.UserID == nil {
			return false
		}
		return strings.TrimSpace(*c.UserID) == strings.TrimSpace(userID)
	}

	indexByKey := make(map[key]int)
	out := make([]domain.Category, 0, len(items))
	for _, c := range items {
		k := key{
			name:   strings.ToLower(strings.TrimSpace(c.Name)),
			parent: normalizeParent(c.ParentCategoryID),
			typev:  normalizeType(c.Type),
		}
		if idx, ok := indexByKey[k]; ok {
			// Prefer user category over global category for the same logical key.
			if isUserCat(c) && !isUserCat(out[idx]) {
				out[idx] = c
			}
			continue
		}
		indexByKey[k] = len(out)
		out = append(out, c)
	}

	return out, nil
}
