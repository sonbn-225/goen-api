package dto

import (
	"github.com/google/uuid"
)

// CreateContactRequest is used when creating a new contact.
// Used in: ContactHandler, ContactService, ContactInterface
// CreateContactRequest is used when creating a new contact.
// Used in: ContactHandler, ContactService, ContactInterface
type CreateContactRequest struct {
	Name      string  `json:"name" binding:"required"` // Name of the contact
	Email     *string `json:"email,omitempty"`        // Optional email address
	Phone     *string `json:"phone,omitempty"`        // Optional phone number
	AvatarURL *string `json:"avatar_url,omitempty"`   // Optional URL to avatar image
	Notes     *string `json:"notes,omitempty"`        // Optional private notes
}

// UpdateContactRequest is used when updating an existing contact's information.
// Used in: ContactHandler, ContactService, ContactInterface
type UpdateContactRequest struct {
	Name      *string `json:"name,omitempty"`      // New name for the contact
	Email     *string `json:"email,omitempty"`     // New email address
	Phone     *string `json:"phone,omitempty"`     // New phone number
	AvatarURL *string `json:"avatar_url,omitempty"` // New avatar URL
	Notes     *string `json:"notes,omitempty"`     // New private notes
}

// ContactResponse represents the contact information sent back to the client.
// Used in: ContactHandler, ContactService, ContactInterface
type ContactResponse struct {
	ID                uuid.UUID  `json:"id"`                             // Unique contact identifier
	UserID            uuid.UUID  `json:"user_id"`                        // ID of the user who owns this contact
	Name              string     `json:"name"`                           // Name of the contact
	Email             *string    `json:"email,omitempty"`               // Email address
	Phone             *string    `json:"phone,omitempty"`               // Phone number
	AvatarURL         *string    `json:"avatar_url,omitempty"`          // URL to contact's avatar
	LinkedUserID      *uuid.UUID `json:"linked_user_id,omitempty"`      // ID of the platform user this contact is linked to
	Notes             *string    `json:"notes,omitempty"`               // Private notes
	LinkedDisplayName *string    `json:"linked_display_name,omitempty"` // Display name of the linked user (enriched)
	LinkedAvatarURL   *string    `json:"linked_avatar_url,omitempty"`   // Avatar URL of the linked user (enriched)
}
