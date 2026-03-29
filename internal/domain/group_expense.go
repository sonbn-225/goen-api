package domain

import (
	"context"
	"time"
)

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

type GroupExpenseRepository interface {
	CreateGroupExpense(ctx context.Context, userID string, tx Transaction, lineItems []TransactionLineItem, tagIDs []string, participants []GroupExpenseParticipant) error
	ListParticipantsByTransaction(ctx context.Context, userID, transactionID string) ([]GroupExpenseParticipant, error)
	SettleParticipant(ctx context.Context, userID, participantID string, settlementTx Transaction, settlementLineItems []TransactionLineItem, settlementTagIDs []string) (settlementTransactionID string, err error)
	ListUniqueParticipantNames(ctx context.Context, userID string, limit int) ([]string, error)
	ListUnsettledParticipantsByName(ctx context.Context, userID string, name string) ([]GroupExpenseParticipant, error)
}

