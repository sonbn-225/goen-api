package dto

import "github.com/sonbn-225/goen-api/internal/domain/entity"

type GroupExpenseParticipantRequest struct {
	Name           string `json:"name"`
	OriginalAmount string `json:"original_amount"`
	CreateDebt     bool   `json:"create_debt"`
}

type CreateGroupExpenseRequest struct {
	ClientID            *string                          `json:"client_id,omitempty"`
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
	Transaction  entity.Transaction               `json:"transaction"`
	Participants []entity.GroupExpenseParticipant `json:"participants"`
}

type GroupExpenseSettleRequest struct {
	OccurredAt   *string `json:"occurred_at,omitempty"`
	OccurredDate *string `json:"occurred_date,omitempty"`
	OccurredTime *string `json:"occurred_time,omitempty"`
	AccountID    string  `json:"account_id"`
}
