package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/pkg/utils"
)

type TagService struct {
	repo interfaces.TagRepository
}

func NewTagService(repo interfaces.TagRepository) *TagService {
	return &TagService{repo: repo}
}

func (s *TagService) Create(ctx context.Context, userID string, req dto.CreateTagRequest) (*entity.Tag, error) {
	nameVI := utils.NormalizeOptionalString(req.NameVI)
	nameEN := utils.NormalizeOptionalString(req.NameEN)
	if nameVI == nil && nameEN == nil {
		return nil, errors.New("at least one name is required")
	}

	color := utils.NormalizeOptionalString(req.Color)

	now := time.Now().UTC()
	t := entity.Tag{
		ID:        uuid.NewString(),
		UserID:    userID,
		NameVI:    nameVI,
		NameEN:    nameEN,
		Color:     color,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.CreateTag(ctx, userID, t); err != nil {
		return nil, err
	}

	return s.repo.GetTag(ctx, userID, t.ID)
}

func (s *TagService) Get(ctx context.Context, userID, tagID string) (*entity.Tag, error) {
	return s.repo.GetTag(ctx, userID, tagID)
}

func (s *TagService) List(ctx context.Context, userID string) ([]entity.Tag, error) {
	return s.repo.ListTags(ctx, userID)
}

func (s *TagService) GetOrCreateByName(ctx context.Context, userID, name, langHint string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("tag name cannot be empty")
	}

	tags, err := s.repo.ListTags(ctx, userID)
	if err != nil {
		return "", err
	}

	normalizedSearch := strings.ToLower(name)
	for _, t := range tags {
		if t.NameVI != nil && strings.ToLower(*t.NameVI) == normalizedSearch {
			return t.ID, nil
		}
		if t.NameEN != nil && strings.ToLower(*t.NameEN) == normalizedSearch {
			return t.ID, nil
		}
	}

	// Create new tag
	now := time.Now().UTC()
	id := uuid.NewString()
	t := entity.Tag{
		ID:        id,
		UserID:    userID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if langHint == "vi" {
		t.NameVI = &name
	} else {
		t.NameEN = &name
	}

	if err := s.repo.CreateTag(ctx, userID, t); err != nil {
		return "", err
	}

	return id, nil
}
