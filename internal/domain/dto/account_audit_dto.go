package dto

import (
	"time"

	"github.com/google/uuid"
)

// AccountAuditEventResponse represents an audit event related to an account.
// Used in: AccountHandler, AccountService.ListAuditEvents, AccountInterface
type AccountAuditEventResponse struct {
	ID          uuid.UUID      `json:"id"`            // Unique identifier for the audit event
	AccountID   uuid.UUID      `json:"account_id"`     // ID of the account the event relates to
	ActorUserID uuid.UUID      `json:"actor_user_id"`  // ID of the user who performed the action
	Action      string         `json:"action"`         // Type of action (e.g., "create", "update")
	EntityType  string         `json:"entity_type"`    // Type of entity (always "account")
	EntityID    uuid.UUID      `json:"entity_id"`      // ID of the modified entity record
	OccurredAt  time.Time      `json:"occurred_at"`    // Timestamp when the event occurred
	Diff        map[string]any `json:"diff,omitempty"` // Detailed field changes (old vs new)
}
