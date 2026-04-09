package entity

import (
	"time"

	"github.com/google/uuid"
)


// AuditLog represents a unified record of an action taken within the system.
type AuditLog struct {
	ID           uuid.UUID         `json:"id"`
	OccurredAt   time.Time         `json:"occurred_at"`
	ActorUserID  uuid.UUID         `json:"actor_user_id"`
	AccountID    *uuid.UUID        `json:"account_id,omitempty"`
	ResourceType AuditResourceType `json:"resource_type"`
	ResourceID   uuid.UUID         `json:"resource_id"`
	Action       AuditAction       `json:"action"`
	Metadata     map[string]any    `json:"metadata"` // Can store diffs, descriptions, IPs, etc.
}

// FieldDiff represents a change to a single field.
type FieldDiff struct {
	Old any `json:"old"`
	New any `json:"new"`
}

// DiffMap is a collection of field changes.
type DiffMap map[string]FieldDiff

// AuditLogFilter defines criteria for querying audit logs.
type AuditLogFilter struct {
	ActorUserID  *uuid.UUID         `json:"actor_user_id,omitempty"`
	AccountID    *uuid.UUID         `json:"account_id,omitempty"`
	ResourceType *AuditResourceType `json:"resource_type,omitempty"`
	ResourceID   *uuid.UUID         `json:"resource_id,omitempty"`
	Action       *AuditAction       `json:"action,omitempty"`
	From         *time.Time         `json:"from,omitempty"`
	To           *time.Time         `json:"to,omitempty"`
	Limit        int                `json:"limit"`
	Offset       int                `json:"offset"`
}
