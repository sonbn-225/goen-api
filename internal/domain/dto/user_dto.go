package dto

import "github.com/google/uuid"

// UserResponse represents the public and private profile information of a user.
// Used in: UserHandler, UserService, UserInterface, AuthResponse
// UserResponse represents the public and private profile information of a user.
// Used in: UserHandler, UserService, UserInterface, AuthResponse
type UserResponse struct {
	ID             uuid.UUID `json:"id"`                         // Unique user identity
	Email          *string   `json:"email,omitempty"`             // User's email address
	Phone          *string   `json:"phone,omitempty"`             // User's phone number
	DisplayName    *string   `json:"display_name,omitempty"`      // User's full name or nickname
	AvatarURL      *string   `json:"avatar_url,omitempty"`        // User's profile picture URL
	Username       string    `json:"username"`                   // User's unique handle
	PublicShareURL *string   `json:"public_share_url,omitempty"`    // URL for user's public profile share
	Settings       any       `json:"settings,omitempty"`          // User-specific configuration (JSON)
	Status         string    `json:"status"`                     // Current account status (active/inactive)
}

