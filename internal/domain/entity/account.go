package entity

import (
	"time"

	"github.com/google/uuid"
)

type Account struct {
	AuditEntity
	Name                string     `json:"name"`
	AccountNumber       *string    `json:"account_number,omitempty"`
	Color               *string    `json:"color,omitempty"`
	AccountType         string     `json:"account_type"`
	Currency            string     `json:"currency"`
	ParentAccountID     *uuid.UUID `json:"parent_account_id,omitempty"`
	Status              string     `json:"status"`
	ClosedAt            *time.Time `json:"closed_at,omitempty"`
	Balance             string     `json:"balance"`               // Joined
	InvestmentAccountID *uuid.UUID `json:"investment_account_id"` // Joined
}

type AccountPatch struct {
	Name   *string `json:"name,omitempty"`
	Color  *string `json:"color,omitempty"`
	Status *string `json:"status,omitempty"`
}

type AccountBalance struct {
	AccountID uuid.UUID `json:"account_id"`
	Currency  string    `json:"currency"`
	Balance   string    `json:"balance"`
}

type AccountShare struct {
	AuditEntity
	AccountID       uuid.UUID  `json:"account_id"`
	UserID          uuid.UUID  `json:"user_id"`
	Permission      string     `json:"permission"`
	Status          string     `json:"status"`
	RevokedAt       *time.Time `json:"revoked_at,omitempty"`
	UserEmail       *string    `json:"user_email,omitempty"`
	UserPhone       *string    `json:"user_phone,omitempty"`
	UserDisplayName *string    `json:"user_display_name,omitempty"`
}

type AccountAuditEvent struct {
	BaseEntity
	AccountID   uuid.UUID      `json:"account_id"`
	ActorUserID uuid.UUID      `json:"actor_user_id"`
	Action      string         `json:"action"`
	EntityType  string         `json:"entity_type"`
	EntityID    uuid.UUID      `json:"entity_id"`
	OccurredAt  time.Time      `json:"occurred_at"`
	Diff        map[string]any `json:"diff,omitempty"`
}
