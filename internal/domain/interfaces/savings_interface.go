package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// SavingsRepository định nghĩa lớp truy cập dữ liệu cho các mục tiêu và tài khoản tiết kiệm.
type SavingsRepository interface {
	// --- Nhóm 1: Truy vấn Tiết kiệm (Flexible Tx) ---

	// GetSavingsTx lấy thông tin chi tiết của một mục tiêu tiết kiệm cụ thể.
	GetSavingsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, savingsID uuid.UUID) (*entity.Savings, error)
	// ListSavingsTx trả về toàn bộ các mục tiêu tiết kiệm của một người dùng.
	ListSavingsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]entity.Savings, error)

	// --- Nhóm 2: Thao tác ghi (Transactional) ---

	// UpdateSavingsTx cập nhật các tham số của mục tiêu tiết kiệm.
	UpdateSavingsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, s entity.Savings) error
	// DeleteSavingsTx xóa mềm một mục tiêu tiết kiệm.
	DeleteSavingsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, savingsID uuid.UUID) error
}

// SavingsService định nghĩa lớp nghiệp vụ cho quản lý tiết kiệm cá nhân.
type SavingsService interface {
	// CreateSavings xử lý quy trình thiết lập một mục tiêu tiết kiệm mới.
	CreateSavings(ctx context.Context, userID uuid.UUID, req dto.CreateSavingsRequest) (*dto.SavingsResponse, error)
	// GetSavings trả về thông tin mục tiêu đã được định dạng.
	GetSavings(ctx context.Context, userID, savingsID uuid.UUID) (*dto.SavingsResponse, error)
	// ListSavings trả về danh sách tất cả các mục tiêu đang hoạt động để hiển thị.
	ListSavings(ctx context.Context, userID uuid.UUID) ([]dto.SavingsResponse, error)
	// PatchSavings cập nhật các trường cụ thể của mục tiêu (số tiền định hướng, ngày đến hạn, v.v.).
	PatchSavings(ctx context.Context, userID, savingsID uuid.UUID, req dto.PatchSavingsRequest) (*dto.SavingsResponse, error)
	// DeleteSavings xóa một mục tiêu tiết kiệm.
	DeleteSavings(ctx context.Context, userID, savingsID uuid.UUID) error
}
