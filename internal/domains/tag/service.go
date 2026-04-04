package tag

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/core/logx"
)

type service struct {
	repo Repository
}

var _ Service = (*service)(nil)

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, userID string, input CreateInput) (*Tag, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "tag", "operation", "create")
	logger.Info("tag_create_started", "user_id", userID)

	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}

	nameVI := normalizeOptionalString(input.NameVI)
	nameEN := normalizeOptionalString(input.NameEN)
	if nameVI == nil && nameEN == nil {
		return nil, apperrors.New(apperrors.KindValidation, "at least one name is required")
	}

	now := time.Now().UTC()
	tag := Tag{
		ID:        uuid.NewString(),
		UserID:    userID,
		NameVI:    nameVI,
		NameEN:    nameEN,
		Color:     normalizeOptionalString(input.Color),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, userID, tag); err != nil {
		logger.Error("tag_create_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to create tag", err)
	}

	created, err := s.repo.GetByID(ctx, userID, tag.ID)
	if err != nil {
		logger.Error("tag_create_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to load tag", err)
	}
	if created == nil {
		return nil, apperrors.New(apperrors.KindInternal, "created tag not found")
	}

	logger.Info("tag_create_succeeded", "tag_id", created.ID)
	return created, nil
}

func (s *service) Get(ctx context.Context, userID, tagID string) (*Tag, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "tag", "operation", "get")
	logger.Info("tag_get_started", "user_id", userID, "tag_id", tagID)

	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	if strings.TrimSpace(tagID) == "" {
		return nil, apperrors.New(apperrors.KindValidation, "tagId is required")
	}

	t, err := s.repo.GetByID(ctx, userID, tagID)
	if err != nil {
		logger.Error("tag_get_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to get tag", err)
	}
	if t == nil {
		return nil, apperrors.New(apperrors.KindNotFound, "tag not found")
	}

	logger.Info("tag_get_succeeded", "tag_id", t.ID)
	return t, nil
}

func (s *service) List(ctx context.Context, userID string) ([]Tag, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "tag", "operation", "list")
	logger.Info("tag_list_started", "user_id", userID)

	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}

	items, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		logger.Error("tag_list_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to list tags", err)
	}

	logger.Info("tag_list_succeeded", "count", len(items))
	return items, nil
}

func (s *service) GetOrCreateByName(ctx context.Context, userID, name, langHint string) (string, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "tag", "operation", "get_or_create_by_name")
	name = strings.TrimSpace(name)
	if strings.TrimSpace(userID) == "" {
		return "", apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	if name == "" {
		return "", apperrors.New(apperrors.KindValidation, "tag name cannot be empty")
	}

	items, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		logger.Error("tag_get_or_create_failed", "error", err)
		return "", apperrors.Wrap(apperrors.KindInternal, "failed to list tags", err)
	}

	search := strings.ToLower(name)
	for _, item := range items {
		if item.NameVI != nil && strings.ToLower(*item.NameVI) == search {
			return item.ID, nil
		}
		if item.NameEN != nil && strings.ToLower(*item.NameEN) == search {
			return item.ID, nil
		}
	}

	in := CreateInput{}
	if strings.ToLower(strings.TrimSpace(langHint)) == "vi" {
		in.NameVI = &name
	} else {
		in.NameEN = &name
	}

	created, err := s.Create(ctx, userID, in)
	if err != nil {
		return "", err
	}
	return created.ID, nil
}

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
