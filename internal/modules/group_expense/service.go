package group_expense

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

type ParticipantRequest struct {
	Name           string `json:"name"`
	OriginalAmount string `json:"original_amount"`
}

type CreateRequest struct {
	ClientID            *string              `json:"client_id,omitempty"`
	ExternalRef         *string              `json:"external_ref,omitempty"`
	OccurredAt          *string              `json:"occurred_at,omitempty"`
	OccurredDate        *string              `json:"occurred_date,omitempty"`
	OccurredTime        *string              `json:"occurred_time,omitempty"`
	Amount              string               `json:"amount"`
	Description         *string              `json:"description,omitempty"`
	Notes               *string              `json:"notes,omitempty"`
	TagIDs              []string             `json:"tag_ids,omitempty"`
	AccountID           string               `json:"account_id"`
	CategoryID          string               `json:"category_id"`
	OwnerOriginalAmount *string              `json:"owner_original_amount,omitempty"`
	Participants        []ParticipantRequest `json:"participants"`
}

type CreateResponse struct {
	Transaction  domain.Transaction               `json:"transaction"`
	Participants []domain.GroupExpenseParticipant `json:"participants"`
}

type SettleRequest struct {
	OccurredAt   *string `json:"occurred_at,omitempty"`
	OccurredDate *string `json:"occurred_date,omitempty"`
	OccurredTime *string `json:"occurred_time,omitempty"`
	AccountID    string  `json:"account_id"`
}

type Service struct {
	txSvc TransactionServiceInterface
	repo  domain.GroupExpenseRepository
}

func NewService(txSvc TransactionServiceInterface, repo domain.GroupExpenseRepository) *Service {
	return &Service{txSvc: txSvc, repo: repo}
}

func (s *Service) Create(ctx context.Context, userID string, req CreateRequest) (*CreateResponse, error) {
	accountID := strings.TrimSpace(req.AccountID)
	if accountID == "" {
		return nil, apperrors.Validation("account_id is required", map[string]any{"field": "account_id"})
	}
	categoryID := strings.TrimSpace(req.CategoryID)
	if categoryID == "" {
		return nil, apperrors.Validation("category_id is required", map[string]any{"field": "category_id"})
	}

	amount := strings.TrimSpace(req.Amount)
	if amount == "" {
		return nil, apperrors.Validation("amount is required", map[string]any{"field": "amount"})
	}
	totalPaid, ok := new(big.Rat).SetString(amount)
	if !ok {
		return nil, apperrors.Validation("amount must be a decimal string", map[string]any{"field": "amount"})
	}
	if totalPaid.Cmp(new(big.Rat)) <= 0 {
		return nil, apperrors.Validation("amount must be > 0", map[string]any{"field": "amount"})
	}

	occurredAt, occurredDate, err := normalizeOccurredAt(req.OccurredAt, req.OccurredDate, req.OccurredTime)
	if err != nil {
		return nil, err
	}

	// Build involved people: owner (optional) + participants.
	type person struct {
		name        *string
		original    *big.Rat
		originalStr string
	}
	involved := []person{}

	if req.OwnerOriginalAmount != nil {
		v := strings.TrimSpace(*req.OwnerOriginalAmount)
		if v != "" {
			r, ok := new(big.Rat).SetString(v)
			if !ok {
				return nil, apperrors.Validation("owner_original_amount must be a decimal string", map[string]any{"field": "owner_original_amount"})
			}
			if r.Sign() < 0 {
				return nil, apperrors.Validation("owner_original_amount must be >= 0", map[string]any{"field": "owner_original_amount"})
			}
			if r.Sign() > 0 {
				involved = append(involved, person{name: nil, original: r, originalStr: r.FloatString(2)})
			}
		}
	}

	for _, p := range req.Participants {
		n := strings.TrimSpace(p.Name)
		if n == "" {
			continue
		}
		oa := strings.TrimSpace(p.OriginalAmount)
		if oa == "" {
			continue
		}
		r, ok := new(big.Rat).SetString(oa)
		if !ok {
			return nil, apperrors.Validation("participants.original_amount must be a decimal string", map[string]any{"field": "participants.original_amount"})
		}
		if r.Sign() <= 0 {
			continue
		}
		name := n
		involved = append(involved, person{name: &name, original: r, originalStr: r.FloatString(2)})
	}

	if len(involved) == 0 {
		return nil, apperrors.Validation("no valid participants", nil)
	}

	sumOriginal := new(big.Rat)
	for _, p := range involved {
		sumOriginal.Add(sumOriginal, p.original)
	}
	if sumOriginal.Sign() <= 0 {
		return nil, apperrors.Validation("total original amount must be > 0", nil)
	}

	// Allocate shares (scale 2 to match numeric(18,2)).
	shares := make([]*big.Rat, 0, len(involved))
	allocated := new(big.Rat)
	for i, p := range involved {
		if i < len(involved)-1 {
			raw := new(big.Rat).Mul(totalPaid, p.original)
			raw.Quo(raw, sumOriginal)
			rounded := roundRat(raw, 2)
			shares = append(shares, rounded)
			allocated.Add(allocated, rounded)
			continue
		}
		last := new(big.Rat).Sub(totalPaid, allocated)
		if last.Sign() <= 0 {
			// Fallback: at least 0.01 to avoid violating CHECK; prefer failing fast.
			return nil, apperrors.Validation("invalid share allocation", nil)
		}
		shares = append(shares, roundRat(last, 2))
	}

	// Build transaction (expense) and participant rows (exclude owner row).
	now := time.Now().UTC()
	txID := uuid.NewString()
	mergedDescription := normalizeOptionalString(req.Description)
	if mergedDescription == nil {
		mergedDescription = normalizeOptionalString(req.Notes)
	}
	tx := domain.Transaction{
		ID:           txID,
		ClientID:     normalizeOptionalString(req.ClientID),
		ExternalRef:  normalizeOptionalString(req.ExternalRef),
		Type:         "expense",
		OccurredAt:   occurredAt,
		OccurredDate: occurredDate,
		Amount:       formatRatDecimalScale(totalPaid, 2),
		Description:  mergedDescription,
		AccountID:    &accountID,
		Status:       "posted",
		CreatedAt:    now,
		UpdatedAt:    now,
		CreatedBy:    &userID,
		UpdatedBy:    &userID,
	}

	lineItems := []domain.TransactionLineItem{
		{
			ID:         uuid.NewString(),
			CategoryID: &categoryID,
			Amount:     tx.Amount,
		},
	}

	participants := []domain.GroupExpenseParticipant{}
	for i, p := range involved {
		if p.name == nil {
			continue
		}
		share := shares[i]
		if share.Sign() <= 0 {
			continue
		}
		participants = append(participants, domain.GroupExpenseParticipant{
			ID:              uuid.NewString(),
			UserID:          userID,
			TransactionID:   txID,
			ParticipantName: *p.name,
			OriginalAmount:  p.originalStr,
			ShareAmount:     formatRatDecimalScale(share, 2),
			IsSettled:       false,
			CreatedAt:       now,
			UpdatedAt:       now,
		})
	}

	if len(participants) == 0 {
		return nil, apperrors.Validation("no non-owner participants", nil)
	}

	tagIDs := normalizeTagIDs(req.TagIDs)
	if err := s.repo.CreateGroupExpense(ctx, userID, tx, lineItems, tagIDs, participants); err != nil {
		if errors.Is(err, apperrors.ErrTransactionForbidden) {
			return nil, apperrors.Wrap(apperrors.KindForbidden, "forbidden", err)
		}
		if errors.Is(err, apperrors.ErrAccountNotFound) {
			return nil, apperrors.Wrap(apperrors.KindNotFound, "account not found", err)
		}
		if errors.Is(err, apperrors.ErrAccountClosed) {
			return nil, apperrors.Wrap(apperrors.KindValidation, "account is closed", err)
		}
		return nil, err
	}

	createdTx, err := s.txSvc.Get(ctx, userID, txID)
	if err != nil {
		return nil, err
	}
	createdParticipants, err := s.repo.ListParticipantsByTransaction(ctx, userID, txID)
	if err != nil {
		return nil, err
	}
	return &CreateResponse{Transaction: *createdTx, Participants: createdParticipants}, nil
}

func (s *Service) ListByTransaction(ctx context.Context, userID, transactionID string) ([]domain.GroupExpenseParticipant, error) {
	id := strings.TrimSpace(transactionID)
	if id == "" {
		return nil, apperrors.Validation("transaction_id is required", map[string]any{"field": "transaction_id"})
	}
	return s.repo.ListParticipantsByTransaction(ctx, userID, id)
}

func (s *Service) Settle(ctx context.Context, userID, participantID string, req SettleRequest) (*domain.Transaction, error) {
	pid := strings.TrimSpace(participantID)
	if pid == "" {
		return nil, apperrors.Validation("participantId is required", map[string]any{"field": "participantId"})
	}
	accountID := strings.TrimSpace(req.AccountID)
	if accountID == "" {
		return nil, apperrors.Validation("account_id is required", map[string]any{"field": "account_id"})
	}

	occurredAt, occurredDate, err := normalizeOccurredAt(req.OccurredAt, req.OccurredDate, req.OccurredTime)
	if err != nil {
		return nil, err
	}

	// Settlement is an income transaction into the chosen account.
	now := time.Now().UTC()
	settleTxID := uuid.NewString()
	settleTx := domain.Transaction{
		ID:           settleTxID,
		Type:         "income",
		OccurredAt:   occurredAt,
		OccurredDate: occurredDate,
		Amount:       "0.00", // filled by repo from participant share_amount
		AccountID:    &accountID,
		Status:       "posted",
		CreatedAt:    now,
		UpdatedAt:    now,
		CreatedBy:    &userID,
		UpdatedBy:    &userID,
	}

	catID := "cat_def_income_reimbursement"
	settleLineItems := []domain.TransactionLineItem{
		{ID: uuid.NewString(), CategoryID: &catID, Amount: "0.00"}, // filled by repo
	}

	settleTagIDs := []string{}
	createdID, err := s.repo.SettleParticipant(ctx, userID, pid, settleTx, settleLineItems, settleTagIDs)
	if err != nil {
		if errors.Is(err, apperrors.ErrGroupExpenseParticipantNotFound) {
			return nil, apperrors.NotFound("group expense participant not found", nil)
		}
		if errors.Is(err, apperrors.ErrGroupExpenseParticipantAlreadySettled) {
			return nil, apperrors.Validation("participant already settled", nil)
		}
		if errors.Is(err, apperrors.ErrTransactionForbidden) {
			return nil, apperrors.Wrap(apperrors.KindForbidden, "forbidden", err)
		}
		return nil, err
	}

	createdTx, err := s.txSvc.Get(ctx, userID, createdID)
	if err != nil {
		return nil, err
	}
	return createdTx, nil
}

func (s *Service) ListUniqueParticipantNames(ctx context.Context, userID string, limit int) ([]string, error) {
	return s.repo.ListUniqueParticipantNames(ctx, userID, limit)
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
	h, m := 0, 0
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

func roundRat(r *big.Rat, scale int) *big.Rat {
	if r == nil {
		return big.NewRat(0, 1)
	}
	if scale < 0 {
		scale = 0
	}
	factor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil)
	num := new(big.Int).Mul(r.Num(), factor)
	den := new(big.Int).Set(r.Denom())
	q, rem := new(big.Int).QuoRem(num, den, new(big.Int))
	if rem.Sign() >= 0 {
		twoRem := new(big.Int).Mul(rem, big.NewInt(2))
		if twoRem.Cmp(den) >= 0 {
			q.Add(q, big.NewInt(1))
		}
	}
	return new(big.Rat).SetFrac(q, factor)
}

func formatRatDecimalScale(r *big.Rat, scale int) string {
	rr := roundRat(r, scale)
	return rr.FloatString(scale)
}

