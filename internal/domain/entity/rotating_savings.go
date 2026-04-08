package entity

import (
	"time"

	"github.com/google/uuid"
)

// RotatingSavingsGroup represents a "Choi Ho/Hui/Phuong" group or ROSCA.
type RotatingSavingsGroup struct {
	AuditEntity
	UserID              uuid.UUID                     `json:"user_id"`              // ID of the user who created the group
	AccountID           uuid.UUID                     `json:"account_id"`           // ID of the account linked for contributions/payouts
	Name                string                        `json:"name"`                 // Name of the rotating savings group
	Currency            *string                       `json:"currency"`             // Currency used in the group
	MemberCount         int                           `json:"member_count"`         // Total number of members in the group
	UserSlots           int                           `json:"user_slots"`           // Number of slots taken by the user
	ContributionAmount  float64                       `json:"contribution_amount"`  // Amount contributed per slot per cycle
	PayoutCycleNo       *int                          `json:"payout_cycle_no"`      // Cycle number when the user receives the payout
	FixedInterestAmount *float64                      `json:"fixed_interest_amount"` // Fixed interest added to the payout (if any)
	CycleFrequency      RotatingSavingsCycleFrequency `json:"cycle_frequency"`      // How often cycles occur (weekly/monthly)
	StartDate           string                        `json:"start_date"`           // Start date of the first cycle (YYYY-MM-DD)
	Status              RotatingSavingsStatus         `json:"status"`               // Current group status (active/completed/closed)
}

// RotatingSavingsContribution represents an individual payment or payout in a group cycle.
type RotatingSavingsContribution struct {
	BaseEntity
	GroupID             uuid.UUID                       `json:"group_id"`               // ID of the rotating savings group
	TransactionID       uuid.UUID                       `json:"transaction_id"`          // ID of the financial transaction
	Kind                RotatingSavingsContributionKind `json:"kind"`                   // Type (contribution/payout/collected)
	CycleNo             *int                            `json:"cycle_no"`               // Cycle number for this payment
	DueDate             *string                         `json:"due_date"`               // Due date for the payment (YYYY-MM-DD)
	Amount              float64                         `json:"amount"`                 // Amount of the payment/payout
	SlotsTaken          int                             `json:"slots_taken"`            // Number of user slots this payment covers
	CollectedFeePerSlot float64                         `json:"collected_fee_per_slot"` // Fee per slot for the collector
	OccurredAt          time.Time                       `json:"occurred_at"`            // Timestamp of the payment
	Note                *string                         `json:"note"`                   // Optional memo
	CreatedAt           time.Time                       `json:"created_at"`             // Creation timestamp
	UpdatedAt           time.Time                       `json:"updated_at"`             // Update timestamp
}

// RotatingSavingsAuditLog tracks modifications and actions taken within a rotating savings group.
type RotatingSavingsAuditLog struct {
	BaseEntity
	UserID    uuid.UUID                  `json:"user_id"`    // ID of the user who performed the action
	GroupID   *uuid.UUID                 `json:"group_id"`   // ID of the affected group
	Action    RotatingSavingsAuditAction `json:"action"`    // Type of action (group_created/updated/etc.)
	Details   map[string]any             `json:"details"`   // Detailed changes or event data
	CreatedAt time.Time                  `json:"created_at"` // Creation timestamp
	UpdatedAt time.Time                  `json:"updated_at"` // Update timestamp
}


