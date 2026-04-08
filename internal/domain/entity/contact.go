package entity

import (
	"github.com/google/uuid"
)

// Contact represents a person or entity associated with the user for debt or group expenses.
type Contact struct {
	AuditEntity
	UserID            uuid.UUID  `json:"user_id"`                        // ID of the user who owns this contact
	Name              string     `json:"name"`                           // Name of the contact
	Email             *string    `json:"email,omitempty"`               // Optional email of the contact
	Phone             *string    `json:"phone,omitempty"`               // Optional phone number of the contact
	AvatarURL         *string    `json:"avatar_url,omitempty"`          // URL to the contact's avatar image
	LinkedUserID      *uuid.UUID `json:"linked_user_id,omitempty"`      // ID of the platform user this contact is linked to
	Notes             *string    `json:"notes,omitempty"`               // Private notes about the contact
	LinkedDisplayName *string    `json:"linked_display_name,omitempty"` // Display name of the linked user (enriched)
	LinkedAvatarURL   *string    `json:"linked_avatar_url,omitempty"`   // Avatar URL of the linked user (enriched)
}

