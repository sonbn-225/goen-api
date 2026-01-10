package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain"
)

type TagService interface {
	Create(ctx context.Context, userID string, req CreateTagRequest) (*domain.Tag, error)
	Get(ctx context.Context, userID string, tagID string) (*domain.Tag, error)
	List(ctx context.Context, userID string) ([]domain.Tag, error)
}

type CreateTagRequest struct {
	Name  string  `json:"name"`
	Color *string `json:"color,omitempty"`
}

type tagService struct {
	repo domain.TagRepository
}

func NewTagService(repo domain.TagRepository) TagService {
	return &tagService{repo: repo}
}

func (s *tagService) Create(ctx context.Context, userID string, req CreateTagRequest) (*domain.Tag, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ValidationError("name is required", map[string]any{"field": "name"})
	}

	color := normalizeOptionalString(req.Color)

	now := time.Now().UTC()
	t := domain.Tag{
		ID:        uuid.NewString(),
		UserID:    userID,
		Name:      name,
		Color:     color,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.CreateTag(ctx, userID, t); err != nil {
		return nil, err
	}

	created, err := s.repo.GetTag(ctx, userID, t.ID)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *tagService) Get(ctx context.Context, userID string, tagID string) (*domain.Tag, error) {
	t, err := s.repo.GetTag(ctx, userID, tagID)
	if err != nil {
		if errors.Is(err, domain.ErrTagNotFound) {
			return nil, NotFoundErrorWithCause("tag not found", nil, err)
		}
		return nil, err
	}
	return t, nil
}

func (s *tagService) List(ctx context.Context, userID string) ([]domain.Tag, error) {
	return s.repo.ListTags(ctx, userID)
}
