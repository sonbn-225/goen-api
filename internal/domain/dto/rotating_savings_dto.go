package dto
 
import (
	"time"
 
	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)
 
// RotatingSavings Requests/Responses
// CreateRotatingSavingsGroupRequest is the payload for creating a new rotating savings group (Hui/Họ).
// Used in: RotatingSavingsHandler, RotatingSavingsService, RotatingSavingsInterface
type CreateRotatingSavingsGroupRequest struct {
	AccountID           uuid.UUID                            `json:"account_id"`           // ID of the account linked for contributions
	Name                string                               `json:"name"`                 // Name of the group
	MemberCount         int                                  `json:"member_count"`         // Total number of members in the group
	UserSlots           int                                  `json:"user_slots"`           // Number of slots taken by the user
	ContributionAmount  float64                              `json:"contribution_amount"`  // Amount per slot per cycle
	FixedInterestAmount *float64                             `json:"fixed_interest_amount"` // Optional fixed interest added to payout
	CycleFrequency      entity.RotatingSavingsCycleFrequency `json:"cycle_frequency"`      // How often cycles occur (weekly/monthly)
	StartDate           string                               `json:"start_date"`           // Start date of the first cycle (YYYY-MM-DD)
	Status              entity.RotatingSavingsStatus         `json:"status"`               // Initial status of the group
}
 
// UpdateRotatingSavingsGroupRequest is the payload for updating a rotating savings group's settings.
// Used in: RotatingSavingsHandler, RotatingSavingsService, RotatingSavingsInterface
type UpdateRotatingSavingsGroupRequest struct {
	AccountID           *uuid.UUID                    `json:"account_id"`           // New account ID
	Name                *string                       `json:"name"`                 // New group name
	ContributionAmount  *float64                      `json:"contribution_amount"`  // New contribution amount
	FixedInterestAmount *float64                      `json:"fixed_interest_amount"` // New fixed interest amount
	PayoutCycleNo       *int                          `json:"payout_cycle_no"`      // New payout cycle for the user
	Status              *entity.RotatingSavingsStatus `json:"status"`               // New group status
}
 
// RotatingSavingsGroupResponse represents a rotating savings group's basic information.
// Used in: RotatingSavingsHandler, RotatingSavingsService, RotatingSavingsInterface
type RotatingSavingsGroupResponse struct {
	ID                  uuid.UUID                            `json:"id"`                    // Unique group identifier
	UserID              uuid.UUID                            `json:"user_id"`               // ID of the user who owns the group record
	AccountID           uuid.UUID                            `json:"account_id"`            // ID of the linked funding account
	Name                string                               `json:"name"`                  // Group name
	Currency            *string                              `json:"currency"`              // Currency used (enriched)
	MemberCount         int                                  `json:"member_count"`          // Total members
	UserSlots           int                                  `json:"user_slots"`            // Slots held by user
	ContributionAmount  float64                              `json:"contribution_amount"`   // Amount per slot
	PayoutCycleNo       *int                                 `json:"payout_cycle_no"`       // User's payout cycle
	FixedInterestAmount *float64                             `json:"fixed_interest_amount"`  // Fixed interest amount
	CycleFrequency      entity.RotatingSavingsCycleFrequency `json:"cycle_frequency"`       // Weekly or Monthly
	StartDate           string                               `json:"start_date"`            // Group start date
	Status              entity.RotatingSavingsStatus         `json:"status"`                // Current group status
}
 
// RotatingSavingsGroupSummary provides a high-level summary of a group's financial state.
// Used in: RotatingSavingsHandler, RotatingSavingsService, RotatingSavingsInterface
type RotatingSavingsGroupSummary struct {
	Group           RotatingSavingsGroupResponse `json:"group"`            // Basic group information
	TotalPaid       float64                      `json:"total_paid"`       // Total amount user has contributed so far
	TotalReceived   float64                      `json:"total_received"`   // Total payout user has received
	RemainingAmount float64                      `json:"remaining_amount"` // Total amount yet to be paid (excluding payout)
	CompletedCycles int                          `json:"completed_cycles"` // Number of cycles already finished
	TotalCycles     int                          `json:"total_cycles"`     // Total cycles in the group life
	NextDueDate     *string                      `json:"next_due_date"`    // Date of the upcoming contribution (YYYY-MM-DD)
}
 
// RotatingSavingsContributionRequest is the payload for recording a contribution or payout within a group.
// Used in: RotatingSavingsHandler, RotatingSavingsService, RotatingSavingsInterface
type RotatingSavingsContributionRequest struct {
	Kind                entity.RotatingSavingsContributionKind `json:"kind"`                // Type of payment (contribution/payout/fee)
	AccountID           *uuid.UUID                             `json:"account_id"`          // ID of the account used for this payment
	OccurredDate        string                                 `json:"occurred_date"`       // Date of the payment (YYYY-MM-DD)
	OccurredTime        *string                                `json:"occurred_time"`       // Time of the payment (HH:MM:SS)
	Amount              string                                 `json:"amount"`              // Payment amount (decimal string)
	SlotsTaken          int                                    `json:"slots_taken"`         // Number of slots this payment covers
	CollectedFeePerSlot float64                                `json:"collected_fee_per_slot"` // Fee per slot (if applicable)
	CycleNo             *int                                   `json:"cycle_no"`            // Cycle number this payment belongs to
	DueDate             *string                                `json:"due_date"`            // Scheduled due date (YYYY-MM-DD)
	Note                *string                                `json:"note"`                // Optional payment memo
}
 
// RotatingSavingsContributionResponse represents a single contribution or payout event.
// Used in: RotatingSavingsHandler, RotatingSavingsService, RotatingSavingsInterface
type RotatingSavingsContributionResponse struct {
	ID                  uuid.UUID                              `json:"id"`                    // Unique payment identifier
	GroupID             uuid.UUID                              `json:"group_id"`              // ID of the parent group
	TransactionID       uuid.UUID                              `json:"transaction_id"`        // ID of the linked financial transaction
	Kind                entity.RotatingSavingsContributionKind `json:"kind"`                  // Payment type
	CycleNo             *int                                   `json:"cycle_no"`               // Associated cycle number
	DueDate             *string                                `json:"due_date"`               // Scheduled due date
	Amount              float64                                `json:"amount"`                 // Payment amount
	SlotsTaken          int                                    `json:"slots_taken"`            // Slots covered
	CollectedFeePerSlot float64                                `json:"collected_fee_per_slot"`  // Fee per slot
	Note                *string                                `json:"note"`                   // Payment memo
}
 
// RotatingSavingsAuditLogResponse represents an audit event related to a rotating savings group.
// Used in: RotatingSavingsHandler, RotatingSavingsService, RotatingSavingsInterface
type RotatingSavingsAuditLogResponse struct {
	ID        uuid.UUID                         `json:"id"`         // Unique audit event identifier
	UserID    uuid.UUID                         `json:"user_id"`    // ID of the user who performed the action
	GroupID   *uuid.UUID                        `json:"group_id"`   // ID of the affected group
	Action    entity.RotatingSavingsAuditAction `json:"action"`     // Type of action performed
	Details   map[string]any                    `json:"details"`    // Detailed event data (old/new values)
	CreatedAt time.Time                         `json:"created_at"` // Event timestamp
}
 
// RotatingSavingsScheduleCycle represents a planned cycle in a rotating savings group's schedule.
// Used in: RotatingSavingsHandler, RotatingSavingsService, RotatingSavingsInterface
type RotatingSavingsScheduleCycle struct {
	CycleNo            int        `json:"cycle_no"`            // Cycle ordinal number
	DueDate            string     `json:"due_date"`            // Scheduled date for the cycle
	ExpectedAmount     float64    `json:"expected_amount"`     // Amount user is expected to pay/receive
	Kind               string     `json:"kind"`                // Planned status (e.g., "collected")
	IsPaid             bool       `json:"is_paid"`              // Whether the contribution for this cycle is done
	ContributionID     *uuid.UUID `json:"contribution_id"`      // ID of the payment record if paid
	IsPayout           bool       `json:"is_payout"`            // Whether the user receives a payout in this cycle
	PayoutID           *uuid.UUID `json:"payout_id"`            // ID of the payout record if received
	PayoutAmount       float64    `json:"payout_amount"`        // Expected payout value
	PayoutSlots        int        `json:"payout_slots"`         // Number of slots receiving payout
	UserCollectedSlots int        `json:"user_collected_slots"` // Number of user slots already paid for this cycle
	AccruedInterest    float64    `json:"accrued_interest"`    // Interest earned up to this cycle
}
 
// RotatingSavingsGroupDetailResponse provides the full details of a rotating savings group.
// Used in: RotatingSavingsHandler, RotatingSavingsService, RotatingSavingsInterface
type RotatingSavingsGroupDetailResponse struct {
	Group                  RotatingSavingsGroupResponse          `json:"group"`                    // Basic group information
	Schedule               []RotatingSavingsScheduleCycle        `json:"schedule"`                 // Full group lifecycle schedule
	CollectedSlotsCount    int                                   `json:"collected_slots_count"`     // Total slots user has paid for across all cycles
	CurrentPayoutValue     float64                               `json:"current_payout_value"`      // Theoretical payout value if taken now
	CurrentAccruedInterest float64                               `json:"current_accrued_interest"`  // Total interest accrued so far
	Contributions          []RotatingSavingsContributionResponse `json:"contributions"`           // List of all individual payments
	AuditLogs              []RotatingSavingsAuditLogResponse     `json:"audit_logs"`               // Change history of the group
	TotalPaid              float64                               `json:"total_paid"`               // Sum of all contributions
	TotalReceived          float64                               `json:"total_received"`           // Sum of all payouts received
	NextPayment            float64                               `json:"next_payment"`             // Amount due for the next cycle
	RemainingAmount        float64                               `json:"remaining_amount"`         // Total remaining liability
}
