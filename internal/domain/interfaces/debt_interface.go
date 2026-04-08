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
	// CreateDebtTx là phiên bản transactional để lưu một khoản nợ mới.
	CreateDebtTx(ctx context.Context, tx pgx.Tx, debt entity.Debt) error
	// GetDebt lấy thông tin một khoản nợ cụ thể.
	GetDebt(ctx context.Context, userID uuid.UUID, debtID uuid.UUID) (*entity.Debt, error)
	// ListDebts trả về toàn bộ danh sách nợ của một người dùng.
	ListDebts(ctx context.Context, userID uuid.UUID) ([]entity.Debt, error)
	// UpdateDebtTx là phiên bản transactional để cập nhật thông tin khoản nợ.
	UpdateDebtTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, debt entity.Debt) error
	// DeleteDebt xóa mềm một bản ghi nợ.
	DeleteDebt(ctx context.Context, userID uuid.UUID, debtID uuid.UUID) error
	// DeleteDebtsByOriginatingTransactionTx xóa toàn bộ nợ bắt nguồn từ một giao dịch cụ thể (sử dụng khi hoàn tác giao dịch).
	DeleteDebtsByOriginatingTransactionTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, transactionID uuid.UUID) error

	// CreatePaymentLinkTx là phiên bản transactional để ghi nhận một lần thanh toán nợ, đồng thời cập nhật số dư gốc/lãi.
	CreatePaymentLinkTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, link entity.DebtPaymentLink, newPrincipal string, newOutstandingPrincipal string, newAccruedInterest string, newStatus entity.DebtStatus, closedAt *time.Time) error
	// ListPaymentLinks trả về lịch sử các lần thanh toán của một khoản nợ.
	ListPaymentLinks(ctx context.Context, userID uuid.UUID, debtID uuid.UUID) ([]entity.DebtPaymentLink, error)
	// ListPaymentLinksByTransaction trả về các khoản nợ được thanh toán bởi một giao dịch cụ thể.
	ListPaymentLinksByTransaction(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID) ([]entity.DebtPaymentLink, error)
	// DeletePaymentLinksByTransactionTx xóa các liên kết thanh toán liên quan đến một giao dịch (dùng khi xóa giao dịch thanh toán).
	DeletePaymentLinksByTransactionTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, transactionID uuid.UUID) error

	// CreateInstallment tạo một lịch thanh toán cho khoản nợ.
	CreateInstallment(ctx context.Context, userID uuid.UUID, inst entity.DebtInstallment) error
	// ListInstallments trả về lịch thanh toán của một khoản nợ.
	ListInstallments(ctx context.Context, userID uuid.UUID, debtID uuid.UUID) ([]entity.DebtInstallment, error)

	// ListPublicParticipants trả về danh sách tên người tham gia cho các hồ sơ công khai.
	ListPublicParticipants(ctx context.Context, userID uuid.UUID) ([]string, error)
	// ListPublicDebtsByParticipant trả về các khoản nợ liên quan đến một người tham gia để xem công khai.
	ListPublicDebtsByParticipant(ctx context.Context, userID uuid.UUID, participantName string) ([]entity.PublicDebt, error)
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
}

