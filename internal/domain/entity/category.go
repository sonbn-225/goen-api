package entity

import (
	"github.com/google/uuid"
)

// Category represents a classification for transactions (e.g. Food, Salary).
type Category struct {
	BaseEntity
	Key              string        `json:"key"`                           // Unique identifier for the category (e.g., "food", "salary")
	ParentKey        *string       `json:"parent_key,omitempty"`          // Key of the parent category for sub-categories
	ParentCategoryID *uuid.UUID    `json:"parent_category_id,omitempty"`   // ID of the parent category
	Type             *CategoryType `json:"type,omitempty"`             // Income or Expense
	SortOrder        *int          `json:"sort_order,omitempty"`          // Order for UI display
	IsActive         bool          `json:"is_active"`                     // Whether the category is currently available for use
	Icon             *string       `json:"icon,omitempty"`                // Icon name or identifier for the UI
	Color            *string       `json:"color,omitempty"`               // UI color representation in hex
}
