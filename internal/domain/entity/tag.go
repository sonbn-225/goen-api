package entity

import (
	"github.com/google/uuid"
)

// Tag represents a label that can be attached to transactions for further categorization.
type Tag struct {
	AuditEntity
	UserID uuid.UUID `json:"user_id"`          // ID of the user who owns this tag
	NameVI *string   `json:"name_vi,omitempty"` // Name of the tag in Vietnamese
	NameEN *string   `json:"name_en,omitempty"` // Name of the tag in English
	Color  *string   `json:"color,omitempty"`   // UI color representation in hex
}

