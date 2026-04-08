package entity

import (
	"time"

	"github.com/google/uuid"
)

// Account represents a financial account (bank, wallet, cash, etc.) owned by a user.
type Account struct {
	AuditEntity
	Name                string          `json:"name"`                          // Account name (e.g., "Main Savings", "MoMo")
	AccountNumber       *string         `json:"account_number,omitempty"`       // Optional account number or identifier
	AccountType         AccountType     `json:"account_type"`                  // Type of account (bank, wallet, brokerage, etc.)
	Currency            string          `json:"currency"`                      // Primary currency (e.g., "VND", "USD")
	ParentAccountID     *uuid.UUID      `json:"parent_account_id,omitempty"`     // ID of the parent account for sub-accounts
	Status              AccountStatus   `json:"status"`                        // Current account status (active/closed)
	ClosedAt            *time.Time      `json:"closed_at,omitempty"`            // Timestamp when the account was closed
	Balance             string          `json:"balance"`                       // Current calculated balance (decimal string)
	Settings            AccountSettings `json:"settings"`                      // Specialized configuration for specific account types
}


// AccountPatch defines the fields that can be updated for an Account.
type AccountPatch struct {
	Name     *string          `json:"name,omitempty"`     // New account name
	Status   *AccountStatus   `json:"status,omitempty"`   // New account status
	Settings *AccountSettings `json:"settings,omitempty"` // Updated settings configuration
}

// AccountBalance represents the current balance of an account in a specific currency.
type AccountBalance struct {
	AccountID uuid.UUID `json:"account_id"` // ID of the account
	Currency  string    `json:"currency"`   // Currency of the balance
	Balance   string    `json:"balance"`    // Balance amount (decimal string)
}

