package entity

import (
	"time"

	"github.com/google/uuid"
)

type RotatingSavingsCycleFrequency string

const (
	RotatingSavingsCycleFrequencyWeekly  RotatingSavingsCycleFrequency = "weekly"
	RotatingSavingsCycleFrequencyMonthly RotatingSavingsCycleFrequency = "monthly"
)

type RotatingSavingsStatus string

const (
	RotatingSavingsStatusActive    RotatingSavingsStatus = "active"
	RotatingSavingsStatusCompleted RotatingSavingsStatus = "completed"
	RotatingSavingsStatusClosed    RotatingSavingsStatus = "closed"
)

// RotatingSavingsGroup represents a "Choi Ho/Hui/Phuong" group.
type RotatingSavingsGroup struct {
	AuditEntity
	UserID              uuid.UUID                     `json:"user_id"`
	AccountID           uuid.UUID                     `json:"account_id"`
	Name                string                        `json:"name"`
	Currency            *string                       `json:"currency"`
	MemberCount         int                           `json:"member_count"`
	UserSlots           int                           `json:"user_slots"`
	ContributionAmount  float64                       `json:"contribution_amount"`
	PayoutCycleNo       *int                          `json:"payout_cycle_no"`
	FixedInterestAmount *float64                      `json:"fixed_interest_amount"`
	CycleFrequency      RotatingSavingsCycleFrequency `json:"cycle_frequency"`
	StartDate           string                        `json:"start_date"` // YYYY-MM-DD
	Status              RotatingSavingsStatus         `json:"status"`
}

type RotatingSavingsContributionKind string

const (
	RotatingSavingsContributionKindContribution RotatingSavingsContributionKind = "contribution"
	RotatingSavingsContributionKindPayout       RotatingSavingsContributionKind = "payout"
	RotatingSavingsContributionKindCollected    RotatingSavingsContributionKind = "collected"
)

// RotatingSavingsContribution represents an individual payment or payout in a group.
type RotatingSavingsContribution struct {
	BaseEntity
	GroupID             uuid.UUID                       `json:"group_id"`
	TransactionID       uuid.UUID                       `json:"transaction_id"`
	Kind                RotatingSavingsContributionKind `json:"kind"`
	CycleNo             *int                            `json:"cycle_no"`
	DueDate             *string                         `json:"due_date"`
	Amount              float64                         `json:"amount"`
	SlotsTaken          int                             `json:"slots_taken"`
	CollectedFeePerSlot float64                         `json:"collected_fee_per_slot"`
	OccurredAt          time.Time                       `json:"occurred_at"`
	Note                *string                         `json:"note"`
	CreatedAt           time.Time                       `json:"created_at"`
}

type RotatingSavingsAuditAction string

const (
	RotatingSavingsAuditActionGroupCreated        RotatingSavingsAuditAction = "group_created"
	RotatingSavingsAuditActionGroupUpdated        RotatingSavingsAuditAction = "group_updated"
	RotatingSavingsAuditActionContributionCreated RotatingSavingsAuditAction = "contribution_created"
)

// RotatingSavingsAuditLog tracks changes and actions within a group.
type RotatingSavingsAuditLog struct {
	BaseEntity
	UserID    uuid.UUID                  `json:"user_id"`
	GroupID   *uuid.UUID                 `json:"group_id"`
	Action    RotatingSavingsAuditAction `json:"action"`
	Details   map[string]any             `json:"details"`
	CreatedAt time.Time                  `json:"created_at"`
}


