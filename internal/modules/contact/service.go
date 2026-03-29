package contact

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/apperrors"
	"github.com/sonbn-225/goen-api/internal/domain"
)

type CreateRequest struct {
	Name      string  `json:"name"`
	Email     *string `json:"email,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
	Notes     *string `json:"notes,omitempty"`
}

type UpdateRequest struct {
	Name      string  `json:"name"`
	Email     *string `json:"email,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
	Notes     *string `json:"notes,omitempty"`
}

type Service struct {
	repo domain.ContactRepository
}

func NewService(repo domain.ContactRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, userID string, req CreateRequest) (*domain.Contact, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, apperrors.Validation("name is required", nil)
	}

	now := time.Now().UTC()
	c := domain.Contact{
		ID:        uuid.NewString(),
		UserID:    userID,
		Name:      name,
		Email:     normalizeOptionalString(req.Email),
		Phone:     normalizeOptionalString(req.Phone),
		AvatarURL: normalizeOptionalString(req.AvatarURL),
		Notes:     normalizeOptionalString(req.Notes),
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Try to link with existing Goen user
	s.tryLinkUser(ctx, &c)

	if err := s.repo.CreateContact(ctx, c); err != nil {
		return nil, err
	}

	return s.repo.GetContact(ctx, userID, c.ID)
}

func (s *Service) Get(ctx context.Context, userID, contactID string) (*domain.Contact, error) {
	return s.repo.GetContact(ctx, userID, contactID)
}

func (s *Service) List(ctx context.Context, userID string) ([]domain.Contact, error) {
	return s.repo.ListContacts(ctx, userID)
}

func (s *Service) Update(ctx context.Context, userID, contactID string, req UpdateRequest) (*domain.Contact, error) {
	existing, err := s.repo.GetContact(ctx, userID, contactID)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, apperrors.Validation("name is required", nil)
	}

	existing.Name = name
	existing.Email = normalizeOptionalString(req.Email)
	existing.Phone = normalizeOptionalString(req.Phone)
	existing.AvatarURL = normalizeOptionalString(req.AvatarURL)
	existing.Notes = normalizeOptionalString(req.Notes)
	existing.UpdatedAt = time.Now().UTC()

	// Re-evaluate link
	s.tryLinkUser(ctx, existing)

	if err := s.repo.UpdateContact(ctx, userID, *existing); err != nil {
		return nil, err
	}

	return s.repo.GetContact(ctx, userID, contactID)
}

func (s *Service) Delete(ctx context.Context, userID, contactID string) error {
	return s.repo.DeleteContact(ctx, userID, contactID)
}

func (s *Service) tryLinkUser(ctx context.Context, c *domain.Contact) {
	c.LinkedUserID = nil // Reset first

	// Strategy: Email first, then Phone
	if c.Email != nil && *c.Email != "" {
		u, err := s.repo.FindUserByEmail(ctx, *c.Email)
		if err == nil && u != nil {
			c.LinkedUserID = &u.ID
			return
		}
	}

	if c.Phone != nil && *c.Phone != "" {
		u, err := s.repo.FindUserByPhone(ctx, *c.Phone)
		if err == nil && u != nil {
			c.LinkedUserID = &u.ID
			return
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
