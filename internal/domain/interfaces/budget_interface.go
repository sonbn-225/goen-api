package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// BudgetRepository định nghĩa lớp truy cập dữ liệu cho các giới hạn chi tiêu và theo dõi ngân sách.
type BudgetRepository interface {
	// CreateBudget lưu một hạn mức chi tiêu mới cho người dùng/danh mục.
	CreateBudget(ctx context.Context, userID uuid.UUID, b entity.Budget) error
	// GetBudget lấy thông tin một bản ghi ngân sách cụ thể.
	GetBudget(ctx context.Context, userID uuid.UUID, budgetID uuid.UUID) (*entity.Budget, error)
	// ListBudgets trả về tất cả các ngân sách hiện đang hoạt động của người dùng.
	ListBudgets(ctx context.Context, userID uuid.UUID) ([]entity.Budget, error)
	// UpdateBudget chỉnh sửa các ngưỡng ngân sách hoặc kỳ hạn.
	UpdateBudget(ctx context.Context, userID uuid.UUID, b entity.Budget) error
	// DeleteBudget xóa mềm một bản ghi ngân sách.
	DeleteBudget(ctx context.Context, userID uuid.UUID, budgetID uuid.UUID) error
	// ComputeSpent tính toán tổng số tiền đã chi tiêu cho một danh mục trong một khoảng thời gian cụ thể.
	ComputeSpent(ctx context.Context, userID uuid.UUID, categoryID uuid.UUID, startDate string, endDate string) (string, error)
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

