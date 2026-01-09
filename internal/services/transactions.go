package services

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain"
)

type TransactionService interface {
	Create(ctx context.Context, userID string, req CreateTransactionRequest) (*domain.Transaction, error)
	Get(ctx context.Context, userID string, transactionID string) (*domain.Transaction, error)
	List(ctx context.Context, userID string, filter ListTransactionsRequest) ([]domain.Transaction, *string, error)
	Patch(ctx context.Context, userID string, transactionID string, req PatchTransactionRequest) (*domain.Transaction, error)
	Delete(ctx context.Context, userID string, transactionID string) error
}

type CreateTransactionLineItemRequest struct {
	CategoryID *string `json:"category_id,omitempty"`
	Amount     string  `json:"amount"`
	Note       *string `json:"note,omitempty"`
}

type CreateTransactionRequest struct {
	ClientID     *string `json:"client_id,omitempty"`
	ExternalRef  *string `json:"external_ref,omitempty"`
	Type         string  `json:"type"`
	OccurredAt   *string `json:"occurred_at,omitempty"`
	OccurredDate *string `json:"occurred_date,omitempty"`
	OccurredTime *string `json:"occurred_time,omitempty"`
	Amount       string  `json:"amount"`
	FromAmount   *string `json:"from_amount,omitempty"`
	ToAmount     *string `json:"to_amount,omitempty"`
	Description  *string `json:"description,omitempty"`
	AccountID    *string `json:"account_id,omitempty"`
	FromAccountID *string `json:"from_account_id,omitempty"`
	ToAccountID   *string `json:"to_account_id,omitempty"`
	ExchangeRate *string `json:"exchange_rate,omitempty"`
	Counterparty *string `json:"counterparty,omitempty"`
	Notes        *string `json:"notes,omitempty"`
	TagIDs       []string `json:"tag_ids,omitempty"`
	LineItems    []CreateTransactionLineItemRequest `json:"line_items,omitempty"`
}

type ListTransactionsRequest struct {
	AccountID *string
	From      *string
	To        *string
	Cursor    *string
	Limit     int
}

type PatchTransactionRequest struct {
	Description  *string `json:"description,omitempty"`
	Notes        *string `json:"notes,omitempty"`
	Counterparty *string `json:"counterparty,omitempty"`
}

type transactionService struct {
	repo domain.TransactionRepository
}

func NewTransactionService(repo domain.TransactionRepository) TransactionService {
	return &transactionService{repo: repo}
}

func (s *transactionService) Create(ctx context.Context, userID string, req CreateTransactionRequest) (*domain.Transaction, error) {
	kind := strings.TrimSpace(req.Type)
	if kind != "expense" && kind != "income" && kind != "transfer" {
		return nil, errors.New("type is invalid")
	}

	amount := strings.TrimSpace(req.Amount)
	if amount == "" {
		return nil, errors.New("amount is required")
	}
	if !isValidDecimal(amount) {
		return nil, errors.New("amount must be a decimal string")
	}

	fromAmount := normalizeOptionalString(req.FromAmount)
	toAmount := normalizeOptionalString(req.ToAmount)
	if fromAmount != nil {
		v := strings.TrimSpace(*fromAmount)
		if v == "" {
			fromAmount = nil
		} else {
			if !isValidDecimal(v) {
				return nil, errors.New("from_amount must be a decimal string")
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
				return nil, errors.New("to_amount must be a decimal string")
			}
			toAmount = &v
		}
	}
	if (fromAmount != nil) != (toAmount != nil) {
		return nil, errors.New("from_amount and to_amount must be provided together")
	}

	occurredAt, occurredDate, err := normalizeOccurredAt(req.OccurredAt, req.OccurredDate, req.OccurredTime)
	if err != nil {
		return nil, err
	}

	// Optional: validate split sums
	lineItems := make([]domain.TransactionLineItem, 0, len(req.LineItems))
	if len(req.LineItems) > 0 {
		sum := big.NewRat(0, 1)
		for _, li := range req.LineItems {
			liAmt := strings.TrimSpace(li.Amount)
			if liAmt == "" {
				return nil, errors.New("line_items.amount is required")
			}
			if !isValidDecimal(liAmt) {
				return nil, errors.New("line_items.amount must be a decimal string")
			}
			r, ok := new(big.Rat).SetString(liAmt)
			if !ok {
				return nil, errors.New("line_items.amount must be a decimal string")
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
			return nil, errors.New("amount must be a decimal string")
		}
		if sum.Cmp(total) != 0 {
			return nil, errors.New("line_items total must equal amount")
		}
	}

	now := time.Now().UTC()
	id := uuid.NewString()

	tx := domain.Transaction{
		ID:           id,
		ClientID:     normalizeOptionalString(req.ClientID),
		ExternalRef:  normalizeOptionalString(req.ExternalRef),
		Type:         kind,
		OccurredAt:   occurredAt,
		OccurredDate: occurredDate,
		Amount:       amount,
		FromAmount:   fromAmount,
		ToAmount:     toAmount,
		Description:  normalizeOptionalString(req.Description),
		AccountID:    normalizeOptionalString(req.AccountID),
		FromAccountID: normalizeOptionalString(req.FromAccountID),
		ToAccountID:   normalizeOptionalString(req.ToAccountID),
		ExchangeRate: normalizeOptionalString(req.ExchangeRate),
		Counterparty: normalizeOptionalString(req.Counterparty),
		Notes:        normalizeOptionalString(req.Notes),
		Status:       "posted",
		CreatedAt:    now,
		UpdatedAt:    now,
		CreatedBy:    &userID,
		UpdatedBy:    &userID,
	}

	if err := validateTransactionLinkage(tx); err != nil {
		return nil, err
	}

	tagIDs := normalizeTagIDs(req.TagIDs)
	if err := s.repo.CreateTransaction(ctx, userID, tx, lineItems, tagIDs); err != nil {
		return nil, err
	}

	created, err := s.repo.GetTransaction(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func normalizeTagIDs(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, raw := range in {
		v := strings.TrimSpace(raw)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func (s *transactionService) Get(ctx context.Context, userID string, transactionID string) (*domain.Transaction, error) {
	return s.repo.GetTransaction(ctx, userID, transactionID)
}

func (s *transactionService) List(ctx context.Context, userID string, req ListTransactionsRequest) ([]domain.Transaction, *string, error) {
	filter := domain.TransactionListFilter{
		AccountID: normalizeOptionalString(req.AccountID),
		Cursor:    normalizeOptionalString(req.Cursor),
		Limit:     req.Limit,
	}

	if req.From != nil {
		v := strings.TrimSpace(*req.From)
		if v != "" {
			t, err := parseTimeOrDate(v)
			if err != nil {
				return nil, nil, errors.New("from is invalid")
			}
			filter.From = &t
		}
	}
	if req.To != nil {
		v := strings.TrimSpace(*req.To)
		if v != "" {
			t, err := parseTimeOrDate(v)
			if err != nil {
				return nil, nil, errors.New("to is invalid")
			}
			filter.To = &t
		}
	}

	return s.repo.ListTransactions(ctx, userID, filter)
}

func (s *transactionService) Patch(ctx context.Context, userID string, transactionID string, req PatchTransactionRequest) (*domain.Transaction, error) {
	patch := domain.TransactionPatch{
		Description:  normalizeOptionalString(req.Description),
		Notes:        normalizeOptionalString(req.Notes),
		Counterparty: normalizeOptionalString(req.Counterparty),
	}
	return s.repo.PatchTransaction(ctx, userID, transactionID, patch)
}

func (s *transactionService) Delete(ctx context.Context, userID string, transactionID string) error {
	return s.repo.DeleteTransaction(ctx, userID, transactionID)
}

func validateTransactionLinkage(tx domain.Transaction) error {
	switch tx.Type {
	case "expense", "income":
		if tx.AccountID == nil {
			return errors.New("account_id is required")
		}
		if tx.FromAccountID != nil || tx.ToAccountID != nil {
			return errors.New("from_account_id/to_account_id must be empty")
		}
	case "transfer":
		if tx.FromAccountID == nil {
			return errors.New("from_account_id is required")
		}
		if tx.ToAccountID == nil {
			return errors.New("to_account_id is required")
		}
		if tx.AccountID != nil {
			return errors.New("account_id must be empty")
		}
	default:
		return errors.New("type is invalid")
	}
	return nil
}

func normalizeOccurredAt(occurredAt, occurredDate, occurredTime *string) (time.Time, string, error) {
	if occurredAt != nil {
		v := strings.TrimSpace(*occurredAt)
		if v != "" {
			t, err := time.Parse(time.RFC3339, v)
			if err != nil {
				return time.Time{}, "", errors.New("occurred_at is invalid")
			}
			return t.UTC(), t.UTC().Format("2006-01-02"), nil
		}
	}

	if occurredDate == nil || strings.TrimSpace(*occurredDate) == "" {
		return time.Time{}, "", errors.New("occurred_date is required")
	}
	d, err := time.Parse("2006-01-02", strings.TrimSpace(*occurredDate))
	if err != nil {
		return time.Time{}, "", errors.New("occurred_date is invalid")
	}

	h := 0
	m := 0
	if occurredTime != nil {
		v := strings.TrimSpace(*occurredTime)
		if v != "" {
			tm, err := time.Parse("15:04", v)
			if err != nil {
				return time.Time{}, "", errors.New("occurred_time is invalid")
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

func isValidDecimal(s string) bool {
	_, ok := new(big.Rat).SetString(s)
	return ok
}
