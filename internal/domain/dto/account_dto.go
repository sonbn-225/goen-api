package dto

import "time"

type CreateAccountRequest struct {
	Name            string  `json:"name"`
	AccountNumber   *string `json:"account_number,omitempty"`
	Color           *string `json:"color,omitempty"`
	AccountType     string  `json:"account_type"`
	Currency        string  `json:"currency"`
	ParentAccountID *string `json:"parent_account_id,omitempty"`
}

type UpsertShareRequest struct {
	Login      string `json:"login"`
	Permission string `json:"permission"`
}

type AccountResponse struct {
	ID                  string     `json:"id"`
	ClientID            *string    `json:"client_id,omitempty"`
	Name                string     `json:"name"`
	AccountNumber       *string    `json:"account_number,omitempty"`
	Color               *string    `json:"color,omitempty"`
	AccountType         string     `json:"account_type"`
	Currency            string     `json:"currency"`
	ParentAccountID     *string    `json:"parent_account_id,omitempty"`
	Status              string     `json:"status"`
	ClosedAt            *time.Time `json:"closed_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	Balance             string     `json:"balance"`
	InvestmentAccountID *string    `json:"investment_account_id,omitempty"`
}

type AccountBalanceResponse struct {
	AccountID string `json:"account_id"`
	Currency  string `json:"currency"`
	Balance   string `json:"balance"`
}

type AccountShareResponse struct {
	ID              string     `json:"id"`
	AccountID       string     `json:"account_id"`
	UserID          string     `json:"user_id"`
	Permission      string     `json:"permission"`
	Status          string     `json:"status"`
	RevokedAt       *time.Time `json:"revoked_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	UserEmail       *string    `json:"user_email,omitempty"`
	UserPhone       *string    `json:"user_phone,omitempty"`
	UserDisplayName *string    `json:"user_display_name,omitempty"`
}
