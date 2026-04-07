package entity

import (
	"time"

	"github.com/google/uuid"
)

// RotatingSavingsGroup represents a "Choi Ho/Hui/Phuong" group.
type RotatingSavingsGroup struct {
	AuditEntity
	UserID              uuid.UUID `json:"user_id"`
	AccountID           uuid.UUID `json:"account_id"`
	Name                string    `json:"name"`
	Currency            *string   `json:"currency"`
	MemberCount         int       `json:"member_count"`
	UserSlots           int       `json:"user_slots"`
	ContributionAmount  float64   `json:"contribution_amount"`
	PayoutCycleNo       *int      `json:"payout_cycle_no"`
	FixedInterestAmount *float64  `json:"fixed_interest_amount"`
	CycleFrequency      string    `json:"cycle_frequency"` // weekly, monthly
	StartDate           string    `json:"start_date"`      // YYYY-MM-DD
	Status              string    `json:"status"`          // active, completed, closed
}

// RotatingSavingsContribution represents an individual payment or payout in a group.
type RotatingSavingsContribution struct {
	BaseEntity
	GroupID             uuid.UUID `json:"group_id"`
	TransactionID       uuid.UUID `json:"transaction_id"`
	Kind                string    `json:"kind"` // contribution, payout, collected
	CycleNo             *int      `json:"cycle_no"`
	DueDate             *string   `json:"due_date"`
	Amount              float64   `json:"amount"`
	SlotsTaken          int       `json:"slots_taken"`
	CollectedFeePerSlot float64   `json:"collected_fee_per_slot"`
	OccurredAt          time.Time `json:"occurred_at"`
	Note                *string   `json:"note"`
	CreatedAt           time.Time `json:"created_at"`
}

// RotatingSavingsAuditLog tracks changes and actions within a group.
type RotatingSavingsAuditLog struct {
	BaseEntity
	UserID    uuid.UUID      `json:"user_id"`
	GroupID   *uuid.UUID     `json:"group_id"`
	Action    string         `json:"action"`
	Details   map[string]any `json:"details"`
	CreatedAt time.Time      `json:"created_at"`
}

