package rotatingsavings

import (
	"context"
	"time"

	"github.com/sonbn-225/goen-api-v2/internal/domains/transaction"
)

type RotatingSavingsGroup struct {
	ID                  string    `json:"id"`
	UserID              string    `json:"user_id"`
	AccountID           string    `json:"account_id"`
	Name                string    `json:"name"`
	Currency            *string   `json:"currency,omitempty"`
	MemberCount         int       `json:"member_count"`
	UserSlots           int       `json:"user_slots"`
	ContributionAmount  float64   `json:"contribution_amount"`
	PayoutCycleNo       *int      `json:"payout_cycle_no,omitempty"`
	FixedInterestAmount *float64  `json:"fixed_interest_amount,omitempty"`
	CycleFrequency      string    `json:"cycle_frequency"`
	StartDate           string    `json:"start_date"`
	Status              string    `json:"status"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type RotatingSavingsContribution struct {
	ID                  string    `json:"id"`
	GroupID             string    `json:"group_id"`
	TransactionID       string    `json:"transaction_id"`
	Kind                string    `json:"kind"`
	CycleNo             *int      `json:"cycle_no,omitempty"`
	DueDate             *string   `json:"due_date,omitempty"`
	Amount              float64   `json:"amount"`
	SlotsTaken          int       `json:"slots_taken"`
	CollectedFeePerSlot float64   `json:"collected_fee_per_slot"`
	OccurredAt          time.Time `json:"occurred_at"`
	Note                *string   `json:"note,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
}

type RotatingSavingsAuditLog struct {
	ID        string         `json:"id"`
	UserID    string         `json:"user_id"`
	GroupID   *string        `json:"group_id,omitempty"`
	Action    string         `json:"action"`
	Details   map[string]any `json:"details"`
	CreatedAt time.Time      `json:"created_at"`
}

type ScheduleCycle struct {
	CycleNo            int     `json:"cycle_no"`
	DueDate            string  `json:"due_date"`
	ExpectedAmount     float64 `json:"expected_amount"`
	Kind               string  `json:"kind"`
	IsPaid             bool    `json:"is_paid"`
	ContributionID     *string `json:"contribution_id,omitempty"`
	IsPayout           bool    `json:"is_payout"`
	PayoutID           *string `json:"payout_id,omitempty"`
	PayoutAmount       float64 `json:"payout_amount"`
	PayoutSlots        int     `json:"payout_slots"`
	UserCollectedSlots int     `json:"user_collected_slots"`
	AccruedInterest    float64 `json:"accrued_interest"`
}

type GroupSummary struct {
	Group           RotatingSavingsGroup `json:"group"`
	TotalPaid       float64              `json:"total_paid"`
	TotalReceived   float64              `json:"total_received"`
	RemainingAmount float64              `json:"remaining_amount"`
	CompletedCycles int                  `json:"completed_cycles"`
	TotalCycles     int                  `json:"total_cycles"`
	NextDueDate     *string              `json:"next_due_date"`
}

type GroupDetailResponse struct {
	Group                  RotatingSavingsGroup          `json:"group"`
	Schedule               []ScheduleCycle               `json:"schedule"`
	CollectedSlotsCount    int                           `json:"collected_slots_count"`
	CurrentPayoutValue     float64                       `json:"current_payout_value"`
	CurrentAccruedInterest float64                       `json:"current_accrued_interest"`
	Contributions          []RotatingSavingsContribution `json:"contributions"`
	AuditLogs              []RotatingSavingsAuditLog     `json:"audit_logs"`
	TotalPaid              float64                       `json:"total_paid"`
	TotalReceived          float64                       `json:"total_received"`
	NextPayment            float64                       `json:"next_payment"`
	RemainingAmount        float64                       `json:"remaining_amount"`
}

type CreateGroupInput struct {
	AccountID           string
	Name                string
	MemberCount         int
	UserSlots           int
	ContributionAmount  float64
	FixedInterestAmount *float64
	CycleFrequency      string
	StartDate           string
	Status              string
}

type UpdateGroupInput struct {
	AccountID           *string
	Name                *string
	ContributionAmount  *float64
	FixedInterestAmount *float64
	PayoutCycleNo       *int
	Status              *string
}

type CreateContributionInput struct {
	Kind                string
	AccountID           *string
	OccurredDate        string
	OccurredTime        *string
	Amount              float64
	SlotsTaken          int
	CollectedFeePerSlot float64
	CycleNo             *int
	DueDate             *string
	Note                *string
}

type Repository interface {
	CreateGroup(ctx context.Context, group RotatingSavingsGroup) error
	GetGroup(ctx context.Context, userID, groupID string) (*RotatingSavingsGroup, error)
	UpdateGroup(ctx context.Context, group RotatingSavingsGroup) error
	DeleteGroup(ctx context.Context, userID, groupID string) error
	ListGroups(ctx context.Context, userID string) ([]RotatingSavingsGroup, error)

	CreateContribution(ctx context.Context, contribution RotatingSavingsContribution) error
	GetContribution(ctx context.Context, userID, contributionID string) (*RotatingSavingsContribution, error)
	ListContributions(ctx context.Context, userID, groupID string) ([]RotatingSavingsContribution, error)
	DeleteContribution(ctx context.Context, userID, contributionID string) error

	CreateAuditLog(ctx context.Context, log RotatingSavingsAuditLog) error
	ListAuditLogs(ctx context.Context, userID, groupID string) ([]RotatingSavingsAuditLog, error)

	SoftDeleteTransactionForUser(ctx context.Context, userID, transactionID string) error
}

type TransactionService interface {
	Create(ctx context.Context, userID string, input transaction.CreateInput) (*transaction.Transaction, error)
}

type Service interface {
	ListGroups(ctx context.Context, userID string) ([]GroupSummary, error)
	CreateGroup(ctx context.Context, userID string, input CreateGroupInput) (*RotatingSavingsGroup, error)
	GetGroupDetail(ctx context.Context, userID, groupID string) (*GroupDetailResponse, error)
	UpdateGroup(ctx context.Context, userID, groupID string, input UpdateGroupInput) (*RotatingSavingsGroup, error)
	DeleteGroup(ctx context.Context, userID, groupID string) error

	ListContributions(ctx context.Context, userID, groupID string) ([]RotatingSavingsContribution, error)
	CreateContribution(ctx context.Context, userID, groupID string, input CreateContributionInput) (*RotatingSavingsContribution, error)
	DeleteContribution(ctx context.Context, userID, groupID, contributionID string) error
}

type ModuleDeps struct {
	Repo      Repository
	TxService TransactionService
	Service   Service
}

type Module struct {
	Service Service
	Handler *Handler
}
