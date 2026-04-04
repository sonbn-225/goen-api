package budget

import (
	"context"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/core/logx"
)

type service struct {
	repo         Repository
	categoryRepo CategoryRepository
}

var _ Service = (*service)(nil)

func NewService(repo Repository, categoryRepo CategoryRepository) Service {
	return &service{repo: repo, categoryRepo: categoryRepo}
}

func (s *service) Create(ctx context.Context, userID string, input CreateInput) (*WithStats, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "budget", "operation", "create")
	logger.Info("budget_create_started", "user_id", userID)

	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}

	period := strings.TrimSpace(input.Period)
	if period != "month" && period != "week" && period != "custom" {
		return nil, apperrors.New(apperrors.KindValidation, "period is invalid")
	}

	amount := strings.TrimSpace(input.Amount)
	if amount == "" {
		return nil, apperrors.New(apperrors.KindValidation, "amount is required")
	}
	if !isValidDecimal(amount) {
		return nil, apperrors.New(apperrors.KindValidation, "amount must be a decimal string")
	}

	categoryID := normalizeOptionalString(input.CategoryID)
	if categoryID == nil {
		return nil, apperrors.New(apperrors.KindValidation, "category_id is required")
	}
	if s.categoryRepo == nil {
		return nil, apperrors.New(apperrors.KindInternal, "category repository not configured")
	}

	cat, err := s.categoryRepo.GetByID(ctx, userID, *categoryID)
	if err != nil {
		logger.Error("budget_create_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to load category", err)
	}
	if cat == nil || !cat.IsActive || strings.HasPrefix(cat.ID, "cat_sys") {
		return nil, apperrors.New(apperrors.KindValidation, "category_id is invalid")
	}

	start, end, err := normalizeBudgetPeriod(period, input.PeriodStart, input.PeriodEnd)
	if err != nil {
		return nil, err
	}

	alert := input.AlertThresholdPercent
	if alert != nil {
		if *alert < 0 || *alert > 100 {
			return nil, apperrors.New(apperrors.KindValidation, "alert_threshold_percent must be between 0 and 100")
		}
	} else {
		defaultAlert := 80
		alert = &defaultAlert
	}

	rollover := normalizeOptionalString(input.RolloverMode)
	if rollover != nil {
		v := *rollover
		if v != "reset" && v != "carry_forward" && v != "accumulate" {
			return nil, apperrors.New(apperrors.KindValidation, "rollover_mode is invalid")
		}
	}

	now := time.Now().UTC()
	b := Budget{
		ID:                    uuid.NewString(),
		UserID:                userID,
		Name:                  normalizeOptionalString(input.Name),
		Period:                period,
		PeriodStart:           &start,
		PeriodEnd:             &end,
		Amount:                amount,
		AlertThresholdPercent: alert,
		RolloverMode:          rollover,
		CategoryID:            categoryID,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	if err := s.repo.Create(ctx, userID, b); err != nil {
		logger.Error("budget_create_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to create budget", err)
	}

	created, err := s.repo.GetByID(ctx, userID, b.ID)
	if err != nil {
		logger.Error("budget_create_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to load budget", err)
	}
	if created == nil {
		return nil, apperrors.New(apperrors.KindInternal, "created budget not found")
	}

	logger.Info("budget_create_succeeded", "budget_id", created.ID)
	return s.withStats(ctx, userID, *created)
}

func (s *service) Get(ctx context.Context, userID, budgetID string) (*WithStats, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "budget", "operation", "get")
	logger.Info("budget_get_started", "user_id", userID, "budget_id", budgetID)

	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	if strings.TrimSpace(budgetID) == "" {
		return nil, apperrors.New(apperrors.KindValidation, "budgetId is required")
	}

	b, err := s.repo.GetByID(ctx, userID, budgetID)
	if err != nil {
		logger.Error("budget_get_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to get budget", err)
	}
	if b == nil {
		return nil, apperrors.New(apperrors.KindNotFound, "budget not found")
	}

	return s.withStats(ctx, userID, *b)
}

func (s *service) List(ctx context.Context, userID string) ([]WithStats, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "budget", "operation", "list")
	logger.Info("budget_list_started", "user_id", userID)

	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}

	items, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		logger.Error("budget_list_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to list budgets", err)
	}

	out := make([]WithStats, 0, len(items))
	for _, item := range items {
		w, err := s.withStats(ctx, userID, item)
		if err != nil {
			return nil, err
		}
		out = append(out, *w)
	}

	logger.Info("budget_list_succeeded", "count", len(out))
	return out, nil
}

func (s *service) withStats(ctx context.Context, userID string, b Budget) (*WithStats, error) {
	categoryID := ""
	if b.CategoryID != nil {
		categoryID = strings.TrimSpace(*b.CategoryID)
	}
	start := ""
	if b.PeriodStart != nil {
		start = strings.TrimSpace(*b.PeriodStart)
	}
	end := ""
	if b.PeriodEnd != nil {
		end = strings.TrimSpace(*b.PeriodEnd)
	}

	spent := "0"
	if categoryID != "" && start != "" && end != "" {
		v, err := s.repo.ComputeSpent(ctx, userID, categoryID, start, end)
		if err != nil {
			return nil, apperrors.Wrap(apperrors.KindInternal, "failed to compute budget spent", err)
		}
		spent = v
	}

	remaining, percent := computeRemainingAndPercent(b.Amount, spent)
	return &WithStats{Budget: b, Spent: spent, Remaining: remaining, PercentUsed: percent}, nil
}

func normalizeBudgetPeriod(period string, startIn *string, endIn *string) (string, string, error) {
	startStr := strings.TrimSpace(derefOrEmpty(startIn))
	endStr := strings.TrimSpace(derefOrEmpty(endIn))

	if period == "custom" {
		if startStr == "" {
			return "", "", apperrors.New(apperrors.KindValidation, "period_start is required")
		}
		if endStr == "" {
			return "", "", apperrors.New(apperrors.KindValidation, "period_end is required")
		}
		start, err := time.Parse("2006-01-02", startStr)
		if err != nil {
			return "", "", apperrors.New(apperrors.KindValidation, "period_start is invalid")
		}
		end, err := time.Parse("2006-01-02", endStr)
		if err != nil {
			return "", "", apperrors.New(apperrors.KindValidation, "period_end is invalid")
		}
		if end.Before(start) {
			return "", "", apperrors.New(apperrors.KindValidation, "period_end must be >= period_start")
		}
		return startStr, endStr, nil
	}

	if startStr == "" {
		return "", "", apperrors.New(apperrors.KindValidation, "period_start is required")
	}
	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return "", "", apperrors.New(apperrors.KindValidation, "period_start is invalid")
	}

	var end time.Time
	switch period {
	case "month":
		firstOfNext := time.Date(start.Year(), start.Month()+1, 1, 0, 0, 0, 0, time.UTC)
		end = firstOfNext.Add(-24 * time.Hour)
	case "week":
		end = start.AddDate(0, 0, 6)
	default:
		return "", "", apperrors.New(apperrors.KindValidation, "period is invalid")
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
	pctFloat, _ := pct.Float64()
	if pctFloat < 0 {
		pctFloat = 0
	}

	return remainingStr, int(pctFloat)
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
