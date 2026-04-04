package dto

type CreateLineItemRequest struct {
	CategoryID *string  `json:"category_id,omitempty"`
	TagIDs     []string `json:"tag_ids,omitempty"`
	Amount     string   `json:"amount"`
	Note       *string  `json:"note,omitempty"`
}

type GroupParticipantInput struct {
	ParticipantName string `json:"participant_name"`
	OriginalAmount  string `json:"original_amount"`
	ShareAmount     string `json:"share_amount"`
}

type CreateTransactionRequest struct {
	ClientID      *string                 `json:"client_id,omitempty"`
	ExternalRef   *string                 `json:"external_ref,omitempty"`
	Type          string                  `json:"type"`
	OccurredAt    *string                 `json:"occurred_at,omitempty"`
	OccurredDate  *string                 `json:"occurred_date,omitempty"`
	OccurredTime  *string                 `json:"occurred_time,omitempty"`
	Amount        string                  `json:"amount"`
	FromAmount    *string                 `json:"from_amount,omitempty"`
	ToAmount      *string                 `json:"to_amount,omitempty"`
	Description   *string                 `json:"description,omitempty"`
	AccountID     *string                 `json:"account_id,omitempty"`
	FromAccountID *string                 `json:"from_account_id,omitempty"`
	ToAccountID   *string                 `json:"to_account_id,omitempty"`
	ExchangeRate  *string                 `json:"exchange_rate,omitempty"`
	CategoryID    *string                 `json:"category_id,omitempty"`
	TagIDs              []string                `json:"tag_ids,omitempty"`
	LineItems           []CreateLineItemRequest `json:"line_items,omitempty"`
	GroupParticipants   []GroupParticipantInput `json:"group_participants,omitempty"`
	OwnerOriginalAmount *string                 `json:"owner_original_amount,omitempty"`
	Lang                string                  `json:"lang,omitempty"`
}

type TransactionPatchRequest struct {
	Description       *string                  `json:"description,omitempty"`
	CategoryIDs       []string                 `json:"category_ids,omitempty"`
	TagIDs            []string                 `json:"tag_ids,omitempty"`
	Amount            *string                  `json:"amount,omitempty"`
	Status            *string                  `json:"status,omitempty"`
	OccurredAt        *string                  `json:"occurred_at,omitempty"`
	LineItems         *[]CreateLineItemRequest `json:"line_items,omitempty"`
	GroupParticipants *[]GroupParticipantInput `json:"group_participants,omitempty"`
	Lang              string                   `json:"lang,omitempty"`
}

type BatchPatchRequest struct {
	TransactionIDs []string                `json:"transaction_ids"`
	Patch          TransactionPatchRequest `json:"patch"`
	Mode           *string                 `json:"mode,omitempty"`
}

type BatchPatchResult struct {
	Mode         string   `json:"mode"`
	UpdatedCount int      `json:"updated_count"`
	FailedCount  int      `json:"failed_count"`
	UpdatedIDs   []string `json:"updated_ids,omitempty"`
	FailedIDs    []string `json:"failed_ids,omitempty"`
}

// StageImportedItem represents a generic import item.
type StageImportedItem struct {
	TransactionDate string         `json:"transaction_date"`
	Amount          string         `json:"amount"`
	Description     *string        `json:"description,omitempty"`
	TransactionType *string        `json:"transaction_type,omitempty"`
	AccountName     *string        `json:"account_name,omitempty"`
	CategoryName    *string        `json:"category_name,omitempty"`
	Raw             map[string]any `json:"raw,omitempty"`
}

type MappingRuleInput struct {
	Kind       string `json:"kind"`
	SourceName string `json:"source_name"`
	MappedID   string `json:"mapped_id"`
}
