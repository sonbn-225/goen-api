package entity

import (
	"github.com/google/uuid"
)

type Contact struct {
	AuditEntity
	UserID            uuid.UUID  `json:"user_id"`
	Name              string     `json:"name"`
	Email             *string    `json:"email,omitempty"`
	Phone             *string    `json:"phone,omitempty"`
	AvatarURL         *string    `json:"avatar_url,omitempty"`
	LinkedUserID      *uuid.UUID `json:"linked_user_id,omitempty"`
	Notes             *string    `json:"notes,omitempty"`
	LinkedDisplayName *string    `json:"linked_display_name,omitempty"`
	LinkedAvatarURL   *string    `json:"linked_avatar_url,omitempty"`
}

