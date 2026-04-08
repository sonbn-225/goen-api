package entity

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken represents a long-lived credential used to obtain a new access token.
type RefreshToken struct {
	AuditEntity
	UserID    uuid.UUID `json:"user_id"`    // ID of the user this token belongs to
	Token     string    `json:"token"`      // The actual secure token string
	ExpiresAt time.Time `json:"expires_at"` // Timestamp when the token becomes invalid
}

