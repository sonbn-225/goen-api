package service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/pkg/utils"
)

type ContactService struct {
	repo interfaces.ContactRepository
}

func NewContactService(repo interfaces.ContactRepository) *ContactService {
	return &ContactService{repo: repo}
}

func (s *ContactService) Create(ctx context.Context, userID uuid.UUID, req dto.CreateContactRequest) (*dto.ContactResponse, error) {
	c := entity.Contact{
		AuditEntity: entity.AuditEntity{
			BaseEntity: entity.BaseEntity{
				ID: utils.NewID(),
			},
		},
		UserID:    userID,
		Name:      strings.TrimSpace(req.Name),
		Email:     req.Email,
		Phone:     req.Phone,
		AvatarURL: req.AvatarURL,
		Notes:     req.Notes,
	}

	// Try auto-linking by email or phone
	if req.Email != nil && strings.TrimSpace(*req.Email) != "" {
		u, _ := s.repo.FindUserByEmail(ctx, strings.TrimSpace(*req.Email))
		if u != nil {
			c.LinkedUserID = &u.ID
		}
	} else if req.Phone != nil && strings.TrimSpace(*req.Phone) != "" {
		u, _ := s.repo.FindUserByPhone(ctx, strings.TrimSpace(*req.Phone))
		if u != nil {
			c.LinkedUserID = &u.ID
		}
	}

	if err := s.repo.CreateContact(ctx, c); err != nil {
		return nil, err
	}
	return s.Get(ctx, userID, c.ID)
}

func (s *ContactService) Get(ctx context.Context, userID, contactID uuid.UUID) (*dto.ContactResponse, error) {
	it, err := s.repo.GetContact(ctx, userID, contactID)
	if err != nil {
		return nil, err
	}
	if it == nil {
		return nil, nil
	}
	resp := dto.NewContactResponse(*it)
	return &resp, nil
}

func (s *ContactService) List(ctx context.Context, userID uuid.UUID) ([]dto.ContactResponse, error) {
	items, err := s.repo.ListContacts(ctx, userID)
	if err != nil {
		return nil, err
	}
	return dto.NewContactResponses(items), nil
}

func (s *ContactService) Update(ctx context.Context, userID, contactID uuid.UUID, req dto.UpdateContactRequest) (*dto.ContactResponse, error) {
	cur, err := s.repo.GetContact(ctx, userID, contactID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		cur.Name = strings.TrimSpace(*req.Name)
	}
	if req.Email != nil {
		cur.Email = req.Email
	}
	if req.Phone != nil {
		cur.Phone = req.Phone
	}
	if req.AvatarURL != nil {
		cur.AvatarURL = req.AvatarURL
	}
	if req.Notes != nil {
		cur.Notes = req.Notes
	}

	// Re-check linking if email/phone changed
	if req.Email != nil || req.Phone != nil {
		if cur.Email != nil && strings.TrimSpace(*cur.Email) != "" {
			u, _ := s.repo.FindUserByEmail(ctx, strings.TrimSpace(*cur.Email))
			if u != nil {
				cur.LinkedUserID = &u.ID
			} else {
				cur.LinkedUserID = nil
			}
		} else if cur.Phone != nil && strings.TrimSpace(*cur.Phone) != "" {
			u, _ := s.repo.FindUserByPhone(ctx, strings.TrimSpace(*cur.Phone))
			if u != nil {
				cur.LinkedUserID = &u.ID
			} else {
				cur.LinkedUserID = nil
			}
		} else {
			cur.LinkedUserID = nil
		}
	}

	if err := s.repo.UpdateContact(ctx, userID, *cur); err != nil {
		return nil, err
	}
	return s.Get(ctx, userID, contactID)
}

func (s *ContactService) Delete(ctx context.Context, userID, contactID uuid.UUID) error {
	return s.repo.DeleteContact(ctx, userID, contactID)
}

func (s *ContactService) GetOrCreateByName(ctx context.Context, userID uuid.UUID, name string) (uuid.UUID, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return uuid.Nil, nil
	}

	contacts, err := s.repo.ListContacts(ctx, userID)
	if err == nil {
		for _, c := range contacts {
			if strings.EqualFold(c.Name, name) {
				return c.ID, nil
			}
		}
	}

	// Not found, create new
	c, err := s.Create(ctx, userID, dto.CreateContactRequest{Name: name})
	if err != nil {
		return uuid.Nil, err
	}
	return c.ID, nil
}

