package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// TradeRepository defines persistence operations required by TradeService.
type TradeRepository interface {
	GetTradeTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, tradeID uuid.UUID) (*entity.Trade, error)
	ListTradesTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, accountID uuid.UUID) ([]entity.Trade, error)
	ListShareLotsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, accountID uuid.UUID, securityID uuid.UUID) ([]entity.ShareLot, error)
	ListRealizedLogsByTradeIDTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, tradeID uuid.UUID) ([]entity.RealizedTradeLog, error)

	CreateTradeTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, t entity.Trade) error
	DeleteTradeTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, tradeID uuid.UUID) error
	UpsertHoldingTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, h entity.Holding) (*entity.Holding, error)
	CreateShareLotTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, lot entity.ShareLot) error
	UpdateShareLotQuantityTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, lotID uuid.UUID, quantity string) error
	DeleteShareLotsByTradeIDTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, tradeID uuid.UUID) error
	CreateRealizedTradeLogTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, log entity.RealizedTradeLog) error
	DeleteRealizedLogsByTradeIDTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, tradeID uuid.UUID) error
	DeleteTransactionTx(ctx context.Context, tx pgx.Tx, userID, transactionID uuid.UUID) error
}

// TradeService defines business operations for trade history and execution.
// It is intentionally separated from InvestmentService to keep trade as an
// independent component, similar to transaction module boundaries.
type TradeService interface {
	CreateTrade(ctx context.Context, userID, accountID uuid.UUID, req dto.CreateTradeRequest) (*dto.TradeResponse, error)
	UpdateTrade(ctx context.Context, userID, accountID, tradeID uuid.UUID, req dto.CreateTradeRequest) (*dto.TradeResponse, error)
	DeleteTrade(ctx context.Context, userID, accountID, tradeID uuid.UUID) error
	ListTrades(ctx context.Context, userID, accountID uuid.UUID) ([]dto.TradeResponse, error)
}
