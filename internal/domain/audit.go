package domain

import (
	"context"
	"time"
)

type AuditEvent struct {
	ID          string         `json:"id"`
	AccountID   string         `json:"account_id"`
	ActorUserID string         `json:"actor_user_id"`
	Action      string         `json:"action"`
	EntityType  string         `json:"entity_type"`
	EntityID    string         `json:"entity_id"`
	OccurredAt  time.Time      `json:"occurred_at"`
	Diff        map[string]any `json:"diff,omitempty"`
}

type AuditRepository interface {
	ListAuditEventsForAccount(ctx context.Context, userID string, accountID string, limit int) ([]AuditEvent, error)
}

