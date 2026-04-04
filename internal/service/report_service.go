package service

import (
	"context"
	"time"

	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
)

type ReportService struct {
	reportRepo  interfaces.ReportRepository
	accountRepo interfaces.AccountRepository
}

func NewReportService(r interfaces.ReportRepository, a interfaces.AccountRepository) *ReportService {
	return &ReportService{reportRepo: r, accountRepo: a}
}

func (s *ReportService) GetDashboardReport(ctx context.Context, userID string) (*entity.DashboardReport, error) {
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

	report := &entity.DashboardReport{
		TotalBalances:    balances,
		Cashflow6Months:  cashflow,
		TopExpensesMonth: topExpenses,
	}

	// Make sure arrays are not null when serialized
	if report.TotalBalances == nil {
		report.TotalBalances = make([]entity.AccountBalance, 0)
	}
	if report.Cashflow6Months == nil {
		report.Cashflow6Months = make([]entity.CashflowStat, 0)
	}
	if report.TopExpensesMonth == nil {
		report.TopExpensesMonth = make([]entity.CategoryExpenseStat, 0)
	}

	return report, nil
}
