package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/pkg/utils"
)

type AuditService struct {
	repo interfaces.AuditRepository
}

func NewAuditService(repo interfaces.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

func (s *AuditService) Record(ctx context.Context, tx pgx.Tx, actorID uuid.UUID, accountID *uuid.UUID, resourceType entity.AuditResourceType, action entity.AuditAction, resourceID uuid.UUID, oldObj any, newObj any) error {
	log := entity.AuditLog{
		ID:           utils.NewID(),
		OccurredAt:   utils.Now(),
		ActorUserID:  actorID,
		AccountID:    accountID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Action:       action,
		Metadata:     make(map[string]any),
	}

	// Calculate diff if both objects are provided (for Updates)
	if oldObj != nil && newObj != nil {
		diff := utils.CalculateDiff(oldObj, newObj)
		if len(diff) > 0 {
			log.Metadata["diff"] = diff
		}
	} else if newObj != nil {
		// For Creates, we might want to store the initial state
		log.Metadata["payload"] = newObj
	} else if oldObj != nil {
		// For Deletes, we might want to store the final state
		log.Metadata["snapshot"] = oldObj
	}

	return s.repo.RecordTx(ctx, tx, log)
}

func (s *AuditService) List(ctx context.Context, userID uuid.UUID, filter entity.AuditLogFilter) ([]entity.AuditLog, error) {
	// For now, we assume users can only see logs they are actors in or which belong to accounts they own.
	// In a real app, we'd add more complex permission checks here.
	return s.repo.ListTx(ctx, nil, filter)
}
