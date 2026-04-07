package dto

import (
	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type CreateContactRequest struct {
	Name      string  `json:"name" binding:"required"`
	Email     *string `json:"email,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
	Notes     *string `json:"notes,omitempty"`
}

type UpdateContactRequest struct {
	Name      *string `json:"name,omitempty"`
	Email     *string `json:"email,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
	Notes     *string `json:"notes,omitempty"`
}

type ContactResponse struct {
	ID                uuid.UUID  `json:"id"`
	UserID                uuid.UUID  `json:"user_id"`
	Name              string     `json:"name"`
	Email             *string    `json:"email,omitempty"`
	Phone             *string    `json:"phone,omitempty"`
	AvatarURL         *string    `json:"avatar_url,omitempty"`
	LinkedUserID      *uuid.UUID `json:"linked_user_id,omitempty"`
	Notes             *string    `json:"notes,omitempty"`
	LinkedDisplayName *string    `json:"linked_display_name,omitempty"`
	LinkedAvatarURL   *string    `json:"linked_avatar_url,omitempty"`
}

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

