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
	// CreateAccountWithOwnerTx lưu một tài khoản mới và thiết lập liên kết sở hữu trong một transaction có sẵn.
	CreateAccountWithOwnerTx(ctx context.Context, tx pgx.Tx, account entity.Account, ownerUserID uuid.UUID) error
	
	// ListAccountsForUserTx trả về tất cả các tài khoản mà người dùng có thể truy cập.
	ListAccountsForUserTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]entity.Account, error)
	
	// GetAccountForUserTx lấy thông tin một tài khoản duy nhất, xác thực quyền truy cập của người dùng.
	GetAccountForUserTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, accountID uuid.UUID) (*entity.Account, error)
	
	// PatchAccountTx cập nhật một phần metadata của tài khoản.
	PatchAccountTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, accountID uuid.UUID, patch entity.AccountPatch) (*entity.Account, error)
	
	// DeleteAccountTx xóa mềm một tài khoản.
	DeleteAccountTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, accountID uuid.UUID) error
	
	// HasRelatedTransferTransactionsForAccountTx kiểm tra xem tài khoản có liên quan đến bất kỳ giao dịch chuyển khoản nội bộ nào không.
	HasRelatedTransferTransactionsForAccountTx(ctx context.Context, tx pgx.Tx, accountID uuid.UUID) (bool, error)
	
	// ListAccountBalancesForUserTx tổng hợp số dư hiện tại của tất cả các tài khoản có thể truy cập.
	ListAccountBalancesForUserTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]entity.AccountBalance, error)
	
	// ListAccountSharesTx trả về danh sách tất cả người dùng đã được cấp quyền truy cập vào tài khoản này.
	ListAccountSharesTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, accountID uuid.UUID) ([]entity.AccountShare, error)
	
	// UpsertAccountShareTx cấp mới hoặc cập nhật mức độ quyền truy cập của người dùng khác vào tài khoản.
	UpsertAccountShareTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, accountID uuid.UUID, targetUserID uuid.UUID, permission string) (*entity.AccountShare, error)
	
	// RevokeAccountShareTx gỡ bỏ quyền truy cập của người dùng khác vào tài khoản.
	RevokeAccountShareTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, accountID uuid.UUID, targetUserID uuid.UUID) error
	
	// ListAccountAuditEventsTx trả về lịch sử các thay đổi quản trị đối với tài khoản.
	ListAccountAuditEventsTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, accountID uuid.UUID, limit int) ([]entity.AccountAuditEvent, error)
	
	// RecordAccountAuditEventTx ghi lại một hành động cụ thể cho các mục đích kiểm toán.
	RecordAccountAuditEventTx(ctx context.Context, tx pgx.Tx, event entity.AccountAuditEvent) error
}

// AccountService định nghĩa lớp nghiệp vụ để quản lý các tài khoản tài chính và chia sẻ.
type AccountService interface {
	// List trả về tóm tắt tất cả các tài khoản có thể truy cập để hiển thị trên giao diện.
	List(ctx context.Context, userID uuid.UUID) ([]dto.AccountResponse, error)
	// Get trả về thông tin chi tiết về một tài khoản cụ thể.
	Get(ctx context.Context, userID, accountID uuid.UUID) (*dto.AccountResponse, error)
	// Create khởi tạo một tài khoản mới cho người dùng.
	Create(ctx context.Context, userID uuid.UUID, req dto.CreateAccountRequest) (*dto.AccountResponse, error)
	// Patch cập nhật các thiết lập tài khoản đang hoạt động.
	Patch(ctx context.Context, userID, accountID uuid.UUID, req dto.PatchAccountRequest) (*dto.AccountResponse, error)
	// Delete xóa một tài khoản nếu nó không có các phụ thuộc quan trọng.
	Delete(ctx context.Context, userID, accountID uuid.UUID) error
	// ListShares trả về danh sách các cộng tác viên hiện tại của tài khoản.
	ListShares(ctx context.Context, userID, accountID uuid.UUID) ([]dto.AccountShareResponse, error)
	// UpsertShare mời hoặc cập nhật thông tin cộng tác viên thông qua định danh đăng nhập của họ.
	UpsertShare(ctx context.Context, userID, accountID uuid.UUID, login, permission string) (*dto.AccountShareResponse, error)
	// RevokeShare gỡ bỏ một cộng tác viên khỏi tài khoản.
	RevokeShare(ctx context.Context, userID, accountID, targetUserID uuid.UUID) error
	// ListAuditEvents trả về nhật ký định dạng của các thay đổi tài khoản.
	ListAuditEvents(ctx context.Context, userID, accountID uuid.UUID, limit int) ([]dto.AccountAuditEventResponse, error)
}
