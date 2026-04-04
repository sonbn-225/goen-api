package dto

import (
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type CategoryResponse struct {
	ID               string  `json:"id"`
	ParentCategoryID *string `json:"parent_category_id,omitempty"`
	Type             *string `json:"type,omitempty"`
	SortOrder        *int    `index:"sort_order,omitempty" json:"sort_order,omitempty"`
	IsActive         bool    `json:"is_active"`
	Icon             *string `json:"icon,omitempty"`
	Color            *string `json:"color,omitempty"`
}

func NewCategoryResponse(it entity.Category) CategoryResponse {
	return CategoryResponse{
		ID:               it.ID,
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
