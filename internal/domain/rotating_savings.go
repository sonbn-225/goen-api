package domain

import (
	"context"
	"time"
)

type RotatingSavingsGroup struct {
	ID                 string    `json:"id"`
	UserID             string    `json:"user_id"`
	SelfLabel          *string   `json:"self_label,omitempty"`
	AccountID          string    `json:"account_id"`
	Name               string    `json:"name"`
	Currency           string    `json:"currency"`
	MemberCount        int       `json:"member_count"`
	ContributionAmount string    `json:"contribution_amount"`
	EarlyPayoutFeeRate *string   `json:"early_payout_fee_rate,omitempty"`
	CycleFrequency     string    `json:"cycle_frequency"`
	StartDate          string    `json:"start_date"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type RotatingSavingsContribution struct {
	ID            string    `json:"id"`
	GroupID       string    `json:"group_id"`
	TransactionID string    `json:"transaction_id"`
	Kind          string    `json:"kind"`
	CycleNo       *int      `json:"cycle_no,omitempty"`
	DueDate       *string   `json:"due_date,omitempty"`
	Amount        string    `json:"amount"`
	OccurredAt    time.Time `json:"occurred_at"`
	Note          *string   `json:"note,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type RotatingSavingsRepository interface {
	CreateGroup(ctx context.Context, userID string, g RotatingSavingsGroup) error
	GetGroup(ctx context.Context, userID string, groupID string) (*RotatingSavingsGroup, error)
	ListGroups(ctx context.Context, userID string) ([]RotatingSavingsGroup, error)

	CreateContribution(ctx context.Context, userID string, c RotatingSavingsContribution) error
	ListContributions(ctx context.Context, userID string, groupID string) ([]RotatingSavingsContribution, error)
}
