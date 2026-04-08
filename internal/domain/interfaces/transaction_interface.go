package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// TransactionRepository định nghĩa lớp truy cập dữ liệu cho các giao dịch sổ cái trung tâm.
type TransactionRepository interface {
	// CreateTransactionTx là phiên bản transactional để ghi nhận một giao dịch cùng với các phân loại (line items) và nhãn (tags).
	CreateTransactionTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, txEntity entity.Transaction, lineItems []entity.TransactionLineItem, tagIDs []uuid.UUID) error
	// GetTransactionTx lấy thông tin chi tiết một giao dịch, hỗ trợ cả trong và ngoài transaction.
	GetTransactionTx(ctx context.Context, tx pgx.Tx, userID, transactionID uuid.UUID) (*entity.Transaction, error)
	// BatchPatchTransactions áp dụng cập nhật hàng loạt cho nhiều giao dịch.
	BatchPatchTransactions(ctx context.Context, userID uuid.UUID, transactionIDs []uuid.UUID, patches map[uuid.UUID]entity.TransactionPatch, mode string) ([]uuid.UUID, []uuid.UUID, error)
	// ListTransactions trả về danh sách giao dịch của một người dùng dựa trên các tiêu chí lọc.
	ListTransactions(ctx context.Context, userID uuid.UUID, filter entity.TransactionListFilter) ([]entity.Transaction, *string, int, error)
	// PatchTransaction cập nhật một phần thông tin của giao dịch hiện có.
	PatchTransaction(ctx context.Context, userID, transactionID uuid.UUID, patch entity.TransactionPatch) (*entity.Transaction, error)
	// DeleteTransactionTx là phiên bản transactional để xóa một giao dịch.
	DeleteTransactionTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, transactionID uuid.UUID) error
	// DeleteTransactionsByAccountTx xóa mềm tất cả các giao dịch liên quan đến một tài khoản cụ thể.
	DeleteTransactionsByAccountTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, accountID uuid.UUID) error
	// ListTransactionsByIDs trả về danh sách các giao dịch từ một tập hợp các ID.
	ListTransactionsByIDs(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) ([]entity.Transaction, error)
}

// TransactionService định nghĩa các nghiệp vụ cốt lõi cho việc ghi chép sổ cái tài chính.
type TransactionService interface {
	// List trả về danh sách các giao dịch đã qua xử lý định dạng cho UI.
	List(ctx context.Context, userID uuid.UUID, req dto.ListTransactionsRequest) ([]dto.TransactionResponse, *string, int, error)
	// Get chi tiết một giao dịch cho UI.
	Get(ctx context.Context, userID, transactionID uuid.UUID) (*dto.TransactionResponse, error)
	// Create khởi tạo một giao dịch mới với logic kiểm tra số dư và phân loại.
	Create(ctx context.Context, userID uuid.UUID, req dto.CreateTransactionRequest) (*dto.TransactionResponse, error)
	// Patch cập nhật các thông tin của giao dịch.
	Patch(ctx context.Context, userID, transactionID uuid.UUID, req dto.TransactionPatchRequest) (*dto.TransactionResponse, error)
	// BatchPatch thực hiện cập nhật hàng loạt cho nhiều giao dịch.
	BatchPatch(ctx context.Context, userID uuid.UUID, req dto.BatchPatchRequest) (*dto.BatchPatchResult, error)
	// Delete xóa một giao dịch và xử lý các tác động dây chuyền (như hoàn tác nợ/tiết kiệm).
	Delete(ctx context.Context, userID, transactionID uuid.UUID) error
	// Import xử lý việc ghi nhận hàng loạt giao dịch từ dữ liệu thô (CSV/Excel).
	Import(ctx context.Context, userID uuid.UUID, req []dto.CreateTransactionRequest) ([]dto.TransactionResponse, error)
	// ListForExport chuẩn bị dữ liệu giao dịch cho việc xuất file CSV/Excel.
	ListForExport(ctx context.Context, userID uuid.UUID, filter entity.TransactionListFilter) ([]entity.ExportTransactionRow, error)
}
