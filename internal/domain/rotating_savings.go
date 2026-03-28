package domain

import (
	"context"
	"time"
)

type RotatingSavingsGroup struct {
	ID                  string    `db:"id" json:"id"`
	UserID              string    `db:"user_id" json:"user_id"`
	AccountID           string    `db:"account_id" json:"account_id"`
	Name                string    `db:"name" json:"name"`
	Currency            *string   `db:"currency" json:"currency"`
	MemberCount         int       `db:"member_count" json:"member_count"`
	UserSlots           int       `db:"user_slots" json:"user_slots"`
	ContributionAmount  float64   `db:"contribution_amount" json:"contribution_amount"`
	PayoutCycleNo       *int      `db:"payout_cycle_no" json:"payout_cycle_no"`
	FixedInterestAmount *float64  `db:"fixed_interest_amount" json:"fixed_interest_amount"`
	CycleFrequency      string    `db:"cycle_frequency" json:"cycle_frequency"`
	StartDate           string    `db:"start_date" json:"start_date"`
	Status              string    `db:"status" json:"status"`
	CreatedAt           time.Time `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time `db:"updated_at" json:"updated_at"`
}

type GroupSummary struct {
	Group           RotatingSavingsGroup `json:"group"`
	TotalPaid       float64              `json:"total_paid"`
	TotalReceived   float64              `json:"total_received"`
	NetPosition     float64              `json:"net_position"`
	CompletedCycles int                  `json:"completed_cycles"`
	TotalCycles     int                  `json:"total_cycles"`
	NextDueDate     *string              `json:"next_due_date"`
}

type RotatingSavingsContribution struct {
	ID               string    `db:"id" json:"id"`
	GroupID          string    `db:"group_id" json:"group_id"`
	TransactionID    string    `db:"transaction_id" json:"transaction_id"`
	Kind             string    `db:"kind" json:"kind"`
	CycleNo          *int      `db:"cycle_no" json:"cycle_no"`
	DueDate          *string   `db:"due_date" json:"due_date"`
	Amount           float64   `db:"amount" json:"amount"`
	SlotsTaken       int       `db:"slots_taken" json:"slots_taken"`
	CollectedFeePerSlot float64 `db:"collected_fee_per_slot" json:"collected_fee_per_slot"`
	OccurredAt       time.Time `db:"occurred_at" json:"occurred_at"`
	Note             *string   `db:"note" json:"note"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
}

type RotatingSavingsRepository interface {
	CreateGroup(ctx context.Context, g RotatingSavingsGroup) error
	GetGroup(ctx context.Context, userID string, groupID string) (*RotatingSavingsGroup, error)
	UpdateGroup(ctx context.Context, g RotatingSavingsGroup) error
	DeleteGroup(ctx context.Context, userID string, groupID string) error
	ListGroups(ctx context.Context, userID string) ([]RotatingSavingsGroup, error)

	CreateContribution(ctx context.Context, c RotatingSavingsContribution) error
	GetContribution(ctx context.Context, userID string, id string) (*RotatingSavingsContribution, error)
	ListContributions(ctx context.Context, userID string, groupID string) ([]RotatingSavingsContribution, error)
	DeleteContribution(ctx context.Context, userID string, id string) error
}
