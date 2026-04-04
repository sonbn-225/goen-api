package dto

import (
	"time"

	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

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
	ClientID            *string                 `json:"client_id,omitempty"`
	ExternalRef         *string                 `json:"external_ref,omitempty"`
	Type                string                  `json:"type"`
	OccurredAt          *string                 `json:"occurred_at,omitempty"`
	OccurredDate        *string                 `json:"occurred_date,omitempty"`
	OccurredTime        *string                 `json:"occurred_time,omitempty"`
	Amount              string                  `json:"amount"`
	FromAmount          *string                 `json:"from_amount,omitempty"`
	ToAmount            *string                 `json:"to_amount,omitempty"`
	Description         *string                 `json:"description,omitempty"`
	AccountID           *string                 `json:"account_id,omitempty"`
	FromAccountID       *string                 `json:"from_account_id,omitempty"`
	ToAccountID         *string                 `json:"to_account_id,omitempty"`
	ExchangeRate        *string                 `json:"exchange_rate,omitempty"`
	CategoryID          *string                 `json:"category_id,omitempty"`
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

type TransactionLineItemResponse struct {
	ID         string   `json:"id"`
	CategoryID *string  `json:"category_id,omitempty"`
	TagIDs     []string `json:"tag_ids,omitempty"`
	Amount     string   `json:"amount"`
	Note       *string  `json:"note,omitempty"`
}

type TransactionResponse struct {
	ID              string                        `json:"id"`
	ClientID        *string                       `json:"client_id,omitempty"`
	ExternalRef     *string                       `json:"external_ref,omitempty"`
	Type            string                        `json:"type"`
	OccurredAt      time.Time                     `json:"occurred_at"`
	OccurredDate    string                        `json:"occurred_date"`
	Amount          string                        `json:"amount"`
	FromAmount      *string                       `json:"from_amount,omitempty"`
	ToAmount        *string                       `json:"to_amount,omitempty"`
	AccountCurrency *string                       `json:"account_currency,omitempty"`
	FromCurrency    *string                       `json:"from_currency,omitempty"`
	ToCurrency      *string                       `json:"to_currency,omitempty"`
	Description     *string                       `json:"description,omitempty"`
	AccountID       *string                       `json:"account_id,omitempty"`
	FromAccountID   *string                       `json:"from_account_id,omitempty"`
	ToAccountID     *string                       `json:"to_account_id,omitempty"`
	ExchangeRate    *string                       `json:"exchange_rate,omitempty"`
	Status          string                        `json:"status"`
	LineItems       []TransactionLineItemResponse `json:"line_items,omitempty"`
	TagIDs          []string                      `json:"tag_ids,omitempty"`
	CategoryIDs     []string                      `json:"category_ids,omitempty"`
	CategoryNames   []string                      `json:"category_names,omitempty"`
	TagNames        []string                      `json:"tag_names,omitempty"`
	CategoryColors  []string                      `json:"category_colors,omitempty"`
	TagColors       []string                      `json:"tag_colors,omitempty"`
}

func NewTransactionResponse(t entity.Transaction) TransactionResponse {
	lineItems := make([]TransactionLineItemResponse, len(t.LineItems))
	for i, li := range t.LineItems {
		lineItems[i] = TransactionLineItemResponse{
			ID:         li.ID,
			CategoryID: li.CategoryID,
			TagIDs:     li.TagIDs,
			Amount:     li.Amount,
			Note:       li.Note,
		}
	}

	return TransactionResponse{
		ID:              t.ID,
		ClientID:        t.ClientID,
		ExternalRef:     t.ExternalRef,
		Type:            t.Type,
		OccurredAt:      t.OccurredAt,
		OccurredDate:    t.OccurredDate,
		Amount:          t.Amount,
		FromAmount:      t.FromAmount,
		ToAmount:        t.ToAmount,
		AccountCurrency: t.AccountCurrency,
		FromCurrency:    t.FromCurrency,
		ToCurrency:      t.ToCurrency,
		Description:     t.Description,
		AccountID:       t.AccountID,
		FromAccountID:   t.FromAccountID,
		ToAccountID:     t.ToAccountID,
		ExchangeRate:    t.ExchangeRate,
		Status:          t.Status,
		LineItems:       lineItems,
		TagIDs:          t.TagIDs,
		CategoryIDs:     t.CategoryIDs,
		CategoryNames:   t.CategoryNames,
		TagNames:        t.TagNames,
		CategoryColors:  t.CategoryColors,
		TagColors:       t.TagColors,
	}
}

func NewTransactionResponses(items []entity.Transaction) []TransactionResponse {
	out := make([]TransactionResponse, len(items))
	for i, it := range items {
		out[i] = NewTransactionResponse(it)
	}
	return out
}

type ImportedTransactionResponse struct {
	ID                   string         `json:"id"`
	Source               string         `json:"source"`
	TransactionDate      string         `json:"transaction_date"`
	Amount               string         `json:"amount"`
	Description          *string        `json:"description,omitempty"`
	TransactionType      *string        `json:"transaction_type,omitempty"`
	ImportedAccountName  *string        `json:"imported_account_name,omitempty"`
	ImportedCategoryName *string        `json:"imported_category_name,omitempty"`
	MappedAccountID      *string        `json:"mapped_account_id,omitempty"`
	MappedCategoryID     *string        `json:"mapped_category_id,omitempty"`
	RawPayload           map[string]any `json:"raw_payload,omitempty"`
}

func NewImportedTransactionResponse(t entity.ImportedTransaction) ImportedTransactionResponse {
	return ImportedTransactionResponse{
		ID:                   t.ID,
		Source:               t.Source,
		TransactionDate:      t.TransactionDate,
		Amount:               t.Amount,
		Description:          t.Description,
		TransactionType:      t.TransactionType,
		ImportedAccountName:  t.ImportedAccountName,
		ImportedCategoryName: t.ImportedCategoryName,
		MappedAccountID:      t.MappedAccountID,
		MappedCategoryID:     t.MappedCategoryID,
		RawPayload:           t.RawPayload,
	}
}

func NewImportedTransactionResponses(items []entity.ImportedTransaction) []ImportedTransactionResponse {
	out := make([]ImportedTransactionResponse, len(items))
	for i, it := range items {
		out[i] = NewImportedTransactionResponse(it)
	}
	return out
}

type ImportMappingRuleResponse struct {
	ID         string `json:"id"`
	Kind       string `json:"kind"`
	SourceName string `json:"source_name"`
	MappedID   string `json:"mapped_id"`
}

func NewImportMappingRuleResponse(r entity.ImportMappingRule) ImportMappingRuleResponse {
	return ImportMappingRuleResponse{
		ID:         r.ID,
		Kind:       r.Kind,
		SourceName: r.SourceName,
		MappedID:   r.MappedID,
	}
}

func NewImportMappingRuleResponses(items []entity.ImportMappingRule) []ImportMappingRuleResponse {
	out := make([]ImportMappingRuleResponse, len(items))
	for i, it := range items {
		out[i] = NewImportMappingRuleResponse(it)
	}
	return out
}
