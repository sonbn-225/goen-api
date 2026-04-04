package transaction

import (
	"time"

	"github.com/sonbn-225/goen-api-v2/internal/core/money"
)

// HTTP contract models used by handlers and API docs.

type CreateTransactionRequest struct {
	AccountID           *string                                `json:"account_id,omitempty"`
	FromAccountID       *string                                `json:"from_account_id,omitempty"`
	ToAccountID         *string                                `json:"to_account_id,omitempty"`
	Type                string                                 `json:"type"`
	Amount              money.Amount                           `json:"amount"`
	Note                string                                 `json:"note"`
	LineItems           []CreateTransactionLineItemRequest     `json:"line_items,omitempty"`
	GroupParticipants   []CreateGroupExpenseParticipantRequest `json:"group_participants,omitempty"`
	OwnerOriginalAmount *money.Amount                          `json:"owner_original_amount,omitempty"`
}

type CreateTransactionLineItemRequest struct {
	CategoryID *string      `json:"category_id,omitempty"`
	TagIDs     []string     `json:"tag_ids,omitempty"`
	Amount     money.Amount `json:"amount"`
	Note       *string      `json:"note,omitempty"`
}

type CreateGroupExpenseParticipantRequest struct {
	ParticipantName string        `json:"participant_name"`
	OriginalAmount  money.Amount  `json:"original_amount"`
	ShareAmount     *money.Amount `json:"share_amount,omitempty"`
}

type UpdateTransactionRequest struct {
	Note              *string                                 `json:"note,omitempty"`
	LineItems         *[]UpdateTransactionLineItemRequest     `json:"line_items,omitempty"`
	GroupParticipants *[]UpdateGroupExpenseParticipantRequest `json:"group_participants,omitempty"`
}

type UpdateTransactionLineItemRequest struct {
	CategoryID *string      `json:"category_id,omitempty"`
	TagIDs     []string     `json:"tag_ids,omitempty"`
	Amount     money.Amount `json:"amount"`
	Note       *string      `json:"note,omitempty"`
}

type UpdateGroupExpenseParticipantRequest struct {
	ParticipantName string       `json:"participant_name"`
	OriginalAmount  money.Amount `json:"original_amount"`
	ShareAmount     money.Amount `json:"share_amount"`
}

type BatchPatchTransactionsRequest struct {
	TransactionIDs []string                   `json:"transaction_ids"`
	Patch          BatchPatchTransactionsData `json:"patch"`
}

type BatchPatchTransactionsData struct {
	Status string `json:"status"`
}

type ListTransactionsQuery struct {
	Limit             *int       `json:"limit,omitempty"`
	AccountID         *string    `json:"account_id,omitempty"`
	Status            *string    `json:"status,omitempty"`
	Search            *string    `json:"search,omitempty"`
	From              *time.Time `json:"from,omitempty"`
	To                *time.Time `json:"to,omitempty"`
	Type              *string    `json:"type,omitempty"`
	ExternalRefFamily *string    `json:"external_ref_family,omitempty"`
}
