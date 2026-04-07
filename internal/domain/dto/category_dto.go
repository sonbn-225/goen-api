package dto

import (
	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type CategoryResponse struct {
	ID               uuid.UUID  `json:"id"`
	Key              string     `json:"key"`
	ParentKey        *string    `json:"parent_key,omitempty"`
	ParentCategoryID *uuid.UUID `json:"parent_category_id,omitempty"`
	Type             *string    `json:"type,omitempty"`
	SortOrder        *int       `index:"sort_order,omitempty" json:"sort_order,omitempty"`
	IsActive         bool       `json:"is_active"`
	Icon             *string    `json:"icon,omitempty"`
	Color            *string    `json:"color,omitempty"`
}

func NewCategoryResponse(it entity.Category) CategoryResponse {
	return CategoryResponse{
		ID:               it.ID,
		Key:              it.Key,
		ParentKey:        it.ParentKey,
		ParentCategoryID: it.ParentCategoryID,
		Type:             it.Type,
		SortOrder:        it.SortOrder,
		IsActive:         it.IsActive,
		Icon:             it.Icon,
		Color:            it.Color,
	}
}

func NewCategoryResponses(items []entity.Category) []CategoryResponse {
	out := make([]CategoryResponse, len(items))
	for i, it := range items {
		out[i] = NewCategoryResponse(it)
	}
	return out
}
