package dto

import (
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

func NewContactResponse(it entity.Contact) ContactResponse {
	return ContactResponse{
		ID:                it.ID,
		UserID:            it.UserID,
		Name:              it.Name,
		Email:             it.Email,
		Phone:             it.Phone,
		AvatarURL:         it.AvatarURL,
		LinkedUserID:      it.LinkedUserID,
		Notes:             it.Notes,
		LinkedDisplayName: it.LinkedDisplayName,
		LinkedAvatarURL:   it.LinkedAvatarURL,
	}
}

func NewContactResponses(items []entity.Contact) []ContactResponse {
	out := make([]ContactResponse, len(items))
	for i, it := range items {
		out[i] = NewContactResponse(it)
	}
	return out
}
