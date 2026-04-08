package entity
 
// User represents a registered member of the platform.
type User struct {
	AuditEntity
	Email          *string `json:"email,omitempty"`           // User's email address
	Phone          *string `json:"phone,omitempty"`           // User's phone number
	DisplayName    *string `json:"display_name,omitempty"`    // Name shown in the UI
	AvatarURL      *string `json:"avatar_url,omitempty"`      // URL to user's profile picture
	Username       string  `json:"username"`                 // Unique login identifier
	PublicShareURL *string `json:"public_share_url,omitempty"` // URL for sharing public profile/debts
	Settings       any     `json:"settings,omitempty"`       // User-specific application settings (JSON)
	Status         string  `json:"status"`                   // Current account status (active/suspended)
}
 
// UserWithPassword includes the user profile and their hashed password for authentication.
type UserWithPassword struct {
	User
	PasswordHash string // Bcrypt hashed password
}
 
// UpdateUserParams defines the fields that can be modified in a user's profile.
type UpdateUserParams struct {
	DisplayName  *string // New display name
	Email        *string // New email address
	Phone        *string // New phone number
	AvatarURL    *string // New avatar URL
	Username     *string // New username
	PasswordHash *string // New hashed password
}
