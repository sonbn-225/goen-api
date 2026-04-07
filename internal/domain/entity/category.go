package entity

import (
	"github.com/google/uuid"
)

type CategoryType string

const (
	CategoryTypeExpense CategoryType = "expense"
	CategoryTypeIncome  CategoryType = "income"
)

type Category struct {
	AuditEntity
	Key              string        `json:"key"`
	ParentKey        *string       `json:"parent_key,omitempty"`
	ParentCategoryID *uuid.UUID    `json:"parent_category_id,omitempty"`
	Type             *CategoryType `json:"type,omitempty"`
	SortOrder        *int          `json:"sort_order,omitempty"`
	IsActive         bool          `json:"is_active"`
	Icon             *string    `json:"icon,omitempty"`
	Color            *string    `json:"color,omitempty"`
}
