package entity

import (
	"time"
)

type Contact struct {
	ID                string     `json:"id"`
	UserID            string     `json:"user_id"`
	Name              string     `json:"name"`
	Email             *string    `json:"email,omitempty"`
	Phone             *string    `json:"phone,omitempty"`
	AvatarURL         *string    `json:"avatar_url,omitempty"`
	LinkedUserID      *string    `json:"linked_user_id,omitempty"`
	Notes             *string    `json:"notes,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	DeletedAt         *time.Time `json:"deleted_at,omitempty"`
	LinkedDisplayName *string    `json:"linked_display_name,omitempty"`
	LinkedAvatarURL   *string    `json:"linked_avatar_url,omitempty"`
}
