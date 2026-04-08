package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// ReportRepository định nghĩa lớp truy cập dữ liệu cho các báo cáo phân tích và thống kê.
type ReportRepository interface {
	// --- Nhóm 1: Truy vấn Báo cáo (Read-only Optimized) ---

	// GetCashflow trả về dữ liệu lịch sử thu nhập và chi phí cho một số tháng nhất định.
	GetCashflow(ctx context.Context, userID uuid.UUID, months int) ([]entity.CashflowStat, error)
	// GetTopExpenses trả về các danh mục chi tiêu cao nhất cho một tháng cụ thể.
	GetTopExpenses(ctx context.Context, userID uuid.UUID, year int, month int, limit int) ([]entity.CategoryExpenseStat, error)
}

// ReportService định nghĩa lớp nghiệp vụ để tạo các bảng điều khiển tài chính và tóm tắt.
type ReportService interface {
	// GetDashboardReport tổng hợp số dư, dòng tiền và các chi tiêu hàng đầu cho màn hình chính của người dùng.
	GetDashboardReport(ctx context.Context, userID uuid.UUID) (*dto.DashboardReportResponse, error)
}
