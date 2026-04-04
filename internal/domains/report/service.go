package report

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/core/logx"
)

type service struct {
	repo Repository
}

var _ Service = (*service)(nil)

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetDashboardReport(ctx context.Context, userID string) (*DashboardReport, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "report", "operation", "get_dashboard_report")
	logger.Info("report_get_dashboard_started", "user_id", userID)

	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}

	balances, err := s.repo.ListAccountBalances(ctx, userID)
	if err != nil {
		logger.Error("report_get_dashboard_failed", "error", err)
		return nil, passThroughOrWrapInternal("failed to list account balances", err)
	}

	cashflow, err := s.repo.GetCashflow(ctx, userID, 6)
	if err != nil {
		logger.Error("report_get_dashboard_failed", "error", err)
		return nil, passThroughOrWrapInternal("failed to read cashflow", err)
	}

	now := time.Now().UTC()
	topExpenses, err := s.repo.GetTopExpenses(ctx, userID, now.Year(), int(now.Month()), 5)
	if err != nil {
		logger.Error("report_get_dashboard_failed", "error", err)
		return nil, passThroughOrWrapInternal("failed to read top expenses", err)
	}

	report := &DashboardReport{
		TotalBalances:    balances,
		Cashflow6Months:  cashflow,
		TopExpensesMonth: topExpenses,
	}

	if report.TotalBalances == nil {
		report.TotalBalances = make([]AccountBalance, 0)
	}
	if report.Cashflow6Months == nil {
		report.Cashflow6Months = make([]CashflowStat, 0)
	}
	if report.TopExpensesMonth == nil {
		report.TopExpensesMonth = make([]CategoryExpenseStat, 0)
	}

	logger.Info("report_get_dashboard_succeeded", "balances_count", len(report.TotalBalances), "cashflow_count", len(report.Cashflow6Months), "top_expenses_count", len(report.TopExpensesMonth))
	return report, nil
}

func passThroughOrWrapInternal(message string, err error) error {
	if err == nil {
		return nil
	}
	var appErr *apperrors.Error
	if errors.As(err, &appErr) {
		return err
	}
	return apperrors.Wrap(apperrors.KindInternal, message, err)
}
