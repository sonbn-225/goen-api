package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// BudgetRepository định nghĩa lớp truy cập dữ liệu cho các giới hạn chi tiêu và theo dõi ngân sách.
type BudgetRepository interface {
	// --- Nhóm 1: Truy vấn Thống kê & Danh sách (Flexible Tx) ---

	// GetBudgetTx lấy thông tin một bản ghi ngân sách cụ thể (phiên bản transactional).
	GetBudgetTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, budgetID uuid.UUID) (*entity.Budget, error)
	// ListBudgetsTx trả về tất cả các ngân sách hiện đang hoạt động của người dùng (phiên bản transactional).
	ListBudgetsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]entity.Budget, error)
	// ComputeSpentTx tính toán tổng số tiền đã chi tiêu cho một danh mục trong một khoảng thời gian cụ thể (phiên bản transactional).
	ComputeSpentTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, categoryID uuid.UUID, startDate string, endDate string) (string, error)

	// --- Nhóm 2: Thao tác ghi (Transactional) ---

	// CreateBudgetTx lưu một hạn mức chi tiêu mới.
	CreateBudgetTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, b entity.Budget) error
	// UpdateBudgetTx chỉnh sửa các ngưỡng ngân sách hoặc kỳ hạn.
	UpdateBudgetTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, b entity.Budget) error
	// DeleteBudgetTx xóa mềm bản ghi ngân sách.
	DeleteBudgetTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, budgetID uuid.UUID) error
}

// BudgetService định nghĩa lớp nghiệp vụ cho hoạch định tài chính và cảnh báo chi tiêu.
type BudgetService interface {
	// Create khởi tạo một ngân sách mới và tính toán các thống kê sử dụng ban đầu.
	Create(ctx context.Context, userID uuid.UUID, req dto.CreateBudgetRequest) (*dto.BudgetWithStatsResponse, error)
	// Get trả về một ngân sách bao gồm tiến độ sử dụng theo thời gian thực so với hạn mức.
	Get(ctx context.Context, userID uuid.UUID, budgetID uuid.UUID) (*dto.BudgetWithStatsResponse, error)
	// List trả về tất cả ngân sách của người dùng kèm theo tiến độ chi tiêu hiện tại.
	List(ctx context.Context, userID uuid.UUID) ([]dto.BudgetWithStatsResponse, error)
	// Update chỉnh sửa các mục tiêu ngân sách và tính toán lại mức độ sử dụng.
	Update(ctx context.Context, userID uuid.UUID, budgetID uuid.UUID, req dto.UpdateBudgetRequest) (*dto.BudgetWithStatsResponse, error)
	// Delete xóa một trình theo dõi ngân sách.
	Delete(ctx context.Context, userID uuid.UUID, budgetID uuid.UUID) error
}
