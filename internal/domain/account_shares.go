package domain

import (
	"context"
	"time"
)

type AccountShare struct {
	ID         string     `json:"id"`
	AccountID  string     `json:"account_id"`
	UserID     string     `json:"user_id"`
	Permission string     `json:"permission"`
	Status     string     `json:"status"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	CreatedBy  *string    `json:"created_by,omitempty"`
	UpdatedBy  *string    `json:"updated_by,omitempty"`

	UserEmail       *string `json:"user_email,omitempty"`
	UserPhone       *string `json:"user_phone,omitempty"`
	UserDisplayName *string `json:"user_display_name,omitempty"`
}

type AccountShareRepository interface {
	ListAccountShares(ctx context.Context, actorUserID string, accountID string) ([]AccountShare, error)
	UpsertAccountShare(ctx context.Context, actorUserID string, accountID string, targetUserID string, permission string) (*AccountShare, error)
	RevokeAccountShare(ctx context.Context, actorUserID string, accountID string, targetUserID string) error
}

