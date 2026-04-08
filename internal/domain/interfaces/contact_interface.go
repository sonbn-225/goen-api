package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// ContactRepository định nghĩa lớp truy cập dữ liệu cho danh bạ cá nhân và những người tham gia cùng.
type ContactRepository interface {
	// --- Nhóm 1: Truy vấn danh bạ (Flexible Tx) ---

	// GetContactTx lấy thông tin một liên hệ cụ thể theo ID.
	GetContactTx(ctx context.Context, tx pgx.Tx, userID, contactID uuid.UUID) (*entity.Contact, error)
	// ListContactsTx trả về tất cả các liên hệ của một người dùng.
	ListContactsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]entity.Contact, error)
	// FindUserByEmailTx cố gắng tìm một người dùng hệ thống theo email để liên kết với liên hệ.
	FindUserByEmailTx(ctx context.Context, tx pgx.Tx, email string) (*entity.User, error)
	// FindUserByPhoneTx cố gắng tìm một người dùng hệ thống theo số điện thoại để liên kết liên hệ.
	FindUserByPhoneTx(ctx context.Context, tx pgx.Tx, phone string) (*entity.User, error)

	// --- Nhóm 2: Thao tác ghi (Transactional) ---

	// CreateContactTx lưu một mục liên hệ mới.
	CreateContactTx(ctx context.Context, tx pgx.Tx, c entity.Contact) error
	// UpdateContactTx chỉnh sửa thông tin metadata của liên hệ.
	UpdateContactTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, c entity.Contact) error
	// DeleteContactTx xóa mềm một bản ghi liên hệ.
	DeleteContactTx(ctx context.Context, tx pgx.Tx, userID, contactID uuid.UUID) error
}

// ContactService định nghĩa nghiệp vụ cho việc quản lý các mối quan hệ cá nhân và người tham gia.
type ContactService interface {
	// Create xử lý việc tạo liên hệ và xác định các liên kết hệ thống tiềm năng.
	Create(ctx context.Context, userID uuid.UUID, req dto.CreateContactRequest) (*dto.ContactResponse, error)
	// Get trả về thông tin liên hệ đã được định dạng.
	Get(ctx context.Context, userID, contactID uuid.UUID) (*dto.ContactResponse, error)
	// List trả về tất cả các liên hệ của người dùng cho giao diện.
	List(ctx context.Context, userID uuid.UUID) ([]dto.ContactResponse, error)
	// Update chỉnh sửa các chi tiết của liên hệ.
	Update(ctx context.Context, userID, contactID uuid.UUID, req dto.UpdateContactRequest) (*dto.ContactResponse, error)
	// Delete xóa một liên hệ.
	Delete(ctx context.Context, userID, contactID uuid.UUID) error
	// GetOrCreateByName là công cụ hỗ trợ cho các luồng chi phí dùng chung để liên kết những người tham gia theo tên.
	GetOrCreateByName(ctx context.Context, userID uuid.UUID, name string) (uuid.UUID, error)
}
