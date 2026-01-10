package budget

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain"
	"github.com/sonbn-225/goen-api/internal/apperrors"
)

// WithStats extends Budget with computed statistics.
type WithStats struct {
	domain.Budget
	Spent       string `json:"spent"`
	Remaining   string `json:"remaining"`
	PercentUsed int    `json:"percent_used"`
}

// CreateRequest contains budget create parameters.
type CreateRequest struct {
	Name                  *string `json:"name,omitempty"`
	Period                string  `json:"period"`
	PeriodStart           *string `json:"period_start,omitempty"`
	PeriodEnd             *string `json:"period_end,omitempty"`
	Amount                string  `json:"amount"`
	AlertThresholdPercent *int    `json:"alert_threshold_percent,omitempty"`
	RolloverMode          *string `json:"rollover_mode,omitempty"`
	CategoryID            *string `json:"category_id,omitempty"`
}

// Service handles budget business logic.
type Service struct {
	repo         domain.BudgetRepository
	categoryRepo domain.CategoryRepository
}

// NewService creates a new budget service.
func NewService(repo domain.BudgetRepository, categoryRepo domain.CategoryRepository) *Service {
	return &Service{repo: repo, categoryRepo: categoryRepo}
}

// Create creates a new budget.
func (s *Service) Create(ctx context.Context, userID string, req CreateRequest) (*WithStats, error) {
	period := strings.TrimSpace(req.Period)
	if period != "month" && period != "week" && period != "custom" {
		return nil, apperrors.Validation("period is invalid", nil)
	}

	amount := strings.TrimSpace(req.Amount)
	if amount == "" {
		return nil, apperrors.Validation("amount is required", nil)
	}
	if !isValidDecimal(amount) {
		return nil, apperrors.Validation("amount must be a decimal string", nil)
	}

	categoryID := normalizeOptionalString(req.CategoryID)
	if categoryID == nil {
		return nil, apperrors.Validation("category_id is required", map[string]any{"field": "category_id"})
	}
	cat, err := s.categoryRepo.GetCategory(ctx, userID, *categoryID)
	if err != nil {
		if errors.Is(err, apperrors.ErrCategoryNotFound) {
			return nil, apperrors.Wrap(apperrors.KindNotFound, "category not found", err)
		}
		return nil, err
	}
	if cat == nil || !cat.IsActive {
		return nil, apperrors.Validation("category_id is invalid", map[string]any{"field": "category_id"})
	}
	if cat.IsSystem {
		return nil, apperrors.Validation("category_id is invalid", map[string]any{"field": "category_id"})
	}

	start, end, err := normalizeBudgetPeriod(period, req.PeriodStart, req.PeriodEnd)
	if err != nil {
		return nil, err
	}

	alert := req.AlertThresholdPercent
	if alert != nil {
		if *alert < 0 || *alert > 100 {
			return nil, apperrors.Validation("alert_threshold_percent must be between 0 and 100", nil)
		}
	} else {
		v := 80
		alert = &v
	}

	over := normalizeOptionalString(req.RolloverMode)
	if over != nil {
		v := *over
		if v != "reset" && v != "carry_forward" && v != "accumulate" {
			return nil, apperrors.Validation("rollover_mode is invalid", nil)
		}
	}

	now := time.Now().UTC()
	b := domain.Budget{
		ID:                    uuid.NewString(),
		UserID:                userID,
		Name:                  normalizeOptionalString(req.Name),
		Period:                period,
		PeriodStart:           &start,
		PeriodEnd:             &end,
		Amount:                amount,
		AlertThresholdPercent: alert,
		RolloverMode:          over,
		CategoryID:            categoryID,
		CreatedAt:             now,
		UpdatedAt:             now,
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

// Get retrieves a budget by ID with stats.
func (s *Service) Get(ctx context.Context, userID, budgetID string) (*WithStats, error) {
	b, err := s.repo.GetBudget(ctx, userID, budgetID)
	if err != nil {
		if errors.Is(err, apperrors.ErrBudgetNotFound) {
			return nil, apperrors.Wrap(apperrors.KindNotFound, "budget not found", err)
		}
		return nil, err
	}
	return s.withStats(ctx, userID, *b)
}

// List returns all budgets for a user with stats.
func (s *Service) List(ctx context.Context, userID string) ([]WithStats, error) {
	items, err := s.repo.ListBudgets(ctx, userID)
	if err != nil {
		return nil, err
	}

	out := make([]WithStats, 0, len(items))
	for _, b := range items {
		w, err := s.withStats(ctx, userID, b)
		if err != nil {
			return nil, err
		}
		out = append(out, *w)
	}
	return out, nil
}

func (s *Service) withStats(ctx context.Context, userID string, b domain.Budget) (*WithStats, error) {
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

	return &WithStats{
		Budget:      b,
		Spent:       spent,
		Remaining:   remaining,
		PercentUsed: percent,
	}, nil
}

func normalizeBudgetPeriod(period string, startIn *string, endIn *string) (string, string, error) {
	startStr := strings.TrimSpace(derefOrEmpty(startIn))
	endStr := strings.TrimSpace(derefOrEmpty(endIn))

	if period == "custom" {
		if startStr == "" {
			return "", "", apperrors.Validation("period_start is required", nil)
		}
		if endStr == "" {
			return "", "", apperrors.Validation("period_end is required", nil)
		}
		start, err := time.Parse("2006-01-02", startStr)
		if err != nil {
			return "", "", apperrors.Validation("period_start is invalid", nil)
		}
		end, err := time.Parse("2006-01-02", endStr)
		if err != nil {
			return "", "", apperrors.Validation("period_end is invalid", nil)
		}
		if end.Before(start) {
			return "", "", apperrors.Validation("period_end must be >= period_start", nil)
		}
		return startStr, endStr, nil
	}

	if startStr == "" {
		return "", "", apperrors.Validation("period_start is required", nil)
	}
	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return "", "", apperrors.Validation("period_start is invalid", nil)
	}

	var end time.Time
	switch period {
	case "month":
		firstOfNext := time.Date(start.Year(), start.Month()+1, 1, 0, 0, 0, 0, time.UTC)
		end = firstOfNext.Add(-24 * time.Hour)
	case "week":
		end = start.AddDate(0, 0, 6)
	default:
		return "", "", apperrors.Validation("period is invalid", nil)
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
	remainingStr := remaining.FloatString(2)

	pct := new(big.Rat).Mul(new(big.Rat).Quo(spent, amount), big.NewRat(100, 1))
	pctInt, _ := pct.Float64()
	if pctInt < 0 {
		pctInt = 0
	}
	return remainingStr, int(pctInt)
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

func derefOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func isValidDecimal(s string) bool {
	_, ok := new(big.Rat).SetString(s)
	return ok
}
