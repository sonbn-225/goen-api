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
	// Quyền bầu chọn (Elections)
	// UpsertSecurityEventElection ghi nhận hoặc cập nhật lựa chọn của người dùng cho các sự kiện doanh nghiệp (ví dụ: nhận cổ tức).
	UpsertSecurityEventElection(ctx context.Context, userID uuid.UUID, e entity.SecurityEventElection) (*entity.SecurityEventElection, error)
	// ListSecurityEventElections trả về danh sách các lựa chọn đã thực hiện cho các sự kiện doanh nghiệp trong một tài khoản cụ thể.
	ListSecurityEventElections(ctx context.Context, userID uuid.UUID, accountID uuid.UUID, status *string) ([]entity.SecurityEventElection, error)

	// Giao dịch (Trades)
	// CreateTradeTx là phiên bản transactional để ghi nhận một giao dịch mua/bán chứng khoán.
	CreateTradeTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, t entity.Trade) error
	// GetTrade lấy thông tin chi tiết của một giao dịch lịch sử.
	GetTrade(ctx context.Context, userID uuid.UUID, tradeID uuid.UUID) (*entity.Trade, error)
	// ListTrades trả về lịch sử hoạt động giao dịch của một tài khoản đầu tư.
	ListTrades(ctx context.Context, userID uuid.UUID, accountID uuid.UUID) ([]entity.Trade, error)
	// DeleteTrade xóa mềm một bản ghi giao dịch.
	DeleteTrade(ctx context.Context, userID uuid.UUID, tradeID uuid.UUID) error

	// Vị thế nắm giữ (Holdings)
	// ListHoldings trả về các vị thế chứng khoán hiện tại của một tài khoản.
	ListHoldings(ctx context.Context, userID uuid.UUID, accountID uuid.UUID) ([]entity.Holding, error)
	// GetHolding lấy thông tin vị thế của một mã chứng khoán cụ thể.
	GetHolding(ctx context.Context, userID uuid.UUID, accountID uuid.UUID, securityID uuid.UUID) (*entity.Holding, error)
	// UpsertHolding tạo mới hoặc cập nhật ảnh chụp vị thế nắm giữ.
	UpsertHolding(ctx context.Context, userID uuid.UUID, h entity.Holding) (*entity.Holding, error)

	// Lô cổ phiếu (Share lots)
	// ListShareLots trả về các lô cổ phiếu của một mã chứng khoán (dùng cho việc theo dõi FIFO/LIFO).
	ListShareLots(ctx context.Context, userID uuid.UUID, accountID uuid.UUID, securityID uuid.UUID) ([]entity.ShareLot, error)
	// CreateShareLotTx là phiên bản transactional để ghi nhận một lô cổ phiếu mới được mua.
	CreateShareLotTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, lot entity.ShareLot) error
	// UpdateShareLotQuantityTx là phiên bản transactional để cập nhật số lượng còn lại trong một lô.
	UpdateShareLotQuantityTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, lotID uuid.UUID, quantity string) error
	// DeleteShareLotsByTradeID xóa các lô cổ phiếu bắt nguồn từ một giao dịch cụ thể.
	DeleteShareLotsByTradeID(ctx context.Context, userID uuid.UUID, tradeID uuid.UUID) error
	// CreateRealizedTradeLogTx là phiên bản transactional để ghi nhận lợi nhuận/thua lỗ (P&L) sau khi bán.
	CreateRealizedTradeLogTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, log entity.RealizedTradeLog) error
	// ListRealizedLogsByTradeID trả về chi tiết P&L cho một giao dịch cụ thể.
	ListRealizedLogsByTradeID(ctx context.Context, userID uuid.UUID, tradeID uuid.UUID) ([]entity.RealizedTradeLog, error)
	// DeleteRealizedLogsByTradeID xóa các nhật ký P&L của một giao dịch cụ thể.
	DeleteRealizedLogsByTradeID(ctx context.Context, userID uuid.UUID, tradeID uuid.UUID) error
	// ListRealizedLogs trả về lịch sử P&L của một tài khoản.
	ListRealizedLogs(ctx context.Context, userID uuid.UUID, accountID uuid.UUID) ([]entity.RealizedTradeLog, error)
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

