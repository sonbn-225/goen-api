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
	// --- Nhóm 1: Truy vấn danh sách & Chi tiết (Flexible Tx) ---

	// ListTransactionsTx trả về danh sách giao dịch dựa trên các tiêu chí lọc.
	ListTransactionsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, filter entity.TransactionListFilter) ([]entity.Transaction, *string, int, error)
	// GetTransactionTx lấy thông tin chi tiết một giao dịch (hỗ trợ transaction).
	GetTransactionTx(ctx context.Context, tx pgx.Tx, userID, transactionID uuid.UUID) (*entity.Transaction, error)
	// ListTransactionsByIDsTx trả về danh sách các giao dịch từ một tập hợp các ID.
	ListTransactionsByIDsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, ids []uuid.UUID) ([]entity.Transaction, error)

	// --- Nhóm 2: Thao tác ghi (Transactional) ---

	// CreateTransactionTx ghi nhận một giao dịch cùng với các phân loại và nhãn.
	CreateTransactionTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, txEntity entity.Transaction, lineItems []entity.TransactionLineItem, tagIDs []uuid.UUID) error
	// PatchTransactionTx cập nhật một phần thông tin của giao dịch hiện có.
	PatchTransactionTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, transactionID uuid.UUID, patch entity.TransactionPatch) (*entity.Transaction, error)
	// DeleteTransactionTx xóa mềm một giao dịch.
	DeleteTransactionTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, transactionID uuid.UUID) error
	// DeleteTransactionsByAccountTx xóa mềm tất cả các giao dịch liên quan đến một tài khoản.
	DeleteTransactionsByAccountTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, accountID uuid.UUID) error

	// --- Nhóm 3: Thao tác hàng loạt (Transactional) ---

	// BatchPatchTransactionsTx áp dụng cập nhật hàng loạt cho nhiều giao dịch.
	BatchPatchTransactionsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, transactionIDs []uuid.UUID, patches map[uuid.UUID]entity.TransactionPatch, mode string) ([]uuid.UUID, []uuid.UUID, error)
}

// TransactionService định nghĩa các nghiệp vụ cốt lõi cho việc ghi chép sổ cái tài chính.
type TransactionService interface {
	List(ctx context.Context, userID uuid.UUID, req dto.ListTransactionsRequest) ([]dto.TransactionResponse, *string, int, error)
	Get(ctx context.Context, userID, transactionID uuid.UUID) (*dto.TransactionResponse, error)
	Create(ctx context.Context, userID uuid.UUID, req dto.CreateTransactionRequest) (*dto.TransactionResponse, error)
	Patch(ctx context.Context, userID, transactionID uuid.UUID, req dto.TransactionPatchRequest) (*dto.TransactionResponse, error)
	BatchPatch(ctx context.Context, userID uuid.UUID, req dto.BatchPatchRequest) (*dto.BatchPatchResult, error)
	Delete(ctx context.Context, userID, transactionID uuid.UUID) error
	Import(ctx context.Context, userID uuid.UUID, req []dto.CreateTransactionRequest) ([]dto.TransactionResponse, error)
	ListForExport(ctx context.Context, userID uuid.UUID, filter entity.TransactionListFilter) ([]entity.ExportTransactionRow, error)
}
