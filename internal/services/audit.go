package services

import (
	"context"
	"errors"
	"strings"

	"github.com/sonbn-225/goen-api/internal/domain"
)

type AuditService interface {
	ListAuditEvents(ctx context.Context, userID string, accountID string, limit int) ([]domain.AuditEvent, error)
}

type auditService struct {
	repo domain.AuditRepository
}

func NewAuditService(repo domain.AuditRepository) AuditService {
	return &auditService{repo: repo}
}

func (s *auditService) ListAuditEvents(ctx context.Context, userID string, accountID string, limit int) ([]domain.AuditEvent, error) {
	id := strings.TrimSpace(accountID)
	if id == "" {
		return nil, ValidationError("accountId is required", map[string]any{"field": "accountId"})
	}
	items, err := s.repo.ListAuditEventsForAccount(ctx, userID, id, limit)
	if err != nil {
		if errors.Is(err, domain.ErrAuditForbidden) || errors.Is(err, domain.ErrAccountForbidden) {
			return nil, ForbiddenErrorWithCause("forbidden", nil, err)
		}
		return nil, err
	}
	return items, nil
}
