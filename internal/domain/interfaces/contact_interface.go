package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// ContactRepository định nghĩa lớp truy cập dữ liệu cho danh bạ cá nhân và những người tham gia cùng.
type ContactRepository interface {
	// CreateContact lưu một mục liên hệ mới.
	CreateContact(ctx context.Context, c entity.Contact) error
	// GetContact lấy thông tin một liên hệ cụ thể theo ID.
	GetContact(ctx context.Context, userID, contactID uuid.UUID) (*entity.Contact, error)
	// ListContacts trả về tất cả các liên hệ của một người dùng.
	ListContacts(ctx context.Context, userID uuid.UUID) ([]entity.Contact, error)
	// UpdateContact chỉnh sửa thông tin metadata của liên hệ (tên, số điện thoại, email, v.v.).
	UpdateContact(ctx context.Context, userID uuid.UUID, c entity.Contact) error
	// DeleteContact xóa mềm một bản ghi liên hệ.
	DeleteContact(ctx context.Context, userID, contactID uuid.UUID) error

	// Tìm kiếm người dùng để liên kết
	// FindUserByEmail cố gắng tìm một người dùng hệ thống theo email để liên kết với liên hệ.
	FindUserByEmail(ctx context.Context, email string) (*entity.User, error)
	// FindUserByPhone cố gắng tìm một người dùng hệ thống theo số điện thoại để liên kết liên hệ.
	FindUserByPhone(ctx context.Context, phone string) (*entity.User, error)
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

