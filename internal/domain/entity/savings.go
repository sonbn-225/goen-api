package entity

import (
	"time"
)

// SavingsInstrument represents a simple savings product like a Term Deposit or Goal.
type SavingsInstrument struct {
	ID               string     `json:"id"`
	SavingsAccountID string     `json:"savings_account_id"`
	ParentAccountID  string     `json:"parent_account_id"`
	Principal        string     `json:"principal"`
	InterestRate     *string    `json:"interest_rate,omitempty"`
	TermMonths       *int       `json:"term_months,omitempty"`
	StartDate        *string    `json:"start_date,omitempty"`
	MaturityDate     *string    `json:"maturity_date,omitempty"`
	AutoRenew        bool       `json:"auto_renew"`
	AccruedInterest  string     `json:"accrued_interest"`
	Status           string     `json:"status"`
	ClosedAt         *time.Time `json:"closed_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// RotatingSavingsGroup represents a "Chơi Họ/Hụi/Phường" group.
type RotatingSavingsGroup struct {
	ID                  string    `json:"id"`
	UserID              string    `json:"user_id"`
	AccountID           string    `json:"account_id"`
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
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// RotatingSavingsContribution represents an individual payment or payout in a group.
type RotatingSavingsContribution struct {
	ID                  string    `json:"id"`
	GroupID             string    `json:"group_id"`
	TransactionID       string    `json:"transaction_id"`
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
	ID        string         `json:"id"`
	UserID    string         `json:"user_id"`
	GroupID   *string        `json:"group_id"`
	Action    string         `json:"action"`
	Details   map[string]any `json:"details"`
	CreatedAt time.Time      `json:"created_at"`
}
