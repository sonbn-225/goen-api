package dto

import (
	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type GroupExpenseParticipantRequest struct {
	Name           string `json:"name"`
	OriginalAmount string `json:"original_amount"`
	CreateDebt     bool   `json:"create_debt"`
}

type CreateGroupExpenseRequest struct {
	ExternalRef         *string                          `json:"external_ref,omitempty"`
	OccurredAt          *string                          `json:"occurred_at,omitempty"`
	OccurredDate        *string                          `json:"occurred_date,omitempty"`
	OccurredTime        *string                          `json:"occurred_time,omitempty"`
	Amount              string                           `json:"amount"`
	Description         *string                          `json:"description,omitempty"`
	Notes               *string                          `json:"notes,omitempty"`
	TagIDs              []string                         `json:"tag_ids,omitempty"`
	AccountID           string                           `json:"account_id"`
	CategoryID          string                           `json:"category_id"`
	OwnerOriginalAmount *string                          `json:"owner_original_amount,omitempty"`
	Participants        []GroupExpenseParticipantRequest `json:"participants"`
}

type CreateGroupExpenseResponse struct {
	Transaction  TransactionResponse               `json:"transaction"`
	Participants []GroupExpenseParticipantResponse `json:"participants"`
}

type GroupExpenseParticipantResponse struct {
	ID                      uuid.UUID  `json:"id"`
	UserID                  uuid.UUID  `json:"user_id"`
	TransactionID           uuid.UUID  `json:"transaction_id"`
	ParticipantName         string     `json:"participant_name"`
	OriginalAmount          string     `json:"original_amount"`
	ShareAmount             string     `json:"share_amount"`
	IsSettled               bool       `json:"is_settled"`
	SettlementTransactionID *uuid.UUID `json:"settlement_transaction_id,omitempty"`
}

type GroupExpenseSettleRequest struct {
	OccurredAt   *string `json:"occurred_at,omitempty"`
	OccurredDate *string `json:"occurred_date,omitempty"`
	OccurredTime *string `json:"occurred_time,omitempty"`
	AccountID    string  `json:"account_id"`
}

func NewGroupExpenseParticipantResponse(p entity.GroupExpenseParticipant) GroupExpenseParticipantResponse {
	return GroupExpenseParticipantResponse{
		ID:                      p.ID,
		UserID:                  p.UserID,
		TransactionID:           p.TransactionID,
		ParticipantName:         p.ParticipantName,
		OriginalAmount:          p.OriginalAmount,
		ShareAmount:             p.ShareAmount,
		IsSettled:               p.IsSettled,
		SettlementTransactionID: p.SettlementTransactionID,
	}
}

func NewGroupExpenseParticipantResponses(items []entity.GroupExpenseParticipant) []GroupExpenseParticipantResponse {
	out := make([]GroupExpenseParticipantResponse, len(items))
	for i, it := range items {
		out[i] = NewGroupExpenseParticipantResponse(it)
	}
	return out
}
