package contact

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

func (s *service) Create(ctx context.Context, userID string, input CreateInput) (*Contact, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "contact", "operation", "create")
	logger.Info("contact_create_started", "user_id", userID)

	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, apperrors.New(apperrors.KindValidation, "name is required")
	}

	now := time.Now().UTC()
	item := Contact{
		ID:        uuid.NewString(),
		UserID:    userID,
		Name:      name,
		Email:     normalizeOptionalString(input.Email),
		Phone:     normalizeOptionalString(input.Phone),
		AvatarURL: normalizeOptionalString(input.AvatarURL),
		Notes:     normalizeOptionalString(input.Notes),
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.tryLinkUser(ctx, &item)

	if err := s.repo.Create(ctx, userID, item); err != nil {
		logger.Error("contact_create_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to create contact", err)
	}

	created, err := s.repo.GetByID(ctx, userID, item.ID)
	if err != nil {
		logger.Error("contact_create_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to load contact", err)
	}
	if created == nil {
		return nil, apperrors.New(apperrors.KindInternal, "created contact not found")
	}

	logger.Info("contact_create_succeeded", "contact_id", created.ID)
	return created, nil
}

func (s *service) Get(ctx context.Context, userID, contactID string) (*Contact, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "contact", "operation", "get")
	logger.Info("contact_get_started", "user_id", userID, "contact_id", contactID)

	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	if strings.TrimSpace(contactID) == "" {
		return nil, apperrors.New(apperrors.KindValidation, "contactId is required")
	}

	item, err := s.repo.GetByID(ctx, userID, contactID)
	if err != nil {
		logger.Error("contact_get_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to get contact", err)
	}
	if item == nil {
		return nil, apperrors.New(apperrors.KindNotFound, "contact not found")
	}

	logger.Info("contact_get_succeeded", "contact_id", item.ID)
	return item, nil
}

func (s *service) List(ctx context.Context, userID string) ([]Contact, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "contact", "operation", "list")
	logger.Info("contact_list_started", "user_id", userID)

	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}

	items, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		logger.Error("contact_list_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to list contacts", err)
	}

	logger.Info("contact_list_succeeded", "count", len(items))
	return items, nil
}

func (s *service) Update(ctx context.Context, userID, contactID string, input UpdateInput) (*Contact, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "contact", "operation", "update")
	logger.Info("contact_update_started", "user_id", userID, "contact_id", contactID)

	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	if strings.TrimSpace(contactID) == "" {
		return nil, apperrors.New(apperrors.KindValidation, "contactId is required")
	}

	existing, err := s.repo.GetByID(ctx, userID, contactID)
	if err != nil {
		logger.Error("contact_update_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to get contact", err)
	}
	if existing == nil {
		return nil, apperrors.New(apperrors.KindNotFound, "contact not found")
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, apperrors.New(apperrors.KindValidation, "name is required")
	}

	existing.Name = name
	existing.Email = normalizeOptionalString(input.Email)
	existing.Phone = normalizeOptionalString(input.Phone)
	existing.AvatarURL = normalizeOptionalString(input.AvatarURL)
	existing.Notes = normalizeOptionalString(input.Notes)
	existing.UpdatedAt = time.Now().UTC()

	s.tryLinkUser(ctx, existing)

	if err := s.repo.Update(ctx, userID, *existing); err != nil {
		logger.Error("contact_update_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to update contact", err)
	}

	updated, err := s.repo.GetByID(ctx, userID, contactID)
	if err != nil {
		logger.Error("contact_update_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to load updated contact", err)
	}
	if updated == nil {
		return nil, apperrors.New(apperrors.KindNotFound, "contact not found")
	}

	logger.Info("contact_update_succeeded", "contact_id", updated.ID)
	return updated, nil
}

func (s *service) Delete(ctx context.Context, userID, contactID string) error {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "contact", "operation", "delete")
	logger.Info("contact_delete_started", "user_id", userID, "contact_id", contactID)

	if strings.TrimSpace(userID) == "" {
		return apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	if strings.TrimSpace(contactID) == "" {
		return apperrors.New(apperrors.KindValidation, "contactId is required")
	}

	existing, err := s.repo.GetByID(ctx, userID, contactID)
	if err != nil {
		logger.Error("contact_delete_failed", "error", err)
		return apperrors.Wrap(apperrors.KindInternal, "failed to get contact", err)
	}
	if existing == nil {
		return apperrors.New(apperrors.KindNotFound, "contact not found")
	}

	if err := s.repo.Delete(ctx, userID, contactID); err != nil {
		logger.Error("contact_delete_failed", "error", err)
		return apperrors.Wrap(apperrors.KindInternal, "failed to delete contact", err)
	}

	logger.Info("contact_delete_succeeded", "contact_id", contactID)
	return nil
}

func (s *service) tryLinkUser(ctx context.Context, c *Contact) {
	c.LinkedUserID = nil
	c.LinkedDisplayName = nil
	c.LinkedAvatarURL = nil

	if c.Email != nil && strings.TrimSpace(*c.Email) != "" {
		u, err := s.repo.FindLinkedUserByEmail(ctx, *c.Email)
		if err == nil && u != nil {
			c.LinkedUserID = &u.ID
			c.LinkedDisplayName = u.DisplayName
			c.LinkedAvatarURL = u.AvatarURL
			return
		}
	}

	if c.Phone != nil && strings.TrimSpace(*c.Phone) != "" {
		u, err := s.repo.FindLinkedUserByPhone(ctx, *c.Phone)
		if err == nil && u != nil {
			c.LinkedUserID = &u.ID
			c.LinkedDisplayName = u.DisplayName
			c.LinkedAvatarURL = u.AvatarURL
		}
	}
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
