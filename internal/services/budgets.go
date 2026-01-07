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

type BudgetService interface {
	Create(ctx context.Context, userID string, req CreateBudgetRequest) (*BudgetWithStats, error)
	Get(ctx context.Context, userID string, budgetID string) (*BudgetWithStats, error)
	List(ctx context.Context, userID string) ([]BudgetWithStats, error)
}

type BudgetWithStats struct {
	domain.Budget
	Spent       string `json:"spent"`
	Remaining   string `json:"remaining"`
	PercentUsed int    `json:"percent_used"`
}

type CreateBudgetRequest struct {
	Name                  *string `json:"name,omitempty"`
	Period                string  `json:"period"`
	PeriodStart           *string `json:"period_start,omitempty"`
	PeriodEnd             *string `json:"period_end,omitempty"`
	Amount                string  `json:"amount"`
	Currency              *string `json:"currency,omitempty"`
	AlertThresholdPercent *int    `json:"alert_threshold_percent,omitempty"`
	RolloverMode          *string `json:"rollover_mode,omitempty"`
	CategoryID            *string `json:"category_id,omitempty"`
}

type budgetService struct {
	repo        domain.BudgetRepository
	categoryRepo domain.CategoryRepository
}

func NewBudgetService(repo domain.BudgetRepository, categoryRepo domain.CategoryRepository) BudgetService {
	return &budgetService{repo: repo, categoryRepo: categoryRepo}
}

func (s *budgetService) Create(ctx context.Context, userID string, req CreateBudgetRequest) (*BudgetWithStats, error) {
	period := strings.TrimSpace(req.Period)
	if period != "month" && period != "week" && period != "custom" {
		return nil, errors.New("period is invalid")
	}

	amount := strings.TrimSpace(req.Amount)
	if amount == "" {
		return nil, errors.New("amount is required")
	}
	if !isValidDecimalLocal(amount) {
		return nil, errors.New("amount must be a decimal string")
	}

	categoryID := normalizeOptionalString(req.CategoryID)
	if categoryID == nil {
		return nil, errors.New("category_id is required")
	}
	if _, err := s.categoryRepo.GetCategory(ctx, userID, *categoryID); err != nil {
		return nil, err
	}

	start, end, err := normalizeBudgetPeriod(period, req.PeriodStart, req.PeriodEnd)
	if err != nil {
		return nil, err
	}

	alert := req.AlertThresholdPercent
	if alert != nil {
		if *alert < 0 || *alert > 100 {
			return nil, errors.New("alert_threshold_percent must be between 0 and 100")
		}
	} else {
		v := 80
		alert = &v
	}

	over := normalizeOptionalString(req.RolloverMode)
	if over != nil {
		v := *over
		if v != "reset" && v != "carry_forward" && v != "accumulate" {
			return nil, errors.New("rollover_mode is invalid")
		}
	}

	now := time.Now().UTC()
	b := domain.Budget{
		ID:                   uuid.NewString(),
		UserID:               userID,
		Name:                 normalizeOptionalString(req.Name),
		Period:               period,
		PeriodStart:          &start,
		PeriodEnd:            &end,
		Amount:               amount,
		Currency:             normalizeOptionalString(req.Currency),
		AlertThresholdPercent: alert,
		RolloverMode:         over,
		CategoryID:           categoryID,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	if err := s.repo.CreateBudget(ctx, userID, b); err != nil {
		return nil, err
	}

	created, err := s.repo.GetBudget(ctx, userID, b.ID)
	if err != nil {
		return nil, err
	}
	return s.withStats(ctx, userID, *created)
}

func (s *budgetService) Get(ctx context.Context, userID string, budgetID string) (*BudgetWithStats, error) {
	b, err := s.repo.GetBudget(ctx, userID, budgetID)
	if err != nil {
		return nil, err
	}
	return s.withStats(ctx, userID, *b)
}

func (s *budgetService) List(ctx context.Context, userID string) ([]BudgetWithStats, error) {
	items, err := s.repo.ListBudgets(ctx, userID)
	if err != nil {
		return nil, err
	}

	out := make([]BudgetWithStats, 0, len(items))
	for _, b := range items {
		w, err := s.withStats(ctx, userID, b)
		if err != nil {
			return nil, err
		}
		out = append(out, *w)
	}
	return out, nil
}

func (s *budgetService) withStats(ctx context.Context, userID string, b domain.Budget) (*BudgetWithStats, error) {
	categoryID := ""
	if b.CategoryID != nil {
		categoryID = strings.TrimSpace(*b.CategoryID)
	}
	start := ""
	end := ""
	if b.PeriodStart != nil {
		start = strings.TrimSpace(*b.PeriodStart)
	}
	if b.PeriodEnd != nil {
		end = strings.TrimSpace(*b.PeriodEnd)
	}

	spent := "0"
	if categoryID != "" && start != "" && end != "" {
		v, err := s.repo.ComputeSpent(ctx, userID, categoryID, start, end)
		if err != nil {
			return nil, err
		}
		spent = v
	}

	remaining, percent := computeRemainingAndPercent(b.Amount, spent)

	return &BudgetWithStats{
		Budget:      b,
		Spent:       spent,
		Remaining:   remaining,
		PercentUsed: percent,
	}, nil
}

func isValidDecimalLocal(v string) bool {
	_, ok := new(big.Rat).SetString(v)
	return ok
}

func normalizeBudgetPeriod(period string, startIn *string, endIn *string) (string, string, error) {
	startStr := strings.TrimSpace(derefOrEmpty(startIn))
	endStr := strings.TrimSpace(derefOrEmpty(endIn))

	if period == "custom" {
		if startStr == "" {
			return "", "", errors.New("period_start is required")
		}
		if endStr == "" {
			return "", "", errors.New("period_end is required")
		}
		start, err := time.Parse("2006-01-02", startStr)
		if err != nil {
			return "", "", errors.New("period_start is invalid")
		}
		end, err := time.Parse("2006-01-02", endStr)
		if err != nil {
			return "", "", errors.New("period_end is invalid")
		}
		if end.Before(start) {
			return "", "", errors.New("period_end must be >= period_start")
		}
		return startStr, endStr, nil
	}

	if startStr == "" {
		return "", "", errors.New("period_start is required")
	}
	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return "", "", errors.New("period_start is invalid")
	}

	var end time.Time
	switch period {
	case "month":
		firstOfNext := time.Date(start.Year(), start.Month()+1, 1, 0, 0, 0, 0, time.UTC)
		end = firstOfNext.Add(-24 * time.Hour)
	case "week":
		end = start.AddDate(0, 0, 6)
	default:
		return "", "", errors.New("period is invalid")
	}

	return startStr, end.Format("2006-01-02"), nil
}

func computeRemainingAndPercent(amountStr string, spentStr string) (string, int) {
	amount, ok := new(big.Rat).SetString(strings.TrimSpace(amountStr))
	if !ok || amount.Sign() == 0 {
		return amountStr, 0
	}
	spent, ok := new(big.Rat).SetString(strings.TrimSpace(spentStr))
	if !ok {
		return amountStr, 0
	}

	remaining := new(big.Rat).Sub(amount, spent)
	// Output 2 decimals like DB numeric(18,2)
	remainingStr := remaining.FloatString(2)

	// percentUsed = floor(spent/amount*100)
	pct := new(big.Rat).Mul(new(big.Rat).Quo(spent, amount), big.NewRat(100, 1))
	pctInt, _ := pct.Float64()
	if pctInt < 0 {
		pctInt = 0
	}
	return remainingStr, int(pctInt)
}

func derefOrEmpty(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
