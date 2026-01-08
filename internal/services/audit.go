package services

import (
	"context"

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
	return s.repo.ListAuditEventsForAccount(ctx, userID, accountID, limit)
}
