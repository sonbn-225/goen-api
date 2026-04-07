package entity

import (
	"github.com/google/uuid"
)

type GroupExpenseParticipant struct {
	AuditEntity
	UserID                  uuid.UUID  `json:"user_id"`
	TransactionID           uuid.UUID  `json:"transaction_id"`
	ParticipantName         string     `json:"participant_name"`
	OriginalAmount          string     `json:"original_amount"`
	ShareAmount             string     `json:"share_amount"`
	IsSettled               bool       `json:"is_settled"`
	SettlementTransactionID *uuid.UUID `json:"settlement_transaction_id,omitempty"`
}

