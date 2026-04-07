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
	"github.com/sonbn-225/goen-api/internal/pkg/utils"
)

type BudgetService struct {
	repo         interfaces.BudgetRepository
	categoryRepo interfaces.CategoryRepository
}

func NewBudgetService(repo interfaces.BudgetRepository, categoryRepo interfaces.CategoryRepository) *BudgetService {
	return &BudgetService{repo: repo, categoryRepo: categoryRepo}
}

func (s *BudgetService) Create(ctx context.Context, userID uuid.UUID, req dto.CreateBudgetRequest) (*dto.BudgetWithStatsResponse, error) {
	period := strings.TrimSpace(req.Period)
	if period != "month" && period != "week" && period != "custom" {
		return nil, errors.New("invalid period")
	}

	amount := strings.TrimSpace(req.Amount)
	if !utils.IsValidDecimal(amount) {
		return nil, errors.New("invalid amount")
	}

	if req.CategoryID == nil || *req.CategoryID == "" {
		return nil, errors.New("category_id is required")
	}

	categoryID, err := uuid.Parse(*req.CategoryID)
	if err != nil {
		return nil, errors.New("invalid category ID")
	}

	cat, err := s.categoryRepo.GetCategory(ctx, userID, categoryID)
	if err != nil {
		return nil, err
	}
	if cat == nil || !cat.IsActive {
		return nil, errors.New("category is invalid or inactive")
	}

	start, end, err := s.normalizeBudgetPeriod(period, req.PeriodStart, req.PeriodEnd)
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

	if err := s.repo.CreateBudget(ctx, userID, b); err != nil {
		return nil, err
	}

	return s.Get(ctx, userID, b.ID)
}

func (s *BudgetService) Get(ctx context.Context, userID uuid.UUID, budgetID uuid.UUID) (*dto.BudgetWithStatsResponse, error) {
	b, err := s.repo.GetBudget(ctx, userID, budgetID)
	if err != nil {
		return nil, err
	}
	return s.withStats(ctx, userID, *b)
}

func (s *BudgetService) List(ctx context.Context, userID uuid.UUID) ([]dto.BudgetWithStatsResponse, error) {
	items, err := s.repo.ListBudgets(ctx, userID)
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
	cur, err := s.repo.GetBudget(ctx, userID, budgetID)
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

	if err := s.repo.UpdateBudget(ctx, userID, *cur); err != nil {
		return nil, err
	}
	return s.Get(ctx, userID, budgetID)
}

func (s *BudgetService) Delete(ctx context.Context, userID uuid.UUID, budgetID uuid.UUID) error {
	return s.repo.DeleteBudget(ctx, userID, budgetID)
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
			return "", "", errors.New("period_start and period_end are required for custom period")
		}
		return startStr, endStr, nil
	}

	if startStr == "" {
		return "", "", errors.New("period_start is required")
	}

	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return "", "", err
	}

	var end time.Time
	switch period {
	case "month":
		next := start.AddDate(0, 1, 0)
		end = next.AddDate(0, 0, -1)
	case "week":
		end = start.AddDate(0, 0, 6)
	default:
		return "", "", errors.New("invalid period")
	}

	return startStr, end.Format("2006-01-02"), nil
}

func (s *BudgetService) withStats(ctx context.Context, userID uuid.UUID, b entity.Budget) (*dto.BudgetWithStatsResponse, error) {
	spent := "0"
	if b.CategoryID != nil && b.PeriodStart != nil && b.PeriodEnd != nil {
		v, err := s.repo.ComputeSpent(ctx, userID, *b.CategoryID, *b.PeriodStart, *b.PeriodEnd)
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

