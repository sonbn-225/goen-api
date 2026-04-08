package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// AccountRepository định nghĩa lớp truy cập dữ liệu cho các tài khoản ngân hàng, ví và quyền truy cập chung.
type AccountRepository interface {
	// --- Nhóm 1: Quản lý chi tiết tài khoản (Flexible Tx) ---

	// ListAccountsForUserTx trả về tất cả các tài khoản mà người dùng có thể truy cập.
	ListAccountsForUserTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]entity.Account, error)
	// GetAccountForUserTx lấy thông tin một tài khoản duy nhất (hỗ trợ transaction).
	GetAccountForUserTx(ctx context.Context, tx pgx.Tx, userID, accountID uuid.UUID) (*entity.Account, error)
	// HasRelatedTransferTransactionsForAccountTx kiểm tra xem tài khoản có liên quan đến các giao dịch chuyển khoản không.
	HasRelatedTransferTransactionsForAccountTx(ctx context.Context, tx pgx.Tx, accountID uuid.UUID) (bool, error)

	// --- Nhóm 2: Thao tác ghi (Transactional) ---

	// CreateAccountWithOwnerTx lưu một tài khoản mới và thiết lập liên kết sở hữu.
	CreateAccountWithOwnerTx(ctx context.Context, tx pgx.Tx, account entity.Account, ownerUserID uuid.UUID) error
	// PatchAccountTx cập nhật một phần metadata của tài khoản.
	PatchAccountTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, accountID uuid.UUID, patch entity.AccountPatch) (*entity.Account, error)
	// DeleteAccountTx xóa mềm một tài khoản.
	DeleteAccountTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, accountID uuid.UUID) error

	// --- Nhóm 3: Chia sẻ & Quyền truy cập ---

	// ListAccountSharesTx trả về danh sách cộng tác viên của tài khoản.
	ListAccountSharesTx(ctx context.Context, tx pgx.Tx, actorUserID, accountID uuid.UUID) ([]entity.AccountShare, error)
	// UpsertAccountShareTx cấp mới hoặc cập nhật quyền truy cập.
	UpsertAccountShareTx(ctx context.Context, tx pgx.Tx, actorUserID, accountID, targetUserID uuid.UUID, permission string) (*entity.AccountShare, error)
	// RevokeAccountShareTx gỡ bỏ quyền truy cập.
	RevokeAccountShareTx(ctx context.Context, tx pgx.Tx, actorUserID, accountID, targetUserID uuid.UUID) error

	// --- Nhóm 4: Phân tích & Kiểm toán ---

	// ListAccountBalancesForUserTx tổng hợp số dư hiện tại của tất cả tài khoản.
	ListAccountBalancesForUserTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]entity.AccountBalance, error)
	// ListAccountAuditEventsTx trả về lịch sử thay đổi quản trị.
	ListAccountAuditEventsTx(ctx context.Context, tx pgx.Tx, actorUserID, accountID uuid.UUID, limit int) ([]entity.AccountAuditEvent, error)
	// RecordAccountAuditEventTx ghi lại sự kiện kiểm toán.
	RecordAccountAuditEventTx(ctx context.Context, tx pgx.Tx, event entity.AccountAuditEvent) error
}

// AccountService định nghĩa lớp nghiệp vụ để quản lý các tài khoản tài chính và chia sẻ.
type AccountService interface {
	List(ctx context.Context, userID uuid.UUID) ([]dto.AccountResponse, error)
	Get(ctx context.Context, userID, accountID uuid.UUID) (*dto.AccountResponse, error)
	Create(ctx context.Context, userID uuid.UUID, req dto.CreateAccountRequest) (*dto.AccountResponse, error)
	Patch(ctx context.Context, userID, accountID uuid.UUID, req dto.PatchAccountRequest) (*dto.AccountResponse, error)
	Delete(ctx context.Context, userID, accountID uuid.UUID) error
	ListShares(ctx context.Context, userID, accountID uuid.UUID) ([]dto.AccountShareResponse, error)
	UpsertShare(ctx context.Context, userID, accountID uuid.UUID, login, permission string) (*dto.AccountShareResponse, error)
	RevokeShare(ctx context.Context, userID, accountID, targetUserID uuid.UUID) error
	ListAuditEvents(ctx context.Context, userID, accountID uuid.UUID, limit int) ([]dto.AccountAuditEventResponse, error)
}
