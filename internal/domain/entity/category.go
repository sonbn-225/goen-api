package entity

import (
	"github.com/google/uuid"
)

type Category struct {
	AuditEntity
	Key              string     `json:"key"`
	ParentCategoryID *uuid.UUID `json:"parent_category_id,omitempty"`
	Type             *string    `json:"type,omitempty"`
	SortOrder        *int       `json:"sort_order,omitempty"`
	IsActive         bool       `json:"is_active"`
	Icon             *string    `json:"icon,omitempty"`
	Color            *string    `json:"color,omitempty"`
}
