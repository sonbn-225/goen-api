package dto

import (
	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// CreateAccountRequest is used when creating a new account.
// Used in: AccountService.Create, AccountInterface.Create
type CreateAccountRequest struct {
	Name            string                  `json:"name"`                        // Name of the account (e.g., "Main Savings")
	AccountNumber   *string                 `json:"account_number,omitempty"`     // Optional account number
	AccountType     string                  `json:"account_type"`                // Type of account (e.g., "bank", "wallet")
	Currency        string                  `json:"currency"`                    // Primary currency code (e.g., "VND", "USD")
	ParentAccountID *string                 `json:"parent_account_id,omitempty"`   // ID of the parent account (for sub-accounts)
	Settings        *AccountSettingsRequest `json:"settings,omitempty"`           // Optional specialized settings for the account
}

// PatchAccountRequest is used when updating an existing account's information.
// Used in: AccountService.Patch, AccountInterface.Patch
type PatchAccountRequest struct {
	Name     *string                 `json:"name,omitempty"`     // New name for the account
	Status   *string                 `json:"status,omitempty"`   // New status (e.g., "active", "closed")
	Settings *AccountSettingsRequest `json:"settings,omitempty"` // Updated specialized settings
}

// AccountResponse represents the account information sent back to the client.
// Used in: AccountHandler, AccountService, AccountInterface
type AccountResponse struct {
	ID              uuid.UUID            `json:"id"`                       // Unique account identifier
	Name            string               `json:"name"`                     // Account name
	AccountNumber   *string              `json:"account_number,omitempty"`  // Account number
	AccountType     entity.AccountType   `json:"account_type"`             // Type of account
	Currency        string               `json:"currency"`                 // Primary currency
	ParentAccountID *uuid.UUID           `json:"parent_account_id,omitempty"` // ID of the parent account
	Status          entity.AccountStatus `json:"status"`                   // Current account status
	Balance         string               `json:"balance"`                  // Current account balance (decimal string)
	Settings        AccountSettingsResponse `json:"settings"`              // Specialized configuration (JSON)
}


// AccountBalanceResponse represents the balance of an account in a specific currency.
// Used in: ReportService, ReportDTO
type AccountBalanceResponse struct {
	AccountID uuid.UUID `json:"account_id"` // ID of the account
	Currency  string    `json:"currency"`   // Currency of the balance
	Balance   string    `json:"balance"`    // Balance amount (decimal string)
}
