package transaction

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/apperrors"
	"github.com/sonbn-225/goen-api/internal/domain"
)

// CreateLineItemRequest contains line item create parameters.
type CreateLineItemRequest struct {
	CategoryID *string `json:"category_id,omitempty"`
	Amount     string  `json:"amount"`
	Note       *string `json:"note,omitempty"`
}

// CreateRequest contains transaction create parameters.
type CreateRequest struct {
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
	Notes         *string                 `json:"notes,omitempty"`
	CategoryID    *string                 `json:"category_id,omitempty"`
	TagIDs        []string                `json:"tag_ids,omitempty"`
	LineItems     []CreateLineItemRequest `json:"line_items,omitempty"`
}

// ListRequest contains transaction list filters.
type ListRequest struct {
	AccountID  *string
	CategoryID *string
	Type       *string
	Search     *string
	From       *string
	To         *string
	Cursor     *string
	Limit      int
}

// PatchRequest contains transaction patch parameters.
type PatchRequest struct {
	Description       *string                   `json:"description,omitempty"`
	Notes             *string                   `json:"notes,omitempty"`
	CategoryIDs       []string                  `json:"category_ids,omitempty"`
	TagIDs            []string                  `json:"tag_ids,omitempty"`
	Amount            *string                   `json:"amount,omitempty"`
	OccurredAt        *string                   `json:"occurred_at,omitempty"`
	LineItems         *[]LineItemInput          `json:"line_items,omitempty"`
	GroupParticipants *[]GroupParticipantInput   `json:"group_participants,omitempty"`
}

// LineItemInput is the line item payload for patch.
type LineItemInput struct {
	CategoryID *string `json:"category_id,omitempty"`
	Amount     string  `json:"amount"`
	Note       *string `json:"note,omitempty"`
}

// GroupParticipantInput is the group participant payload for patch.
type GroupParticipantInput struct {
	ParticipantName string `json:"participant_name"`
	OriginalAmount  string `json:"original_amount"`
	ShareAmount     string `json:"share_amount"`
}

// Service handles transaction business logic.
type Service struct {
	repo domain.TransactionRepository
}

// NewService creates a new transaction service.
func NewService(repo domain.TransactionRepository) *Service {
	return &Service{repo: repo}
}

// Create creates a new transaction.
func (s *Service) Create(ctx context.Context, userID string, req CreateRequest) (*domain.Transaction, error) {
	kind := strings.TrimSpace(req.Type)
	if kind != "expense" && kind != "income" && kind != "transfer" {
		return nil, apperrors.Validation("type is invalid", nil)
	}

	amount := strings.TrimSpace(req.Amount)
	if amount == "" {
		return nil, apperrors.Validation("amount is required", nil)
	}
	if !isValidDecimal(amount) {
		return nil, apperrors.Validation("amount must be a decimal string", nil)
	}

	fromAmount := normalizeOptionalString(req.FromAmount)
	toAmount := normalizeOptionalString(req.ToAmount)
	if fromAmount != nil {
		v := strings.TrimSpace(*fromAmount)
		if v == "" {
			fromAmount = nil
		} else {
			if !isValidDecimal(v) {
				return nil, apperrors.Validation("from_amount must be a decimal string", nil)
			}
			fromAmount = &v
		}
	}
	if toAmount != nil {
		v := strings.TrimSpace(*toAmount)
		if v == "" {
			toAmount = nil
		} else {
			if !isValidDecimal(v) {
				return nil, apperrors.Validation("to_amount must be a decimal string", nil)
			}
			toAmount = &v
		}
	}
	if (fromAmount != nil) != (toAmount != nil) {
		return nil, apperrors.Validation("from_amount and to_amount must be provided together", nil)
	}

	occurredAt, occurredDate, err := normalizeOccurredAt(req.OccurredAt, req.OccurredDate, req.OccurredTime)
	if err != nil {
		return nil, err
	}

	lineItems := make([]domain.TransactionLineItem, 0, len(req.LineItems))
	
	// If CategoryID is provided at top-level and no lineItems, create a default lineItem.
	if len(req.LineItems) == 0 && req.CategoryID != nil && strings.TrimSpace(*req.CategoryID) != "" {
		lineItems = append(lineItems, domain.TransactionLineItem{
			ID:         uuid.NewString(),
			CategoryID: normalizeOptionalString(req.CategoryID),
			Amount:     amount,
		})
	}

	if len(req.LineItems) > 0 {
		sum := big.NewRat(0, 1)
		for _, li := range req.LineItems {
			liAmt := strings.TrimSpace(li.Amount)
			if liAmt == "" {
				return nil, apperrors.Validation("line_items.amount is required", nil)
			}
			if !isValidDecimal(liAmt) {
				return nil, apperrors.Validation("line_items.amount must be a decimal string", nil)
			}
			r, ok := new(big.Rat).SetString(liAmt)
			if !ok {
				return nil, apperrors.Validation("line_items.amount must be a decimal string", nil)
			}
			sum.Add(sum, r)

			lineItems = append(lineItems, domain.TransactionLineItem{
				ID:         uuid.NewString(),
				CategoryID: normalizeOptionalString(li.CategoryID),
				Amount:     liAmt,
				Note:       normalizeOptionalString(li.Note),
			})
		}
		total, ok := new(big.Rat).SetString(amount)
		if !ok {
			return nil, apperrors.Validation("amount must be a decimal string", nil)
		}
		if sum.Cmp(total) != 0 {
			return nil, apperrors.Validation("line_items total must equal amount", nil)
		}
	}

	now := time.Now().UTC()
	id := uuid.NewString()

	tx := domain.Transaction{
		ID:            id,
		ClientID:      normalizeOptionalString(req.ClientID),
		ExternalRef:   normalizeOptionalString(req.ExternalRef),
		Type:          kind,
		OccurredAt:    occurredAt,
		OccurredDate:  occurredDate,
		Amount:        amount,
		FromAmount:    fromAmount,
		ToAmount:      toAmount,
		Description:   normalizeOptionalString(req.Description),
		AccountID:     normalizeOptionalString(req.AccountID),
		FromAccountID: normalizeOptionalString(req.FromAccountID),
		ToAccountID:   normalizeOptionalString(req.ToAccountID),
		ExchangeRate:  normalizeOptionalString(req.ExchangeRate),
		Notes:         normalizeOptionalString(req.Notes),
		Status:        "posted",
		CreatedAt:     now,
		UpdatedAt:     now,
		CreatedBy:     &userID,
		UpdatedBy:     &userID,
	}

	if err := validateTransactionLinkage(tx); err != nil {
		return nil, err
	}

	tagIDs := normalizeTagIDs(req.TagIDs)
	if err := s.repo.CreateTransaction(ctx, userID, tx, lineItems, tagIDs); err != nil {
		if errors.Is(err, apperrors.ErrTransactionForbidden) {
			return nil, apperrors.Wrap(apperrors.KindForbidden, "forbidden", err)
		}
		return nil, err
	}

	created, err := s.repo.GetTransaction(ctx, userID, id)
	if err != nil {
		if errors.Is(err, apperrors.ErrTransactionNotFound) {
			return nil, apperrors.Wrap(apperrors.KindNotFound, "transaction not found", err)
		}
		if errors.Is(err, apperrors.ErrTransactionForbidden) {
			return nil, apperrors.Wrap(apperrors.KindForbidden, "forbidden", err)
		}
		return nil, err
	}
	return created, nil
}

// Get retrieves a transaction by ID.
func (s *Service) Get(ctx context.Context, userID, transactionID string) (*domain.Transaction, error) {
	tx, err := s.repo.GetTransaction(ctx, userID, transactionID)
	if err != nil {
		if errors.Is(err, apperrors.ErrTransactionNotFound) {
			return nil, apperrors.Wrap(apperrors.KindNotFound, "transaction not found", err)
		}
		if errors.Is(err, apperrors.ErrTransactionForbidden) {
			return nil, apperrors.Wrap(apperrors.KindForbidden, "forbidden", err)
		}
		return nil, err
	}
	return tx, nil
}

// List returns transactions matching the filter.
func (s *Service) List(ctx context.Context, userID string, req ListRequest) ([]domain.Transaction, *string, error) {
	filter := domain.TransactionListFilter{
		AccountID:  normalizeOptionalString(req.AccountID),
		CategoryID: normalizeOptionalString(req.CategoryID),
		Type:       normalizeOptionalString(req.Type),
		Search:     normalizeOptionalString(req.Search),
		Cursor:     normalizeOptionalString(req.Cursor),
		Limit:      req.Limit,
	}

	if req.From != nil {
		v := strings.TrimSpace(*req.From)
		if v != "" {
			t, err := parseTimeOrDate(v)
			if err != nil {
				return nil, nil, apperrors.Validation("from is invalid", nil)
			}
			filter.From = &t
		}
	}
	if req.To != nil {
		v := strings.TrimSpace(*req.To)
		if v != "" {
			t, err := parseTimeOrDate(v)
			if err != nil {
				return nil, nil, apperrors.Validation("to is invalid", nil)
			}
			filter.To = &t
		}
	}

	return s.repo.ListTransactions(ctx, userID, filter)
}

// Patch updates transaction fields.
func (s *Service) Patch(ctx context.Context, userID, transactionID string, req PatchRequest) (*domain.Transaction, error) {
	patch := domain.TransactionPatch{
		Description: normalizeOptionalString(req.Description),
		Notes:       normalizeOptionalString(req.Notes),
		CategoryIDs: req.CategoryIDs,
		TagIDs:      req.TagIDs,
		Amount:      normalizeOptionalString(req.Amount),
	}

	// Convert LineItemInput → domain.TransactionLineItem
	if req.LineItems != nil {
		items := make([]domain.TransactionLineItem, len(*req.LineItems))
		for i, li := range *req.LineItems {
			items[i] = domain.TransactionLineItem{
				ID:         uuid.NewString(),
				CategoryID: li.CategoryID,
				Amount:     li.Amount,
				Note:       li.Note,
			}
		}
		patch.LineItems = &items
	}

	// Convert GroupParticipantInput → domain.GroupExpenseParticipant
	if req.GroupParticipants != nil {
		now := time.Now().UTC()
		parts := make([]domain.GroupExpenseParticipant, len(*req.GroupParticipants))
		for i, g := range *req.GroupParticipants {
			parts[i] = domain.GroupExpenseParticipant{
				ID:              uuid.NewString(),
				UserID:          userID,
				TransactionID:   transactionID,
				ParticipantName: g.ParticipantName,
				OriginalAmount:  g.OriginalAmount,
				ShareAmount:     g.ShareAmount,
				IsSettled:       false,
				CreatedAt:       now,
				UpdatedAt:       now,
			}
		}
		patch.GroupParticipants = &parts
	}

	if req.OccurredAt != nil {
		v := strings.TrimSpace(*req.OccurredAt)
		if v != "" {
			t, err := parseTimeOrDate(v)
			if err != nil {
				return nil, apperrors.Validation("occurred_at is invalid", nil)
			}
			patch.OccurredAt = &t
		}
	}

	tx, err := s.repo.PatchTransaction(ctx, userID, transactionID, patch)
	if err != nil {
		if errors.Is(err, apperrors.ErrTransactionNotFound) {
			return nil, apperrors.Wrap(apperrors.KindNotFound, "transaction not found", err)
		}
		if errors.Is(err, apperrors.ErrTransactionForbidden) {
			return nil, apperrors.Wrap(apperrors.KindForbidden, "forbidden", err)
		}
		return nil, err
	}
	return tx, nil
}

// Delete soft-deletes a transaction.
func (s *Service) Delete(ctx context.Context, userID, transactionID string) error {
	err := s.repo.DeleteTransaction(ctx, userID, transactionID)
	if err != nil {
		if errors.Is(err, apperrors.ErrTransactionNotFound) {
			return apperrors.Wrap(apperrors.KindNotFound, "transaction not found", err)
		}
		if errors.Is(err, apperrors.ErrTransactionForbidden) {
			return apperrors.Wrap(apperrors.KindForbidden, "forbidden", err)
		}
		return err
	}
	return nil
}

func validateTransactionLinkage(tx domain.Transaction) error {
	switch tx.Type {
	case "expense", "income":
		if tx.AccountID == nil {
			return apperrors.Validation("account_id is required", nil)
		}
		if tx.FromAccountID != nil || tx.ToAccountID != nil {
			return apperrors.Validation("from_account_id/to_account_id must be empty", nil)
		}
	case "transfer":
		if tx.FromAccountID == nil {
			return apperrors.Validation("from_account_id is required", nil)
		}
		if tx.ToAccountID == nil {
			return apperrors.Validation("to_account_id is required", nil)
		}
		if tx.AccountID != nil {
			return apperrors.Validation("account_id must be empty", nil)
		}
	default:
		return apperrors.Validation("type is invalid", nil)
	}
	return nil
}

func normalizeOccurredAt(occurredAt, occurredDate, occurredTime *string) (time.Time, string, error) {
	if occurredAt != nil {
		v := strings.TrimSpace(*occurredAt)
		if v != "" {
			t, err := time.Parse(time.RFC3339, v)
			if err != nil {
				return time.Time{}, "", apperrors.Validation("occurred_at is invalid", nil)
			}
			return t.UTC(), t.UTC().Format("2006-01-02"), nil
		}
	}

	if occurredDate == nil || strings.TrimSpace(*occurredDate) == "" {
		return time.Time{}, "", apperrors.Validation("occurred_date is required", nil)
	}
	d, err := time.Parse("2006-01-02", strings.TrimSpace(*occurredDate))
	if err != nil {
		return time.Time{}, "", apperrors.Validation("occurred_date is invalid", nil)
	}

	h := 0
	m := 0
	if occurredTime != nil {
		v := strings.TrimSpace(*occurredTime)
		if v != "" {
			tm, err := time.Parse("15:04", v)
			if err != nil {
				return time.Time{}, "", apperrors.Validation("occurred_time is invalid", nil)
			}
			h = tm.Hour()
			m = tm.Minute()
		}
	}

	t := time.Date(d.Year(), d.Month(), d.Day(), h, m, 0, 0, time.UTC)
	return t, t.Format("2006-01-02"), nil
}

func parseTimeOrDate(v string) (time.Time, error) {
	if strings.Contains(v, "T") {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return time.Time{}, err
		}
		return t.UTC(), nil
	}
	d, err := time.Parse("2006-01-02", v)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC), nil
}

func normalizeOptionalString(s *string) *string {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil
	}
	return &v
}

func normalizeTagIDs(ids []string) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		v := strings.TrimSpace(id)
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

func isValidDecimal(s string) bool {
	_, ok := new(big.Rat).SetString(s)
	return ok
}
