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
	// CreateRotatingGroupTx là phiên bản transactional để ghi nhận một nhóm Hội mới.
	CreateRotatingGroupTx(ctx context.Context, tx pgx.Tx, g entity.RotatingSavingsGroup) error
	// GetRotatingGroup lấy thông tin metadata của nhóm.
	GetRotatingGroup(ctx context.Context, userID, groupID uuid.UUID) (*entity.RotatingSavingsGroup, error)
	// ListRotatingGroups trả về tất cả các nhóm mà người dùng tham gia hoặc sở hữu.
	ListRotatingGroups(ctx context.Context, userID uuid.UUID) ([]entity.RotatingSavingsGroup, error)
	// UpdateRotatingGroupTx là phiên bản transactional để cập nhật các thiết lập của nhóm.
	UpdateRotatingGroupTx(ctx context.Context, tx pgx.Tx, g entity.RotatingSavingsGroup) error
	// DeleteRotatingGroup xóa mềm một nhóm và các nhật ký lịch sử liên quan.
	DeleteRotatingGroup(ctx context.Context, userID, groupID uuid.UUID) error

	// CreateContributionTx là phiên bản transactional để ghi nhận một khoản đóng góp của người tham gia trong một kỳ.
	CreateContributionTx(ctx context.Context, tx pgx.Tx, c entity.RotatingSavingsContribution) error
	// GetContributions trả về lịch sử đóng góp của một nhóm.
	GetContributions(ctx context.Context, groupID uuid.UUID) ([]entity.RotatingSavingsContribution, error)
	// DeleteContribution xóa một bản ghi đóng góp.
	DeleteContribution(ctx context.Context, contributionID uuid.UUID) error

	// AddAuditLogTx là phiên bản transactional để ghi lại một sự kiện vận hành (bắt đầu kỳ, chọn người hốt hụi).
	AddAuditLogTx(ctx context.Context, tx pgx.Tx, log entity.RotatingSavingsAuditLog) error
	// GetAuditLogs trả về lịch sử vận hành của một nhóm.
	GetAuditLogs(ctx context.Context, groupID uuid.UUID) ([]entity.RotatingSavingsAuditLog, error)
}

// RotatingSavingsService định nghĩa lớp nghiệp vụ để quản lý các nhóm tiết kiệm chung.
type RotatingSavingsService interface {
	// CreateGroup xử lý việc khởi tạo và thiết lập người tham gia.
	CreateGroup(ctx context.Context, userID uuid.UUID, req dto.CreateRotatingSavingsGroupRequest) (*dto.RotatingSavingsGroupResponse, error)
	// GetGroup trả về thông tin cơ bản của nhóm.
	GetGroup(ctx context.Context, userID, groupID uuid.UUID) (*dto.RotatingSavingsGroupResponse, error)
	// GetGroupDetail trả về trạng thái đầy đủ của nhóm bao gồm người tham gia, các kỳ và nhật ký P&L.
	GetGroupDetail(ctx context.Context, userID, groupID uuid.UUID) (*dto.RotatingSavingsGroupDetailResponse, error)
	// UpdateGroup cập nhật cấu hình nhóm đang hoạt động.
	UpdateGroup(ctx context.Context, userID, groupID uuid.UUID, req dto.UpdateRotatingSavingsGroupRequest) (*dto.RotatingSavingsGroupResponse, error)
	// DeleteGroup xóa một nhóm và hoàn tác các giao dịch sổ cái liên quan.
	DeleteGroup(ctx context.Context, userID, groupID uuid.UUID) error
	// ListGroups trả về thông tin tóm tắt của tất cả các nhóm của người dùng.
	ListGroups(ctx context.Context, userID uuid.UUID) ([]dto.RotatingSavingsGroupSummary, error)

	// CreateContribution xử lý việc đóng đóng góp theo kỳ và tích hợp vào sổ cái.
	CreateContribution(ctx context.Context, userID, groupID uuid.UUID, req dto.RotatingSavingsContributionRequest) (*dto.RotatingSavingsContributionResponse, error)
	// DeleteContribution xóa một khoản đóng góp kỳ và hoàn tác tác động đến số dư.
	DeleteContribution(ctx context.Context, userID, groupID, contributionID uuid.UUID) error
}

