package entity

import (
	"time"

	"github.com/google/uuid"
)

// AccountShare tracks which users have access to an account and their permission level.
type AccountShare struct {
	AuditEntity
	AccountID       uuid.UUID              `json:"account_id"`                  // ID of the account being shared
	UserID          uuid.UUID              `json:"user_id"`                     // ID of the user receiving access
	Permission      AccountSharePermission `json:"permission"`                 // Access level (viewer/editor/owner)
	Status          AccountShareStatus     `json:"status"`                     // Current sharing status (active/revoked)
	RevokedAt       *time.Time             `json:"revoked_at,omitempty"`         // Timestamp when access was revoked
	UserEmail       *string                `json:"user_email,omitempty"`         // Email of the shared-to user (enriched)
	UserPhone       *string                `json:"user_phone,omitempty"`         // Phone of the shared-to user (enriched)
	UserDisplayName *string                `json:"user_display_name,omitempty"` // Display name of the shared-to user (enriched)
}
