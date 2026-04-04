package service

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
)

type GroupExpenseService struct {
	txSvc   interfaces.TransactionService
	debtSvc interfaces.DebtService
	repo    interfaces.GroupExpenseRepository
}

func NewGroupExpenseService(txSvc interfaces.TransactionService, debtSvc interfaces.DebtService, repo interfaces.GroupExpenseRepository) *GroupExpenseService {
	return &GroupExpenseService{txSvc: txSvc, debtSvc: debtSvc, repo: repo}
}

func (s *GroupExpenseService) Create(ctx context.Context, userID string, req dto.CreateGroupExpenseRequest) (*dto.CreateGroupExpenseResponse, error) {
	accountID := strings.TrimSpace(req.AccountID)
	if accountID == "" {
		return nil, errors.New("account_id is required")
	}
	categoryID := strings.TrimSpace(req.CategoryID)
	if categoryID == "" {
		return nil, errors.New("category_id is required")
	}

	amountStr := strings.TrimSpace(req.Amount)
	if amountStr == "" {
		return nil, errors.New("amount is required")
	}
	totalPaid, ok := new(big.Rat).SetString(amountStr)
	if !ok {
		return nil, errors.New("amount must be a decimal string")
	}
	if totalPaid.Sign() <= 0 {
		return nil, errors.New("amount must be > 0")
	}

	occurredAt, occurredDate, err := s.normalizeOccurredAt(req.OccurredAt, req.OccurredDate, req.OccurredTime)
	if err != nil {
		return nil, err
	}

	// Internal person struct for calculation
	type person struct {
		name        *string
		original    *big.Rat
		originalStr string
		createDebt  bool
	}
	involved := []person{}

	// Handle Owner
	if req.OwnerOriginalAmount != nil {
		v := strings.TrimSpace(*req.OwnerOriginalAmount)
		if v != "" {
			r, ok := new(big.Rat).SetString(v)
			if !ok {
				return nil, errors.New("owner_original_amount must be a decimal string")
			}
			if r.Sign() < 0 {
				return nil, errors.New("owner_original_amount must be >= 0")
			}
			if r.Sign() > 0 {
				involved = append(involved, person{name: nil, original: r, originalStr: r.FloatString(2)})
			}
		}
	}

	// Handle Participants
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
			return nil, errors.New("participants.original_amount must be a decimal string")
		}
		if r.Sign() <= 0 {
			continue
		}
		name := n
		involved = append(involved, person{name: &name, original: r, originalStr: r.FloatString(2), createDebt: p.CreateDebt})
	}

	if len(involved) == 0 {
		return nil, errors.New("no valid participants")
	}

	sumOriginal := new(big.Rat)
	for _, p := range involved {
		sumOriginal.Add(sumOriginal, p.original)
	}
	if sumOriginal.Sign() <= 0 {
		return nil, errors.New("total original amount must be > 0")
	}

	// Allocate shares (scale 2 to match numeric(18,2))
	shares := make([]*big.Rat, 0, len(involved))
	allocated := new(big.Rat)
	for i, p := range involved {
		if i < len(involved)-1 {
			raw := new(big.Rat).Mul(totalPaid, p.original)
			raw.Quo(raw, sumOriginal)
			rounded := s.roundRat(raw, 2)
			shares = append(shares, rounded)
			allocated.Add(allocated, rounded)
			continue
		}
		last := new(big.Rat).Sub(totalPaid, allocated)
		if last.Sign() <= 0 {
			return nil, errors.New("invalid share allocation")
		}
		shares = append(shares, s.roundRat(last, 2))
	}

	// Build Transaction
	now := time.Now().UTC()
	txID := uuid.NewString()
	description := req.Description
	if description == nil || *description == "" {
		description = req.Notes
	}

	tx := entity.Transaction{
		ID:           txID,
		ClientID:     req.ClientID,
		ExternalRef:  req.ExternalRef,
		Type:         "expense",
		OccurredAt:   occurredAt,
		OccurredDate: occurredDate,
		Amount:       s.formatRatDecimalScale(totalPaid, 2),
		Description:  description,
		AccountID:    &accountID,
		Status:       "posted",
		CreatedAt:    now,
		UpdatedAt:    now,
		CreatedBy:    &userID,
		UpdatedBy:    &userID,
	}

	lineItems := []entity.TransactionLineItem{
		{
			ID:         uuid.NewString(),
			CategoryID: &categoryID,
			Amount:     tx.Amount,
		},
	}

	participants := []entity.GroupExpenseParticipant{}
	for i, p := range involved {
		if p.name == nil {
			continue
		}
		share := shares[i]
		if share.Sign() <= 0 {
			continue
		}
		participants = append(participants, entity.GroupExpenseParticipant{
			ID:              uuid.NewString(),
			UserID:          userID,
			TransactionID:   txID,
			ParticipantName: *p.name,
			OriginalAmount:  p.originalStr,
			ShareAmount:     s.formatRatDecimalScale(share, 2),
			IsSettled:       false,
			CreatedAt:       now,
			UpdatedAt:       now,
		})
	}

	// Create in Repo
	if err := s.repo.CreateGroupExpense(ctx, userID, tx, lineItems, req.TagIDs, participants); err != nil {
		return nil, err
	}

	// Create Debts if requested
	for i, p := range involved {
		if p.name == nil || !p.createDebt {
			continue
		}
		share := shares[i]
		if share.Sign() <= 0 {
			continue
		}

		foreverDate := "2099-12-31"
		debtName := *p.name
		if tx.Description != nil && *tx.Description != "" {
			debtName = *tx.Description + " (" + debtName + ")"
		}

		principal := s.formatRatDecimalScale(share, 2)
		interest := "0"

		_, _ = s.debtSvc.Create(ctx, userID, dto.CreateDebtRequest{
			AccountID:    accountID,
			Direction:    "lent",
			Name:         &debtName,
			Principal:    principal,
			StartDate:    occurredDate,
			DueDate:      foreverDate,
			InterestRate: &interest,
		})
	}

	// Fetch result
	createdTx, err := s.txSvc.Get(ctx, userID, txID)
	if err != nil {
		return nil, err
	}
	createdParticipants, err := s.repo.ListParticipantsByTransaction(ctx, userID, txID)
	if err != nil {
		return nil, err
	}

	return &dto.CreateGroupExpenseResponse{
		Transaction:  *createdTx,
		Participants: dto.NewGroupExpenseParticipantResponses(createdParticipants),
	}, nil
}

func (s *GroupExpenseService) ListByTransaction(ctx context.Context, userID, transactionID string) ([]dto.GroupExpenseParticipantResponse, error) {
	items, err := s.repo.ListParticipantsByTransaction(ctx, userID, transactionID)
	if err != nil {
		return nil, err
	}
	return dto.NewGroupExpenseParticipantResponses(items), nil
}

func (s *GroupExpenseService) Settle(ctx context.Context, userID, participantID string, req dto.GroupExpenseSettleRequest) (*dto.TransactionResponse, error) {
	occurredAt, occurredDate, err := s.normalizeOccurredAt(req.OccurredAt, req.OccurredDate, req.OccurredTime)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	settleTx := entity.Transaction{
		ID:           uuid.NewString(),
		Type:         "income",
		OccurredAt:   occurredAt,
		OccurredDate: occurredDate,
		AccountID:    &req.AccountID,
		Status:       "posted",
		CreatedAt:    now,
		UpdatedAt:    now,
		CreatedBy:    &userID,
		UpdatedBy:    &userID,
	}

	catID := "cat_def_income_reimbursement"
	settleLineItems := []entity.TransactionLineItem{
		{ID: uuid.NewString(), CategoryID: &catID, Amount: "0.00"}, // filled by repo
	}

	id, err := s.repo.SettleParticipant(ctx, userID, participantID, settleTx, settleLineItems, nil)
	if err != nil {
		return nil, err
	}

	return s.txSvc.Get(ctx, userID, id)
}

func (s *GroupExpenseService) ListUniqueParticipantNames(ctx context.Context, userID string, limit int) ([]string, error) {
	return s.repo.ListUniqueParticipantNames(ctx, userID, limit)
}

// Helpers

func (s *GroupExpenseService) normalizeOccurredAt(occurredAt, occurredDate, occurredTime *string) (time.Time, string, error) {
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
	h, m := 0, 0
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

func (s *GroupExpenseService) roundRat(r *big.Rat, scale int) *big.Rat {
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

func (s *GroupExpenseService) formatRatDecimalScale(r *big.Rat, scale int) string {
	rr := s.roundRat(r, scale)
	return rr.FloatString(scale)
}
