package dto

import (
	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// CategoryResponse represents the category information sent back to the client.
// Used in: CategoryHandler, CategoryService, CategoryInterface
// CategoryResponse represents the category information sent back to the client.
// Used in: CategoryHandler, CategoryService, CategoryInterface
type CategoryResponse struct {
	ID               uuid.UUID            `json:"id"`                             // Unique category identifier
	Key              string               `json:"key"`                            // Stable identifier key (e.g., "food")
	ParentKey        *string              `json:"parent_key,omitempty"`           // Key of the parent category
	ParentCategoryID *uuid.UUID           `json:"parent_category_id,omitempty"`    // ID of the parent category
	Type             *entity.CategoryType `json:"type,omitempty"`              // Income or Expense
	SortOrder        *int                 `index:"sort_order,omitempty" json:"sort_order,omitempty"` // Display order in UI
	IsActive         bool                 `json:"is_active"`                      // Whether the category can be used for new transactions
	Icon             *string              `json:"icon,omitempty"`                 // Icon name for UI display
	Color            *string              `json:"color,omitempty"`                // UI color representation in hex
}
