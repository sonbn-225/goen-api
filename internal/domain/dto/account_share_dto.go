package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// UpsertShareRequest is used when sharing an account with another user or updating existing sharing permissions.
// Used in: AccountHandler, AccountService.UpsertShare, AccountInterface.UpsertShare
// UpsertShareRequest is used when sharing an account with another user or updating existing sharing permissions.
// Used in: AccountHandler, AccountService.UpsertShare, AccountInterface.UpsertShare
type UpsertShareRequest struct {
	Login      string `json:"login"`      // Email or username of the user to share with
	Permission string `json:"permission"` // Permission level (e.g., "read", "write", "admin")
}

// AccountShareResponse represents the sharing details of an account with another user.
// Used in: AccountHandler, AccountService (ListShares, UpsertShare), AccountInterface
type AccountShareResponse struct {
	ID              uuid.UUID                     `json:"id"`                             // Unique identifier for the share record
	AccountID       uuid.UUID                     `json:"account_id"`                     // ID of the shared account
	UserID          uuid.UUID                     `json:"user_id"`                        // ID of the user access is shared with
	Permission      entity.AccountSharePermission `json:"permission"`                   // Current permission level
	Status          entity.AccountShareStatus     `json:"status"`                       // Current sharing status (active/revoked)
	RevokedAt       *time.Time                    `json:"revoked_at,omitempty"`           // Timestamp when access was revoked
	CreatedAt       time.Time                     `json:"created_at"`                     // Timestamp when the share was created
	UpdatedAt       time.Time                     `json:"updated_at"`                     // Timestamp of the last update
	UserEmail       *string                       `json:"user_email,omitempty"`           // Email of the shared-to user (enriched)
	UserPhone       *string                       `json:"user_phone,omitempty"`           // Phone of the shared-to user (enriched)
	UserDisplayName *string                       `json:"user_display_name,omitempty"`    // Display name of the shared-to user (enriched)
}
