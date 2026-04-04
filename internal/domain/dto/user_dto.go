package dto

import "time"

type UserResponse struct {
	ID             string    `json:"id"`
	Email          *string   `json:"email,omitempty"`
	Phone          *string   `json:"phone,omitempty"`
	DisplayName    *string   `json:"display_name,omitempty"`
	AvatarURL      *string   `json:"avatar_url,omitempty"`
	Username       string    `json:"username"`
	PublicShareURL *string   `json:"public_share_url,omitempty"`
	Settings       any       `json:"settings,omitempty"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
