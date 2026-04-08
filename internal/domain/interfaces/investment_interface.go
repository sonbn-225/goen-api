package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// InvestmentRepository định nghĩa lớp truy cập dữ liệu cho chứng khoán, giao dịch và vị thế nắm giữ.
type InvestmentRepository interface {
	// --- Nhóm 1: Truy vấn đầu tư (Flexible Tx) ---

	// GetTradeTx lấy thông tin chi tiết của một giao dịch lịch sử.
	GetTradeTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, tradeID uuid.UUID) (*entity.Trade, error)
	// ListTradesTx trả về lịch sử hoạt động giao dịch của một tài khoản đầu tư.
	ListTradesTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, accountID uuid.UUID) ([]entity.Trade, error)
	// ListHoldingsTx trả về các vị thế chứng khoán hiện tại của một tài khoản.
	ListHoldingsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, accountID uuid.UUID) ([]entity.Holding, error)
	// GetHoldingTx lấy thông tin vị thế của một mã chứng khoán cụ thể.
	GetHoldingTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, accountID uuid.UUID, securityID uuid.UUID) (*entity.Holding, error)
	// ListShareLotsTx trả về các lô cổ phiếu của một mã chứng khoán.
	ListShareLotsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, accountID uuid.UUID, securityID uuid.UUID) ([]entity.ShareLot, error)
	// ListRealizedLogsByTradeIDTx trả về chi tiết P&L cho một giao dịch cụ thể.
	ListRealizedLogsByTradeIDTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, tradeID uuid.UUID) ([]entity.RealizedTradeLog, error)
	// ListRealizedLogsTx trả về lịch sử P&L của một tài khoản.
	ListRealizedLogsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, accountID uuid.UUID) ([]entity.RealizedTradeLog, error)
	// ListSecurityEventElectionsTx trả về danh sách các lựa chọn đã thực hiện cho các sự kiện doanh nghiệp.
	ListSecurityEventElectionsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, accountID uuid.UUID, status *string) ([]entity.SecurityEventElection, error)

	// --- Nhóm 2: Thao tác ghi (Transactional) ---

	// UpsertSecurityEventElectionTx ghi nhận hoặc cập nhật lựa chọn của người dùng cho các sự kiện doanh nghiệp.
	UpsertSecurityEventElectionTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, e entity.SecurityEventElection) (*entity.SecurityEventElection, error)
	// CreateTradeTx ghi nhận một giao dịch mua/bán chứng khoán.
	CreateTradeTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, t entity.Trade) error
	// DeleteTradeTx xóa mềm một bản ghi giao dịch.
	DeleteTradeTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, tradeID uuid.UUID) error
	// UpsertHoldingTx tạo mới hoặc cập nhật ảnh chụp vị thế nắm giữ.
	UpsertHoldingTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, h entity.Holding) (*entity.Holding, error)
	// CreateShareLotTx ghi nhận một lô cổ phiếu mới được mua.
	CreateShareLotTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, lot entity.ShareLot) error
	// UpdateShareLotQuantityTx cập nhật số lượng còn lại trong một lô.
	UpdateShareLotQuantityTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, lotID uuid.UUID, quantity string) error
	// DeleteShareLotsByTradeIDTx xóa các lô cổ phiếu bắt nguồn từ một giao dịch cụ thể.
	DeleteShareLotsByTradeIDTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, tradeID uuid.UUID) error
	// CreateRealizedTradeLogTx ghi nhận lợi nhuận/thua lỗ (P&L) sau khi bán.
	CreateRealizedTradeLogTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, log entity.RealizedTradeLog) error
	// DeleteRealizedLogsByTradeIDTx xóa các nhật ký P&L của một giao dịch cụ thể.
	DeleteRealizedLogsByTradeIDTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, tradeID uuid.UUID) error
	// DeleteTransactionTx dùng để đảo ngược các biến động tiền mặt liên quan đến giao dịch.
	DeleteTransactionTx(ctx context.Context, tx pgx.Tx, userID, transactionID uuid.UUID) error
}

// InvestmentService định nghĩa lớp nghiệp vụ cho quản lý môi giới và danh mục đầu tư.
type InvestmentService interface {
	// CreateTrade xử lý việc thực hiện giao dịch, phân bổ lô cổ phiếu và tích hợp vào sổ cái.
	CreateTrade(ctx context.Context, userID, accountID uuid.UUID, req dto.CreateTradeRequest) (*dto.TradeResponse, error)
	// UpdateTrade xử lý việc chỉnh sửa các giao dịch trong quá khứ.
	UpdateTrade(ctx context.Context, userID, accountID, tradeID uuid.UUID, req dto.CreateTradeRequest) (*dto.TradeResponse, error)
	// DeleteTrade xóa một giao dịch và hoàn tác các tác động đến danh mục/sổ cái.
	DeleteTrade(ctx context.Context, userID, accountID, tradeID uuid.UUID) error
	// ListTrades trả về lịch sử giao dịch.
	ListTrades(ctx context.Context, userID, accountID uuid.UUID) ([]dto.TradeResponse, error)

	// ListHoldings trả về các vị thế danh mục hiện tại kèm theo lãi/lỗ.
	ListHoldings(ctx context.Context, userID, accountID uuid.UUID) ([]dto.HoldingResponse, error)

	// ListEligibleCorporateActions kiểm tra các sự kiện cổ tức/chia tách có thể nhận.
	ListEligibleCorporateActions(ctx context.Context, userID, accountID uuid.UUID) ([]dto.EligibleAction, error)
	// ClaimCorporateAction thực hiện lựa chọn của người dùng cho cổ tức/chia tách.
	ClaimCorporateAction(ctx context.Context, userID, accountID, eventID uuid.UUID, req dto.ClaimCorporateActionRequest) (*dto.TradeResponse, error)
	// GetRealizedPNLReport tổng hợp lãi/lỗ đã thực hiện cho mục đích báo cáo thuế/hiệu suất.
	GetRealizedPNLReport(ctx context.Context, userID, accountID uuid.UUID) (*dto.RealizedPNLReport, error)
	// BackfillTradePrincipalTransactions là công cụ đồng bộ hóa các giao dịch lịch sử với sổ cái trung tâm.
	BackfillTradePrincipalTransactions(ctx context.Context, userID, accountID uuid.UUID) (*dto.BackfillTradePrincipalResponse, error)
}
