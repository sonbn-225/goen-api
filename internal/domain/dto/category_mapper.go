package dto

import (
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

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
