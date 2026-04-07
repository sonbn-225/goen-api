package entity

import (
	"time"

	"github.com/google/uuid"
)

// BaseEntity contains the essential ID field for all independent and dependent entities.
type BaseEntity struct {
	ID uuid.UUID `json:"id"`
}

type AuditEntity struct {
	BaseEntity
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
