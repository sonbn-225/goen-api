package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type AuditLogResponse struct {
	ID           uuid.UUID                `json:"id"`
	OccurredAt   time.Time                `json:"occurred_at"`
	ActorUserID  uuid.UUID                `json:"actor_user_id"`
	AccountID    *uuid.UUID               `json:"account_id,omitempty"`
	ResourceType entity.AuditResourceType `json:"resource_type"`
	ResourceID   uuid.UUID                `json:"resource_id"`
	Action       entity.AuditAction       `json:"action"`
	Metadata     map[string]any           `json:"metadata"`
}

type AuditLogFilterRequest struct {
	ActorUserID  *uuid.UUID                `json:"actor_user_id,omitempty"`
	AccountID    *uuid.UUID                `json:"account_id,omitempty"`
	ResourceType *entity.AuditResourceType `json:"resource_type,omitempty"`
	ResourceID   *uuid.UUID                `json:"resource_id,omitempty"`
	Action       *entity.AuditAction       `json:"action,omitempty"`
	From         *time.Time                `json:"from,omitempty"`
	To           *time.Time                `json:"to,omitempty"`
	Limit        int                       `json:"limit"`
	Offset       int                       `json:"offset"`
}
