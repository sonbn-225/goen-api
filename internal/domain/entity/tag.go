package entity

import (
	"github.com/google/uuid"
)

type Tag struct {
	AuditEntity
	UserID uuid.UUID `json:"user_id"`
	NameVI *string   `json:"name_vi,omitempty"`
	NameEN *string   `json:"name_en,omitempty"`
	Color  *string   `json:"color,omitempty"`
}

