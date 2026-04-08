package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// TagRepository định nghĩa lớp truy cập dữ liệu cho các nhãn phân loại giao dịch (tags).
type TagRepository interface {
	// --- Nhóm 1: Truy vấn Nhãn (Flexible Tx) ---

	// GetTagTx lấy chi tiết một nhãn cụ thể theo ID.
	GetTagTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, tagID uuid.UUID) (*entity.Tag, error)
	// ListTagsTx trả về tất cả các nhãn do người dùng tạo.
	ListTagsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]entity.Tag, error)

	// --- Nhóm 2: Thao tác ghi (Transactional) ---

	// CreateTagTx lưu một nhãn mới.
	CreateTagTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, tag entity.Tag) error
}

// TagService định nghĩa nghiệp vụ cho việc gắn nhãn linh hoạt các giao dịch.
type TagService interface {
	// Create xử lý quy trình tạo một nhãn mới.
	Create(ctx context.Context, userID uuid.UUID, req dto.CreateTagRequest) (*dto.TagResponse, error)
	// Get trả về thông tin nhãn đã được định dạng.
	Get(ctx context.Context, userID, tagID uuid.UUID) (*dto.TagResponse, error)
	// List trả về tất cả các nhãn của người dùng để hiển thị trên giao diện.
	List(ctx context.Context, userID uuid.UUID) ([]dto.TagResponse, error)
	// GetOrCreateByName là công cụ hỗ trợ cho các luồng nhập dữ liệu để phân giải nhãn theo chuỗi tên.
	GetOrCreateByName(ctx context.Context, userID uuid.UUID, name, langHint string) (uuid.UUID, error)
}
