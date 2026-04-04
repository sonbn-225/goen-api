package rotatingsavings

// HTTP contract models used by handlers and API docs.

type CreateGroupRequest struct {
	AccountID           string `json:"account_id"`
	Name                string `json:"name"`
	MemberCount         int    `json:"member_count"`
	UserSlots           int    `json:"user_slots"`
	ContributionAmount  any    `json:"contribution_amount"`
	FixedInterestAmount any    `json:"fixed_interest_amount,omitempty"`
	CycleFrequency      string `json:"cycle_frequency"`
	StartDate           string `json:"start_date"`
	Status              string `json:"status,omitempty"`
}

type UpdateGroupRequest struct {
	AccountID           *string  `json:"account_id,omitempty"`
	Name                *string  `json:"name,omitempty"`
	ContributionAmount  *float64 `json:"contribution_amount,omitempty"`
	FixedInterestAmount *float64 `json:"fixed_interest_amount,omitempty"`
	PayoutCycleNo       *int     `json:"payout_cycle_no,omitempty"`
	Status              *string  `json:"status,omitempty"`
}

type CreateContributionRequest struct {
	Kind                string  `json:"kind"`
	AccountID           *string `json:"account_id,omitempty"`
	OccurredDate        string  `json:"occurred_date"`
	OccurredTime        *string `json:"occurred_time,omitempty"`
	Amount              any     `json:"amount"`
	SlotsTaken          int     `json:"slots_taken"`
	CollectedFeePerSlot any     `json:"collected_fee_per_slot"`
	CycleNo             *int    `json:"cycle_no,omitempty"`
	DueDate             *string `json:"due_date,omitempty"`
	Note                *string `json:"note,omitempty"`
}
