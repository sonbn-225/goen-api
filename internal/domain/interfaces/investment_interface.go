package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// InvestmentRepository định nghĩa lớp truy cập dữ liệu cho holdings, báo cáo và sự kiện đầu tư.
type InvestmentRepository interface {
	// ListHoldingsTx trả về các vị thế chứng khoán hiện tại của một tài khoản.
	ListHoldingsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, accountID uuid.UUID) ([]entity.Holding, error)
	// GetHoldingTx lấy thông tin vị thế của một mã chứng khoán cụ thể.
	GetHoldingTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, accountID uuid.UUID, securityID uuid.UUID) (*entity.Holding, error)
	// ListShareLotsTx trả về các lô cổ phiếu của một mã chứng khoán.
	ListShareLotsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, accountID uuid.UUID, securityID uuid.UUID) ([]entity.ShareLot, error)
	// ListRealizedLogsTx trả về lịch sử P&L của một tài khoản.
	ListRealizedLogsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, accountID uuid.UUID) ([]entity.RealizedTradeLog, error)
	// ListSecurityEventElectionsTx trả về danh sách các lựa chọn đã thực hiện cho các sự kiện doanh nghiệp.
	ListSecurityEventElectionsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, accountID uuid.UUID, status *string) ([]entity.SecurityEventElection, error)
	// UpsertSecurityEventElectionTx ghi nhận hoặc cập nhật lựa chọn của người dùng cho các sự kiện doanh nghiệp.
	UpsertSecurityEventElectionTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, e entity.SecurityEventElection) (*entity.SecurityEventElection, error)
}

// InvestmentService định nghĩa lớp nghiệp vụ cho quản lý môi giới và danh mục đầu tư.
type InvestmentService interface {
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
