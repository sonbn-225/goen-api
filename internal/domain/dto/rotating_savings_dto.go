package dto

import (
	"time"

	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

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

type RotatingSavingsGroupResponse struct {
	ID                  string   `json:"id"`
	UserID              string   `json:"user_id"`
	AccountID           string   `json:"account_id"`
	Name                string   `json:"name"`
	Currency            *string  `json:"currency"`
	MemberCount         int      `json:"member_count"`
	UserSlots           int      `json:"user_slots"`
	ContributionAmount  float64  `json:"contribution_amount"`
	PayoutCycleNo       *int     `json:"payout_cycle_no"`
	FixedInterestAmount *float64 `json:"fixed_interest_amount"`
	CycleFrequency      string   `json:"cycle_frequency"`
	StartDate           string   `json:"start_date"`
	Status              string   `json:"status"`
}

func NewRotatingSavingsGroupResponse(g entity.RotatingSavingsGroup) RotatingSavingsGroupResponse {
	return RotatingSavingsGroupResponse{
		ID:                  g.ID,
		UserID:              g.UserID,
		AccountID:           g.AccountID,
		Name:                g.Name,
		Currency:            g.Currency,
		MemberCount:         g.MemberCount,
		UserSlots:           g.UserSlots,
		ContributionAmount:  g.ContributionAmount,
		PayoutCycleNo:       g.PayoutCycleNo,
		FixedInterestAmount: g.FixedInterestAmount,
		CycleFrequency:      g.CycleFrequency,
		StartDate:           g.StartDate,
		Status:              g.Status,
	}
}

func NewRotatingSavingsGroupResponses(items []entity.RotatingSavingsGroup) []RotatingSavingsGroupResponse {
	out := make([]RotatingSavingsGroupResponse, len(items))
	for i, it := range items {
		out[i] = NewRotatingSavingsGroupResponse(it)
	}
	return out
}

type RotatingSavingsGroupSummary struct {
	Group           RotatingSavingsGroupResponse `json:"group"`
	TotalPaid       float64                      `json:"total_paid"`
	TotalReceived   float64                      `json:"total_received"`
	RemainingAmount float64                      `json:"remaining_amount"`
	CompletedCycles int                          `json:"completed_cycles"`
	TotalCycles     int                          `json:"total_cycles"`
	NextDueDate     *string                      `json:"next_due_date"`
}

type RotatingSavingsContributionRequest struct {
	Kind                string  `json:"kind"` // contribution, payout, collected
	AccountID           *string `json:"account_id"`
	OccurredDate        string  `json:"occurred_date"`
	OccurredTime        *string `json:"occurred_time"`
	Amount              string  `json:"amount"`
	SlotsTaken          int     `json:"slots_taken"`
	CollectedFeePerSlot float64 `json:"collected_fee_per_slot"`
	CycleNo             *int    `json:"cycle_no"`
	DueDate             *string `json:"due_date"`
	Note                *string `json:"note"`
}

type RotatingSavingsContributionResponse struct {
	ID                  string  `json:"id"`
	GroupID             string  `json:"group_id"`
	TransactionID       string  `json:"transaction_id"`
	Kind                string  `json:"kind"`
	CycleNo             *int    `json:"cycle_no"`
	DueDate             *string `json:"due_date"`
	Amount              float64 `json:"amount"`
	SlotsTaken          int     `json:"slots_taken"`
	CollectedFeePerSlot float64 `json:"collected_fee_per_slot"`
	Note                *string `json:"note"`
}

func NewRotatingSavingsContributionResponse(c entity.RotatingSavingsContribution) RotatingSavingsContributionResponse {
	return RotatingSavingsContributionResponse{
		ID:                  c.ID,
		GroupID:             c.GroupID,
		TransactionID:       c.TransactionID,
		Kind:                c.Kind,
		CycleNo:             c.CycleNo,
		DueDate:             c.DueDate,
		Amount:              c.Amount,
		SlotsTaken:          c.SlotsTaken,
		CollectedFeePerSlot: c.CollectedFeePerSlot,
		Note:                c.Note,
	}
}

func NewRotatingSavingsContributionResponses(items []entity.RotatingSavingsContribution) []RotatingSavingsContributionResponse {
	out := make([]RotatingSavingsContributionResponse, len(items))
	for i, it := range items {
		out[i] = NewRotatingSavingsContributionResponse(it)
	}
	return out
}

type RotatingSavingsAuditLogResponse struct {
	ID        string         `json:"id"`
	UserID    string         `json:"user_id"`
	GroupID   *string        `json:"group_id"`
	Action    string         `json:"action"`
	Details   map[string]any `json:"details"`
	CreatedAt time.Time      `json:"created_at"`
}

func NewRotatingSavingsAuditLogResponse(l entity.RotatingSavingsAuditLog) RotatingSavingsAuditLogResponse {
	return RotatingSavingsAuditLogResponse{
		ID:        l.ID,
		UserID:    l.UserID,
		GroupID:   l.GroupID,
		Action:    l.Action,
		Details:   l.Details,
		CreatedAt: l.CreatedAt,
	}
}

func NewRotatingSavingsAuditLogResponses(items []entity.RotatingSavingsAuditLog) []RotatingSavingsAuditLogResponse {
	out := make([]RotatingSavingsAuditLogResponse, len(items))
	for i, it := range items {
		out[i] = NewRotatingSavingsAuditLogResponse(it)
	}
	return out
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
	Group                  RotatingSavingsGroupResponse          `json:"group"`
	Schedule               []RotatingSavingsScheduleCycle        `json:"schedule"`
	CollectedSlotsCount    int                                   `json:"collected_slots_count"`
	CurrentPayoutValue     float64                               `json:"current_payout_value"`
	CurrentAccruedInterest float64                               `json:"current_accrued_interest"`
	Contributions          []RotatingSavingsContributionResponse `json:"contributions"`
	AuditLogs              []RotatingSavingsAuditLogResponse     `json:"audit_logs"`
	TotalPaid              float64                               `json:"total_paid"`
	TotalReceived          float64                               `json:"total_received"`
	NextPayment            float64                               `json:"next_payment"`
	RemainingAmount        float64                               `json:"remaining_amount"`
}
