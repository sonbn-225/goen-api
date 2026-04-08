package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type InvestmentRepository interface {
	// Elections
	UpsertSecurityEventElection(ctx context.Context, userID uuid.UUID, e entity.SecurityEventElection) (*entity.SecurityEventElection, error)
	ListSecurityEventElections(ctx context.Context, userID uuid.UUID, accountID uuid.UUID, status *string) ([]entity.SecurityEventElection, error)

	// Trades
	CreateTrade(ctx context.Context, userID uuid.UUID, t entity.Trade) error
	CreateTradeTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, t entity.Trade) error
	GetTrade(ctx context.Context, userID uuid.UUID, tradeID uuid.UUID) (*entity.Trade, error)
	ListTrades(ctx context.Context, userID uuid.UUID, accountID uuid.UUID) ([]entity.Trade, error)
	DeleteTrade(ctx context.Context, userID uuid.UUID, tradeID uuid.UUID) error

	// Holdings
	ListHoldings(ctx context.Context, userID uuid.UUID, accountID uuid.UUID) ([]entity.Holding, error)
	GetHolding(ctx context.Context, userID uuid.UUID, accountID uuid.UUID, securityID uuid.UUID) (*entity.Holding, error)
	UpsertHolding(ctx context.Context, userID uuid.UUID, h entity.Holding) (*entity.Holding, error)

	// Share lots
	ListShareLots(ctx context.Context, userID uuid.UUID, accountID uuid.UUID, securityID uuid.UUID) ([]entity.ShareLot, error)
	CreateShareLot(ctx context.Context, userID uuid.UUID, lot entity.ShareLot) error
	CreateShareLotTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, lot entity.ShareLot) error
	UpdateShareLotQuantity(ctx context.Context, userID uuid.UUID, lotID uuid.UUID, quantity string) error
	UpdateShareLotQuantityTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, lotID uuid.UUID, quantity string) error
	DeleteShareLotsByTradeID(ctx context.Context, userID uuid.UUID, tradeID uuid.UUID) error
	CreateRealizedTradeLog(ctx context.Context, userID uuid.UUID, log entity.RealizedTradeLog) error
	CreateRealizedTradeLogTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, log entity.RealizedTradeLog) error
	ListRealizedLogsByTradeID(ctx context.Context, userID uuid.UUID, tradeID uuid.UUID) ([]entity.RealizedTradeLog, error)
	DeleteRealizedLogsByTradeID(ctx context.Context, userID uuid.UUID, tradeID uuid.UUID) error
	ListRealizedLogs(ctx context.Context, userID uuid.UUID, accountID uuid.UUID) ([]entity.RealizedTradeLog, error)
	DeleteTransactionTx(ctx context.Context, tx pgx.Tx, userID, transactionID uuid.UUID) error
}

type InvestmentService interface {

	CreateTrade(ctx context.Context, userID, accountID uuid.UUID, req dto.CreateTradeRequest) (*dto.TradeResponse, error)
	UpdateTrade(ctx context.Context, userID, accountID, tradeID uuid.UUID, req dto.CreateTradeRequest) (*dto.TradeResponse, error)
	DeleteTrade(ctx context.Context, userID, accountID, tradeID uuid.UUID) error
	ListTrades(ctx context.Context, userID, accountID uuid.UUID) ([]dto.TradeResponse, error)

	ListHoldings(ctx context.Context, userID, accountID uuid.UUID) ([]dto.HoldingResponse, error)

	ListEligibleCorporateActions(ctx context.Context, userID, accountID uuid.UUID) ([]dto.EligibleAction, error)
	ClaimCorporateAction(ctx context.Context, userID, accountID, eventID uuid.UUID, req dto.ClaimCorporateActionRequest) (*dto.TradeResponse, error)
	GetRealizedPNLReport(ctx context.Context, userID, accountID uuid.UUID) (*dto.RealizedPNLReport, error)
	BackfillTradePrincipalTransactions(ctx context.Context, userID, accountID uuid.UUID) (*dto.BackfillTradePrincipalResponse, error)
}

