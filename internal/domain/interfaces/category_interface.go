package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// CategoryRepository định nghĩa lớp truy cập dữ liệu cho việc phân loại giao dịch.
type CategoryRepository interface {
	// --- Nhóm 1: Truy vấn metadata (Flexible Tx) ---

	// GetCategoryTx lấy thông tin metadata của một danh mục cụ thể.
	GetCategoryTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, categoryID uuid.UUID) (*entity.Category, error)
	// ListCategoriesTx trả về tất cả các danh mục khả dụng cho người dùng.
	ListCategoriesTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]entity.Category, error)
	// GetCategoryByKeyTx lấy thông tin danh mục dựa theo key duy nhất.
	GetCategoryByKeyTx(ctx context.Context, tx pgx.Tx, key string) (*entity.Category, error)
}

// CategoryService định nghĩa nghiệp vụ cho việc tổ chức các giao dịch.
type CategoryService interface {
	// Get trả về thông tin danh mục đã được định dạng.
	Get(ctx context.Context, userID, categoryID uuid.UUID) (*dto.CategoryResponse, error)
	// List trả về các danh mục được lọc theo loại giao dịch (thu nhập/chi phí/chuyển khoản).
	List(ctx context.Context, userID uuid.UUID, txType string) ([]dto.CategoryResponse, error)
}
