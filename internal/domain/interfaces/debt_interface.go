package interfaces

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// DebtRepository định nghĩa lớp truy cập dữ liệu cho các khoản nợ, khoản cho vay và liên kết thanh toán.
type DebtRepository interface {
	// --- Nhóm 1: Truy vấn Nợ & Thanh toán (Flexible Tx) ---

	// GetDebtTx lấy thông tin một khoản nợ cụ thể.
	GetDebtTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, debtID uuid.UUID) (*entity.Debt, error)
	// ListDebtsTx trả về toàn bộ danh sách nợ của một người dùng.
	ListDebtsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]entity.Debt, error)
	// ListPaymentLinksTx trả về lịch sử các lần thanh toán của một khoản nợ.
	ListPaymentLinksTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, debtID uuid.UUID) ([]entity.DebtPaymentLink, error)
	// ListPaymentLinksByTransactionTx trả về các khoản nợ được thanh toán bởi một giao dịch cụ thể.
	ListPaymentLinksByTransactionTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, transactionID uuid.UUID) ([]entity.DebtPaymentLink, error)
	// ListDebtsByOriginatingTransactionTx trả về các khoản nợ bắt nguồn từ một giao dịch cụ thể.
	ListDebtsByOriginatingTransactionTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, transactionID uuid.UUID) ([]entity.Debt, error)
	// ListInstallmentsTx trả về lịch thanh toán của một khoản nợ.
	ListInstallmentsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, debtID uuid.UUID) ([]entity.DebtInstallment, error)
	// ListPublicParticipantsTx trả về danh sách tên người tham gia cho các hồ sơ công khai.
	ListPublicParticipantsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]string, error)
	// ListPublicDebtsByParticipantTx trả về các khoản nợ liên quan đến một người tham gia để xem công khai.
	ListPublicDebtsByParticipantTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, participantName string) ([]entity.PublicDebt, error)

	// --- Nhóm 2: Thao tác ghi (Transactional) ---

	// CreateDebtTx lưu một khoản nợ mới.
	CreateDebtTx(ctx context.Context, tx pgx.Tx, debt entity.Debt) error
	// UpdateDebtTx cập nhật thông tin khoản nợ.
	UpdateDebtTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, debt entity.Debt) error
	// DeleteDebtTx xóa mềm một bản ghi nợ.
	DeleteDebtTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, debtID uuid.UUID) error
	// DeleteDebtsByOriginatingTransactionTx xóa toàn bộ nợ bắt nguồn từ một giao dịch cụ thể.
	DeleteDebtsByOriginatingTransactionTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, transactionID uuid.UUID) error

	// CreatePaymentLinkTx ghi nhận một lần thanh toán nợ, đồng thời cập nhật số dư gốc/lãi.
	CreatePaymentLinkTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, link entity.DebtPaymentLink, newPrincipal string, newOutstandingPrincipal string, newAccruedInterest string, newStatus entity.DebtStatus, closedAt *time.Time) error
	// DeletePaymentLinksByTransactionTx xóa các liên kết thanh toán liên quan đến một giao dịch.
	DeletePaymentLinksByTransactionTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, transactionID uuid.UUID) error
	// CreateInstallmentTx tạo một lịch thanh toán cho khoản nợ.
	CreateInstallmentTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, inst entity.DebtInstallment) error
}

// DebtService định nghĩa các nghiệp vụ quản lý nợ, cho vay và chi phí dùng chung.
type DebtService interface {
	// Create ghi nhận một khoản nợ mới.
	Create(ctx context.Context, userID uuid.UUID, req dto.CreateDebtRequest) (*dto.DebtResponse, error)
	// CreateTx thực hiện tạo khoản nợ một cách nguyên tử.
	CreateTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, req dto.CreateDebtRequest) (*dto.DebtResponse, error)
	// Get lấy chi tiết một khoản nợ.
	Get(ctx context.Context, userID uuid.UUID, debtID uuid.UUID) (*dto.DebtResponse, error)
	// List trả về danh sách nợ kèm theo thông tin tóm tắt cho giao diện.
	List(ctx context.Context, userID uuid.UUID) ([]dto.DebtResponse, error)
	// Update handles modifications to debt metadata.
	Update(ctx context.Context, userID uuid.UUID, debtID uuid.UUID, req dto.UpdateDebtRequest) (*dto.DebtResponse, error)
	// Delete removes a debt record.
	Delete(ctx context.Context, userID uuid.UUID, debtID uuid.UUID) error

	// AddPayment links an existing transaction as a payment towards a debt.
	AddPayment(ctx context.Context, userID uuid.UUID, debtID uuid.UUID, req dto.DebtPaymentRequest) (*dto.DebtResponse, error)
	// Repay creates a new repayment transaction and links it to the debt.
	Repay(ctx context.Context, userID uuid.UUID, debtID uuid.UUID, req dto.DebtRepayRequest) (*dto.DebtResponse, error)
	// ListPayments returns the payment history for a debt.
	ListPayments(ctx context.Context, userID uuid.UUID, debtID uuid.UUID) ([]dto.DebtPaymentLinkResponse, error)
	// CleanupTransactionLinksTx reverts debt states when an originating transaction is deleted.
	CleanupTransactionLinksTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, transactionID uuid.UUID) error
	// ListPaymentLinksByTransaction returns the payment history for a specific transaction.
	ListPaymentLinksByTransaction(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID) ([]dto.DebtPaymentLinkResponse, error)
	// ListDebtsByOriginatingTransaction returns the debts created by a specific transaction (e.g., group participants).
	ListDebtsByOriginatingTransaction(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID) ([]dto.DebtResponse, error)
}
