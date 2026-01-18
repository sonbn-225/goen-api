package tag

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/apperrors"
	"github.com/sonbn-225/goen-api/internal/domain"
)

// CreateTagRequest contains create tag parameters.
type CreateTagRequest struct {
	Name  string  `json:"name"`
	Color *string `json:"color,omitempty"`
}

// Service handles tag business logic.
type Service struct {
	repo domain.TagRepository
}

// NewService creates a new tag service.
func NewService(repo domain.TagRepository) *Service {
	return &Service{repo: repo}
}

// Create creates a new tag.
func (s *Service) Create(ctx context.Context, userID string, req CreateTagRequest) (*domain.Tag, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, apperrors.Validation("name is required", map[string]any{"field": "name"})
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

// Get retrieves a tag by ID.
func (s *Service) Get(ctx context.Context, userID, tagID string) (*domain.Tag, error) {
	t, err := s.repo.GetTag(ctx, userID, tagID)
	if err != nil {
		if errors.Is(err, apperrors.ErrTagNotFound) {
			return nil, apperrors.Wrap(apperrors.KindNotFound, "tag not found", err)
		}
		return nil, err
	}
	return t, nil
}

// List returns all tags for a user.
func (s *Service) List(ctx context.Context, userID string) ([]domain.Tag, error) {
	return s.repo.ListTags(ctx, userID)
}

// normalizeOptionalString trims whitespace from optional string pointers.
func normalizeOptionalString(s *string) *string {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil
	}
	return &v
}
