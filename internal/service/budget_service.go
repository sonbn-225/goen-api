package service

import (
	"context"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/pkg/utils"
	"github.com/sonbn-225/goen-api/internal/pkg/apperr"
)

type BudgetService struct {
	repo         interfaces.BudgetRepository
	categoryRepo interfaces.CategoryRepository
	auditSvc     interfaces.AuditService
}

func NewBudgetService(repo interfaces.BudgetRepository, categoryRepo interfaces.CategoryRepository, auditSvc interfaces.AuditService) *BudgetService {
	return &BudgetService{repo: repo, categoryRepo: categoryRepo, auditSvc: auditSvc}
}

func (s *BudgetService) Create(ctx context.Context, userID uuid.UUID, req dto.CreateBudgetRequest) (*dto.BudgetWithStatsResponse, error) {
	period := entity.BudgetPeriod(strings.TrimSpace(string(req.Period)))
	if period != entity.BudgetPeriodMonth && period != entity.BudgetPeriodWeek && period != entity.BudgetPeriodCustom {
		return nil, apperr.BadRequest("invalid_period", "invalid period").
			WithDetail("field", "period").
			WithDetail("value", req.Period)
	}

	amount := strings.TrimSpace(req.Amount)
	if !utils.IsValidDecimal(amount) {
		return nil, apperr.BadRequest("invalid_amount", "invalid amount").
			WithDetail("field", "amount").
			WithDetail("value", amount)
	}

	if req.CategoryID == nil || *req.CategoryID == "" {
		return nil, apperr.BadRequest("missing_category", "category_id is required").
			WithDetail("field", "category_id")
	}

	categoryID, err := uuid.Parse(*req.CategoryID)
	if err != nil {
		return nil, apperr.BadRequest("invalid_category_id", "invalid category ID").
			WithDetail("field", "category_id").
			WithDetail("value", *req.CategoryID)
	}

	cat, err := s.categoryRepo.GetCategoryTx(ctx, nil, userID, categoryID)
	if err != nil {
		return nil, err
	}
	if cat == nil || !cat.IsActive {
		return nil, apperr.BadRequest("invalid_category", "category is invalid or inactive")
	}

	start, end, err := s.normalizeBudgetPeriod(string(period), req.PeriodStart, req.PeriodEnd)
	if err != nil {
		return nil, err
	}

	b := entity.Budget{
		AuditEntity: entity.AuditEntity{
			BaseEntity: entity.BaseEntity{
				ID: utils.NewID(),
			},
		},
		UserID:                userID,
		Name:                  utils.NormalizeOptionalString(req.Name),
		Period:                period,
		PeriodStart:           &start,
		PeriodEnd:             &end,
		Amount:                amount,
		AlertThresholdPercent: req.AlertThresholdPercent,
		RolloverMode:          req.RolloverMode,
		CategoryID:            &categoryID,
	}

	if err := s.repo.CreateBudgetTx(ctx, nil, userID, b); err != nil {
		return nil, err
	}

	_ = s.auditSvc.Record(ctx, nil, userID, nil, entity.ResourceBudget, entity.ActionCreated, b.ID, nil, b)

	return s.Get(ctx, userID, b.ID)
}

func (s *BudgetService) Get(ctx context.Context, userID uuid.UUID, budgetID uuid.UUID) (*dto.BudgetWithStatsResponse, error) {
	b, err := s.repo.GetBudgetTx(ctx, nil, userID, budgetID)
	if err != nil {
		return nil, err
	}
	return s.withStats(ctx, userID, *b)
}

func (s *BudgetService) List(ctx context.Context, userID uuid.UUID) ([]dto.BudgetWithStatsResponse, error) {
	items, err := s.repo.ListBudgetsTx(ctx, nil, userID)
	if err != nil {
		return nil, err
	}

	results := make([]dto.BudgetWithStatsResponse, 0, len(items))
	for _, b := range items {
		w, err := s.withStats(ctx, userID, b)
		if err == nil {
			results = append(results, *w)
		}
	}
	return results, nil
}

func (s *BudgetService) Update(ctx context.Context, userID uuid.UUID, budgetID uuid.UUID, req dto.UpdateBudgetRequest) (*dto.BudgetWithStatsResponse, error) {
	cur, err := s.repo.GetBudgetTx(ctx, nil, userID, budgetID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		cur.Name = req.Name
	}
	if req.Amount != nil {
		cur.Amount = *req.Amount
	}
	if req.AlertThresholdPercent != nil {
		cur.AlertThresholdPercent = req.AlertThresholdPercent
	}
	if req.RolloverMode != nil {
		cur.RolloverMode = req.RolloverMode
	}

	if err := s.repo.UpdateBudgetTx(ctx, nil, userID, *cur); err != nil {
		return nil, err
	}

	_ = s.auditSvc.Record(ctx, nil, userID, nil, entity.ResourceBudget, entity.ActionUpdated, budgetID, nil, cur)
	return s.Get(ctx, userID, budgetID)
}

func (s *BudgetService) Delete(ctx context.Context, userID uuid.UUID, budgetID uuid.UUID) error {
	return s.repo.DeleteBudgetTx(ctx, nil, userID, budgetID)
}

func (s *BudgetService) normalizeBudgetPeriod(period string, startIn, endIn *string) (string, string, error) {
	startStr := ""
	if startIn != nil {
		startStr = *startIn
	}
	endStr := ""
	if endIn != nil {
		endStr = *endIn
	}

	if period == "custom" {
		if startStr == "" || endStr == "" {
			return "", "", apperr.BadRequest("missing_custom_dates", "period_start and period_end are required for custom period")
		}
		return startStr, endStr, nil
	}

	if startStr == "" {
		return "", "", apperr.BadRequest("missing_start_date", "period_start is required")
	}

	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return "", "", err
	}

	var end time.Time
	switch entity.BudgetPeriod(period) {
	case entity.BudgetPeriodMonth:
		next := start.AddDate(0, 1, 0)
		end = next.AddDate(0, 0, -1)
	case entity.BudgetPeriodWeek:
		end = start.AddDate(0, 0, 6)
	default:
		return "", "", apperr.BadRequest("invalid_period", "invalid period")
	}

	return startStr, end.Format("2006-01-02"), nil
}

func (s *BudgetService) withStats(ctx context.Context, userID uuid.UUID, b entity.Budget) (*dto.BudgetWithStatsResponse, error) {
	spent := "0"
	if b.CategoryID != nil && b.PeriodStart != nil && b.PeriodEnd != nil {
		v, err := s.repo.ComputeSpentTx(ctx, nil, userID, *b.CategoryID, *b.PeriodStart, *b.PeriodEnd)
		if err == nil {
			spent = v
		}
	}

	remaining, percent := s.computeStats(b.Amount, spent)

	res := dto.NewBudgetWithStatsResponse(b)
	res.Spent = spent
	res.Remaining = remaining
	res.PercentUsed = percent
	return &res, nil
}

func (s *BudgetService) computeStats(amountStr, spentStr string) (string, int) {
	amt, ok := new(big.Rat).SetString(amountStr)
	if !ok || amt.Sign() == 0 {
		return "0", 0
	}
	spt, ok := new(big.Rat).SetString(spentStr)
	if !ok {
		spt = big.NewRat(0, 1)
	}

	rem := new(big.Rat).Sub(amt, spt)
	pct := new(big.Rat).Mul(new(big.Rat).Quo(spt, amt), big.NewRat(100, 1))

	fPct, _ := pct.Float64()
	if fPct < 0 {
		fPct = 0
	}

	return rem.FloatString(2), int(fPct)
}

