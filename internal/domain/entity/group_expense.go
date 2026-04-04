package entity

import "time"

type GroupExpenseParticipant struct {
	ID                      string     `json:"id"`
	UserID                  string     `json:"user_id"`
	TransactionID           string     `json:"transaction_id"`
	ParticipantName         string     `json:"participant_name"`
	OriginalAmount          string     `json:"original_amount"`
	ShareAmount             string     `json:"share_amount"`
	IsSettled               bool       `json:"is_settled"`
	SettlementTransactionID *string    `json:"settlement_transaction_id,omitempty"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
	DeletedAt               *time.Time `json:"deleted_at,omitempty"`
}
