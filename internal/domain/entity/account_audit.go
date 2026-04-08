package entity

import (
	"time"

	"github.com/google/uuid"
)

// AccountAuditEvent records a specific change or action taken on an account for auditing purposes.
type AccountAuditEvent struct {
	BaseEntity
	AccountID   uuid.UUID      `json:"account_id"`     // ID of the affected account
	ActorUserID uuid.UUID      `json:"actor_user_id"`  // ID of the user who performed the action
	Action      string         `json:"action"`         // Type of action (e.g., "update", "close")
	EntityType  string         `json:"entity_type"`    // Type of entity modified (always "account")
	EntityID    uuid.UUID      `json:"entity_id"`      // ID of the modified entity
	OccurredAt  time.Time      `json:"occurred_at"`    // Timestamp of the event
	Diff        map[string]any `json:"diff,omitempty"` // Detailed field changes (old vs new)
}
