package interfaces

import (
	"context"

	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type ReportRepository interface {
	GetCashflow(ctx context.Context, userID string, months int) ([]entity.CashflowStat, error)
	GetTopExpenses(ctx context.Context, userID string, year int, month int, limit int) ([]entity.CategoryExpenseStat, error)
}

type ReportService interface {
	GetDashboardReport(ctx context.Context, userID string) (*dto.DashboardReportResponse, error)
}
