package dto

import (
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// SavingsInstrument Requests/Responses
type CreateSavingsInstrumentRequest struct {
	SavingsAccountID string  `json:"savings_account_id"`
	ParentAccountID  string  `json:"parent_account_id"`
	Principal        string  `json:"principal"`
	InterestRate     *string `json:"interest_rate,omitempty"`
	TermMonths       *int    `json:"term_months,omitempty"`
	StartDate        *string `json:"start_date,omitempty"`
	MaturityDate     *string `json:"maturity_date,omitempty"`
	AutoRenew        bool    `json:"auto_renew"`
}

// RotatingSavings Requests/Responses
type CreateRotatingSavingsGroupRequest struct {
	AccountID           string   `json:"account_id"`
	Name                string   `json:"name"`
	MemberCount         int      `json:"member_count"`
	UserSlots           int      `json:"user_slots"`
	ContributionAmount  float64  `json:"contribution_amount"`
	FixedInterestAmount *float64 `json:"fixed_interest_amount"`
	CycleFrequency      string   `json:"cycle_frequency"` // weekly, monthly
	StartDate           string   `json:"start_date"`      // YYYY-MM-DD
	Status              string   `json:"status"`
}

type UpdateRotatingSavingsGroupRequest struct {
	AccountID           *string  `json:"account_id"`
	Name                *string  `json:"name"`
	ContributionAmount  *float64 `json:"contribution_amount"`
	FixedInterestAmount *float64 `json:"fixed_interest_amount"`
	PayoutCycleNo       *int     `json:"payout_cycle_no"`
	Status              *string  `json:"status"`
}

type RotatingSavingsGroupSummary struct {
	Group           entity.RotatingSavingsGroup `json:"group"`
	TotalPaid       float64                     `json:"total_paid"`
	TotalReceived   float64                     `json:"total_received"`
	RemainingAmount float64                     `json:"remaining_amount"`
	CompletedCycles int                         `json:"completed_cycles"`
	TotalCycles     int                         `json:"total_cycles"`
	NextDueDate     *string                     `json:"next_due_date"`
}

type RotatingSavingsContributionRequest struct {
	Kind                string   `json:"kind"` // contribution, payout, collected
	AccountID           *string  `json:"account_id"`
	OccurredDate        string   `json:"occurred_date"`
	OccurredTime        *string  `json:"occurred_time"`
	Amount              string   `json:"amount"`
	SlotsTaken          int      `json:"slots_taken"`
	CollectedFeePerSlot float64  `json:"collected_fee_per_slot"`
	CycleNo             *int     `json:"cycle_no"`
	DueDate             *string  `json:"due_date"`
	Note                *string  `json:"note"`
}

type RotatingSavingsScheduleCycle struct {
	CycleNo            int     `json:"cycle_no"`
	DueDate            string  `json:"due_date"`
	ExpectedAmount     float64 `json:"expected_amount"`
	Kind               string  `json:"kind"` // uncollected, partial_collected, collected
	IsPaid             bool    `json:"is_paid"`
	ContributionID     *string `json:"contribution_id"`
	IsPayout           bool    `json:"is_payout"`
	PayoutID           *string `json:"payout_id"`
	PayoutAmount       float64 `json:"payout_amount"`
	PayoutSlots        int     `json:"payout_slots"`
	UserCollectedSlots int     `json:"user_collected_slots"`
	AccruedInterest    float64 `json:"accrued_interest"`
}

type RotatingSavingsGroupDetailResponse struct {
	Group                  entity.RotatingSavingsGroup          `json:"group"`
	Schedule               []RotatingSavingsScheduleCycle       `json:"schedule"`
	CollectedSlotsCount    int                                  `json:"collected_slots_count"`
	CurrentPayoutValue     float64                              `json:"current_payout_value"`
	CurrentAccruedInterest float64                              `json:"current_accrued_interest"`
	Contributions          []entity.RotatingSavingsContribution `json:"contributions"`
	AuditLogs              []entity.RotatingSavingsAuditLog     `json:"audit_logs"`
	TotalPaid              float64                              `json:"total_paid"`
	TotalReceived          float64                              `json:"total_received"`
	NextPayment            float64                              `json:"next_payment"`
	RemainingAmount        float64                              `json:"remaining_amount"`
}
