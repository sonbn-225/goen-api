package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// RotatingSavingsRepository định nghĩa lớp truy cập dữ liệu cho các nhóm tiết kiệm xoay vòng (Hội/Hụi).
type RotatingSavingsRepository interface {
	// --- Nhóm 1: Quản lý Nhóm (Flexible Tx) ---

	// GetRotatingGroupTx lấy thông tin metadata của nhóm (hỗ trợ transaction).
	GetRotatingGroupTx(ctx context.Context, tx pgx.Tx, userID, groupID uuid.UUID) (*entity.RotatingSavingsGroup, error)
	// ListRotatingGroupsTx trả về tất cả các nhóm mà người dùng tham gia hoặc sở hữu.
	ListRotatingGroupsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]entity.RotatingSavingsGroup, error)

	// --- Nhóm 2: Thao tác Nhóm (Transactional) ---

	// CreateRotatingGroupTx ghi nhận một nhóm Hội mới.
	CreateRotatingGroupTx(ctx context.Context, tx pgx.Tx, g entity.RotatingSavingsGroup) error
	// UpdateRotatingGroupTx cập nhật các thiết lập của nhóm.
	UpdateRotatingGroupTx(ctx context.Context, tx pgx.Tx, g entity.RotatingSavingsGroup) error
	// DeleteRotatingGroupTx xóa mềm một nhóm và các nhật ký lịch sử liên quan.
	DeleteRotatingGroupTx(ctx context.Context, tx pgx.Tx, userID, groupID uuid.UUID) error

	// --- Nhóm 3: Quản lý Đóng góp (Flexible Tx) ---

	// ListContributionsTx trả về lịch sử đóng góp của một nhóm.
	ListContributionsTx(ctx context.Context, tx pgx.Tx, groupID uuid.UUID) ([]entity.RotatingSavingsContribution, error)

	// --- Nhóm 4: Thao tác Đóng góp (Transactional) ---

	// CreateContributionTx ghi nhận một khoản đóng góp của người tham gia trong một kỳ.
	CreateContributionTx(ctx context.Context, tx pgx.Tx, c entity.RotatingSavingsContribution) error
	// DeleteContributionTx xóa một bản ghi đóng góp.
	DeleteContributionTx(ctx context.Context, tx pgx.Tx, contributionID uuid.UUID) error
	// DeleteContributionByTransactionTx xóa mềm bản ghi đóng góp dựa trên mã giao dịch.
	DeleteContributionByTransactionTx(ctx context.Context, tx pgx.Tx, transactionID uuid.UUID) error

	// --- Nhóm 5: Nhật ký & Kiểm toán (Flexible Tx) ---

	// ListAuditLogsTx trả về lịch sử vận hành của một nhóm.
	ListAuditLogsTx(ctx context.Context, tx pgx.Tx, groupID uuid.UUID) ([]entity.RotatingSavingsAuditLog, error)

	// --- Nhóm 6: Nhật ký & Kiểm toán (Transactional) ---

	// AddAuditLogTx ghi lại một sự kiện vận hành.
	AddAuditLogTx(ctx context.Context, tx pgx.Tx, log entity.RotatingSavingsAuditLog) error
}

// RotatingSavingsService định nghĩa lớp nghiệp vụ để quản lý các nhóm tiết kiệm chung.
type RotatingSavingsService interface {
	CreateGroup(ctx context.Context, userID uuid.UUID, req dto.CreateRotatingSavingsGroupRequest) (*dto.RotatingSavingsGroupResponse, error)
	GetGroup(ctx context.Context, userID, groupID uuid.UUID) (*dto.RotatingSavingsGroupResponse, error)
	GetGroupDetail(ctx context.Context, userID, groupID uuid.UUID) (*dto.RotatingSavingsGroupDetailResponse, error)
	UpdateGroup(ctx context.Context, userID, groupID uuid.UUID, req dto.UpdateRotatingSavingsGroupRequest) (*dto.RotatingSavingsGroupResponse, error)
	DeleteGroup(ctx context.Context, userID, groupID uuid.UUID) error
	ListGroups(ctx context.Context, userID uuid.UUID) ([]dto.RotatingSavingsGroupSummary, error)
	CreateContribution(ctx context.Context, userID, groupID uuid.UUID, req dto.RotatingSavingsContributionRequest) (*dto.RotatingSavingsContributionResponse, error)
	DeleteContribution(ctx context.Context, userID, groupID, contributionID uuid.UUID) error
	CleanupTransactionLinksTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, transactionID uuid.UUID) error
}
