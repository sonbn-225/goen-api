package report

import (
	"context"
	"time"

	"github.com/sonbn-225/goen-api/internal/domain"
)

type Service struct {
	reportRepo  domain.ReportRepository
	accountRepo domain.AccountRepository
}

func NewService(r domain.ReportRepository, a domain.AccountRepository) *Service {
	return &Service{reportRepo: r, accountRepo: a}
}

func (s *Service) GetDashboardReport(ctx context.Context, userID string) (*domain.DashboardReport, error) {
	balances, err := s.accountRepo.ListAccountBalancesForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	cashflow, err := s.reportRepo.GetCashflow(ctx, userID, 6)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	currentYear := now.Year()
	currentMonth := int(now.Month())

	topExpenses, err := s.reportRepo.GetTopExpenses(ctx, userID, currentYear, currentMonth, 5)
	if err != nil {
		return nil, err
	}

	report := &domain.DashboardReport{
		TotalBalances:    balances,
		Cashflow6Months:  cashflow,
		TopExpensesMonth: topExpenses,
	}

	// Make sure arrays are not null when serialized
	if report.TotalBalances == nil {
		report.TotalBalances = make([]domain.AccountBalance, 0)
	}
	if report.Cashflow6Months == nil {
		report.Cashflow6Months = make([]domain.CashflowStat, 0)
	}
	if report.TopExpensesMonth == nil {
		report.TopExpensesMonth = make([]domain.CategoryExpenseStat, 0)
	}

	return report, nil
}

