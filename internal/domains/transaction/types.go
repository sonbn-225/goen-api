package transaction

import (
	"context"
	"time"

	"github.com/sonbn-225/goen-api-v2/internal/core/money"
)

type Transaction struct {
	ID            string                `json:"id"`
	ExternalRef   *string               `json:"external_ref,omitempty"`
	UserID        string                `json:"user_id"`
	AccountID     *string               `json:"account_id,omitempty"`
	FromAccountID *string               `json:"from_account_id,omitempty"`
	ToAccountID   *string               `json:"to_account_id,omitempty"`
	Type          string                `json:"type"`
	Status        string                `json:"status,omitempty"`
	Amount        money.Amount          `json:"amount"`
	Note          string                `json:"note,omitempty"`
	OccurredAt    time.Time             `json:"occurred_at,omitempty"`
	CreatedAt     time.Time             `json:"created_at"`
	LineItems     []TransactionLineItem `json:"line_items,omitempty"`
}

type TransactionLineItem struct {
	ID         string   `json:"id"`
	CategoryID *string  `json:"category_id,omitempty"`
	TagIDs     []string `json:"tag_ids,omitempty"`
	Amount     string   `json:"amount"`
	Note       *string  `json:"note,omitempty"`
}

type ListFilter struct {
	AccountID         *string
	Status            *string
	Search            *string
	From              *time.Time
	To                *time.Time
	Type              *string
	ExternalRefFamily *string
	Limit             int
}

type BatchPatchRequest struct {
	TransactionIDs []string       `json:"transaction_ids"`
	Patch          BatchPatchData `json:"patch"`
}

type BatchPatchData struct {
	Status string `json:"status"`
}

type BatchPatchResult struct {
	UpdatedCount int      `json:"updated_count"`
	FailedCount  int      `json:"failed_count"`
	UpdatedIDs   []string `json:"updated_ids,omitempty"`
	FailedIDs    []string `json:"failed_ids,omitempty"`
}

type GroupExpenseParticipant struct {
	ID                      string    `json:"id"`
	UserID                  string    `json:"user_id"`
	TransactionID           string    `json:"transaction_id"`
	ParticipantName         string    `json:"participant_name"`
	OriginalAmount          string    `json:"original_amount"`
	ShareAmount             string    `json:"share_amount"`
	IsSettled               bool      `json:"is_settled"`
	SettlementTransactionID *string   `json:"settlement_transaction_id,omitempty"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

type CreateGroupExpenseParticipantInput struct {
	ParticipantName string        `json:"participant_name"`
	OriginalAmount  money.Amount  `json:"original_amount"`
	ShareAmount     *money.Amount `json:"share_amount,omitempty"`
}

type CreateGroupExpenseParticipant struct {
	ParticipantName string       `json:"participant_name"`
	OriginalAmount  money.Amount `json:"original_amount"`
	ShareAmount     money.Amount `json:"share_amount"`
}

type CreateTransactionLineItemInput struct {
	CategoryID *string      `json:"category_id,omitempty"`
	TagIDs     []string     `json:"tag_ids,omitempty"`
	Amount     money.Amount `json:"amount"`
	Note       *string      `json:"note,omitempty"`
}

type CreateTransactionLineItem struct {
	CategoryID *string      `json:"category_id,omitempty"`
	TagIDs     []string     `json:"tag_ids,omitempty"`
	Amount     money.Amount `json:"amount"`
	Note       *string      `json:"note,omitempty"`
}

type CreateOptions struct {
	LineItems         []CreateTransactionLineItem     `json:"line_items,omitempty"`
	GroupParticipants []CreateGroupExpenseParticipant `json:"group_participants,omitempty"`
}

type CreateInput struct {
	AccountID           *string                              `json:"account_id,omitempty"`
	FromAccountID       *string                              `json:"from_account_id,omitempty"`
	ToAccountID         *string                              `json:"to_account_id,omitempty"`
	Type                string                               `json:"type"`
	Amount              money.Amount                         `json:"amount"`
	Note                string                               `json:"note"`
	LineItems           []CreateTransactionLineItemInput     `json:"line_items,omitempty"`
	GroupParticipants   []CreateGroupExpenseParticipantInput `json:"group_participants,omitempty"`
	OwnerOriginalAmount *money.Amount                        `json:"owner_original_amount,omitempty"`
}

type UpdateTransactionLineItemInput struct {
	CategoryID *string      `json:"category_id,omitempty"`
	TagIDs     []string     `json:"tag_ids,omitempty"`
	Amount     money.Amount `json:"amount"`
	Note       *string      `json:"note,omitempty"`
}

type UpdateGroupExpenseParticipantInput struct {
	ParticipantName string       `json:"participant_name"`
	OriginalAmount  money.Amount `json:"original_amount"`
	ShareAmount     money.Amount `json:"share_amount"`
}

type UpdateInput struct {
	Note              *string                               `json:"note,omitempty"`
	LineItems         *[]UpdateTransactionLineItemInput     `json:"line_items,omitempty"`
	GroupParticipants *[]UpdateGroupExpenseParticipantInput `json:"group_participants,omitempty"`
}

type Repository interface {
	Create(ctx context.Context, tx *Transaction, opts CreateOptions) error
	Update(ctx context.Context, userID, transactionID string, input UpdateInput) (*Transaction, error)
	ListByUser(ctx context.Context, userID string, filter ListFilter) ([]Transaction, int, error)
	GetByID(ctx context.Context, userID, transactionID string) (*Transaction, error)
	BatchPatchStatus(ctx context.Context, userID string, transactionIDs []string, status string) ([]string, error)
	ListGroupParticipantsByTransaction(ctx context.Context, userID, transactionID string) ([]GroupExpenseParticipant, error)
}

type Service interface {
	Create(ctx context.Context, userID string, input CreateInput) (*Transaction, error)
	Update(ctx context.Context, userID, transactionID string, input UpdateInput) (*Transaction, error)
	List(ctx context.Context, userID string, filter ListFilter) ([]Transaction, int, error)
	Get(ctx context.Context, userID, transactionID string) (*Transaction, error)
	BatchPatchStatus(ctx context.Context, userID string, req BatchPatchRequest) (*BatchPatchResult, error)
	ListGroupParticipantsByTransaction(ctx context.Context, userID, transactionID string) ([]GroupExpenseParticipant, error)
}

type ModuleDeps struct {
	Repo    Repository
	Service Service
}

type Module struct {
	Service Service
	Handler *Handler
}
