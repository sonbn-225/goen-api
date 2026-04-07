package entity

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	AuditEntity
	UserID    uuid.UUID `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

