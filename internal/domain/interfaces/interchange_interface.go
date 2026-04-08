package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// InterchangeRepository định nghĩa lớp lưu trữ cho các bản nhập dữ liệu tạm thời (staged imports), các quy tắc ánh xạ và các tác vụ xuất dữ liệu trong tương lai.
type InterchangeRepository interface {
	// --- Nhóm 1: Truy vấn dữ liệu chờ (Flexible Tx) ---

	// ListStagedImportsTx trả về danh sách các bản nhập đang chờ xử lý (hỗ trợ transaction).
	ListStagedImportsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, resourceType string) ([]entity.StagedImport, error)
	// GetStagedImportTx lấy một bản ghi nhập dữ liệu thô duy nhất.
	GetStagedImportTx(ctx context.Context, tx pgx.Tx, userID, id uuid.UUID) (*entity.StagedImport, error)
	// ListImportRulesTx trả về các quy tắc giúp ánh xạ dữ liệu bên ngoài vào các danh mục/nhãn nội bộ.
	ListImportRulesTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, resourceType string) ([]entity.StagedImportRule, error)

	// --- Nhóm 2: Thao tác ghi (Transactional) ---

	// UpsertStagedImportsTx tạo hoặc cập nhật hàng loạt dữ liệu thô.
	UpsertStagedImportsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, items []entity.StagedImportCreate) ([]entity.StagedImport, error)
	// PatchStagedImportTx chỉnh sửa thủ công dữ liệu đã nhập.
	PatchStagedImportTx(ctx context.Context, tx pgx.Tx, userID, id uuid.UUID, patch entity.StagedImportPatch) (*entity.StagedImport, error)
	// DeleteStagedImportTx xóa một bản ghi nhập dữ liệu đang chờ.
	DeleteStagedImportTx(ctx context.Context, tx pgx.Tx, userID, id uuid.UUID) error
	// DeleteAllStagedImportsTx xóa sạch khu vực chờ cho một loại tài nguyên.
	DeleteAllStagedImportsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, resourceType string) (int64, error)

	// UpsertImportRulesTx lưu các quy tắc khớp mẫu.
	UpsertImportRulesTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, rules []entity.StagedImportRuleUpsert) ([]entity.StagedImportRule, error)
	// DeleteImportRuleTx xóa một quy tắc ánh xạ.
	DeleteImportRuleTx(ctx context.Context, tx pgx.Tx, userID, id uuid.UUID) error
}

// InterchangeService định nghĩa nghiệp vụ cho trao đổi dữ liệu chung (Nhập & Xuất).
type InterchangeService interface {
	// Logic Nhập dữ liệu
	// StageImport phân tích dữ liệu bên ngoài thô và đưa vào khu vực chờ (staging area).
	StageImport(ctx context.Context, userID uuid.UUID, resourceType string, source string, items []map[string]any) (int, int, []string, error)
	// ListStaged trả về tóm tắt dữ liệu đang chờ xử lý định dạng cho UI.
	ListStaged(ctx context.Context, userID uuid.UUID, resourceType string) ([]dto.StagedImportResponse, error)
	// PatchStaged cập nhật một bản ghi trong khu vực chờ.
	PatchStaged(ctx context.Context, userID, id uuid.UUID, req dto.PatchStagedImportRequest) (*dto.StagedImportResponse, error)
	// DeleteStaged xóa một bản ghi đang chờ xử lý.
	DeleteStaged(ctx context.Context, userID, id uuid.UUID) error
	// ClearStaged xóa sạch khu vực chờ cho người dùng và loại tài nguyên cụ thể.
	ClearStaged(ctx context.Context, userID uuid.UUID, resourceType string) error

	// UpsertRules quản lý các quy tắc mẫu dùng để xử lý dữ liệu nhập thô.
	UpsertRules(ctx context.Context, userID uuid.UUID, resourceType string, rules []dto.MappingRuleInput) ([]dto.ImportMappingRuleResponse, error)
	// ListRules trả về các mẫu ánh xạ đang hoạt động.
	ListRules(ctx context.Context, userID uuid.UUID, resourceType string) ([]dto.ImportMappingRuleResponse, error)
	// DeleteRule xóa một quy tắc ánh xạ cụ thể.
	DeleteRule(ctx context.Context, userID, id uuid.UUID) error

	// ApplyRulesAndCreate xử lý tất cả các mục đang chờ và cố gắng tạo các thực thể thực tế.
	ApplyRulesAndCreate(ctx context.Context, userID uuid.UUID, resourceType string) (*dto.BatchImportResult, error)
	// CreateManyFromStaged phê duyệt các mục cụ thể từ khu vực chờ thành các bản ghi vĩnh viễn.
	CreateManyFromStaged(ctx context.Context, userID uuid.UUID, resourceType string, ids []uuid.UUID) (*dto.BatchImportResult, error)

	// Logic Xuất dữ liệu
	// ExportToCSV tạo một tệp có thể tải xuống cho một loại tài nguyên mục tiêu.
	ExportToCSV(ctx context.Context, userID uuid.UUID, resourceType string, filter any) ([]byte, string, error)
}
