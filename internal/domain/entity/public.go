package entity

import "github.com/google/uuid"

// PublicProfile represents the publicly visible information of a user.
type PublicProfile struct {
	ID          uuid.UUID `json:"id"`           // Unique user identifier
	Username    string    `json:"username"`     // Public username
	DisplayName string    `json:"display_name"` // Display name shown to others
	AvatarURL   *string   `json:"avatar_url"`   // URL to the user's avatar image
}

// PaymentInfo contains the banking details for public payment requests.
type PaymentInfo struct {
	AccountNumber string `json:"account_number"` // Bank account number for transfers
	BankName      string `json:"bank_name"`      // Name of the bank
}

// PublicDebt represents a debt record shared with a non-authenticated user.
type PublicDebt struct {
	ID          uuid.UUID `json:"id"`           // Unique debt identifier
	CreatedAt   string    `json:"created_at"`   // Date the debt was created
	ShareAmount string    `json:"share_amount"` // The amount shared with this participant
	Status      *string   `json:"status,omitempty"` // Current status of the shared debt
}


