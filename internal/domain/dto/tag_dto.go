package dto

import (
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type CreateTagRequest struct {
	NameVI *string `json:"name_vi,omitempty"`
	NameEN *string `json:"name_en,omitempty"`
	Color  *string `json:"color,omitempty"`
}

type TagResponse struct {
	ID     string  `json:"id"`
	UserID string  `json:"user_id"`
	NameVI *string `json:"name_vi,omitempty"`
	NameEN *string `json:"name_en,omitempty"`
	Color  *string `json:"color,omitempty"`
}

func NewTagResponse(it entity.Tag) TagResponse {
	return TagResponse{
		ID:     it.ID,
		UserID: it.UserID,
		NameVI: it.NameVI,
		NameEN: it.NameEN,
		Color:  it.Color,
	}
}

func NewTagResponses(items []entity.Tag) []TagResponse {
	out := make([]TagResponse, len(items))
	for i, it := range items {
		out[i] = NewTagResponse(it)
	}
	return out
}
