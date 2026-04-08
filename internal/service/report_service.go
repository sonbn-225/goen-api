package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/pkg/utils"
)

type ReportService struct {
	reportRepo  interfaces.ReportRepository
	accountRepo interfaces.AccountRepository
}

func NewReportService(r interfaces.ReportRepository, a interfaces.AccountRepository) *ReportService {
	return &ReportService{reportRepo: r, accountRepo: a}
}

func (s *ReportService) GetDashboardReport(ctx context.Context, userID uuid.UUID) (*dto.DashboardReportResponse, error) {
	balances, err := s.accountRepo.ListAccountBalancesForUserTx(ctx, nil, userID)
	if err != nil {
		return nil, err
	}

	cashflow, err := s.reportRepo.GetCashflow(ctx, userID, 6)
	if err != nil {
		return nil, err
	}

	now := utils.Now()
	currentYear := now.Year()
	currentMonth := int(now.Month())

	topExpenses, err := s.reportRepo.GetTopExpenses(ctx, userID, currentYear, currentMonth, 5)
	if err != nil {
		return nil, err
	}

	report := &dto.DashboardReportResponse{
		TotalBalances:    dto.NewAccountBalanceResponses(balances),
		Cashflow6Months:  dto.NewCashflowStatResponses(cashflow),
		TopExpensesMonth: dto.NewCategoryExpenseStatResponses(topExpenses),
	}

	// Make sure arrays are not null when serialized
	if report.TotalBalances == nil {
		report.TotalBalances = make([]dto.AccountBalanceResponse, 0)
	}
	if report.Cashflow6Months == nil {
		report.Cashflow6Months = make([]dto.CashflowStatResponse, 0)
	}
	if report.TopExpensesMonth == nil {
		report.TopExpensesMonth = make([]dto.CategoryExpenseStatResponse, 0)
	}

	return report, nil
}
