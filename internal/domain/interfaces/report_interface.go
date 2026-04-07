package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type ReportRepository interface {
	GetCashflow(ctx context.Context, userID uuid.UUID, months int) ([]entity.CashflowStat, error)
	GetTopExpenses(ctx context.Context, userID uuid.UUID, year int, month int, limit int) ([]entity.CategoryExpenseStat, error)
}

type ReportService interface {
	GetDashboardReport(ctx context.Context, userID uuid.UUID) (*dto.DashboardReportResponse, error)
}
