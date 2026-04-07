package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type InvestmentRepository interface {
	// Investment accounts
	GetInvestmentAccount(ctx context.Context, userID uuid.UUID, investmentAccountID uuid.UUID) (*entity.InvestmentAccount, error)
	ListInvestmentAccounts(ctx context.Context, userID uuid.UUID) ([]entity.InvestmentAccount, error)
	UpdateInvestmentAccountSettings(ctx context.Context, userID uuid.UUID, investmentAccountID uuid.UUID, feeSettings any, taxSettings any) (*entity.InvestmentAccount, error)

	// Securities
	GetSecurity(ctx context.Context, securityID uuid.UUID) (*entity.Security, error)
	ListSecurities(ctx context.Context) ([]entity.Security, error)

	// Read-only market data
	ListSecurityPrices(ctx context.Context, securityID uuid.UUID, from *string, to *string) ([]entity.SecurityPriceDaily, error)
	ListSecurityEvents(ctx context.Context, securityID uuid.UUID, from *string, to *string) ([]entity.SecurityEvent, error)
	GetSecurityEvent(ctx context.Context, securityEventID uuid.UUID) (*entity.SecurityEvent, error)

	// Elections
	UpsertSecurityEventElection(ctx context.Context, userID uuid.UUID, e entity.SecurityEventElection) (*entity.SecurityEventElection, error)
	ListSecurityEventElections(ctx context.Context, userID uuid.UUID, brokerAccountID uuid.UUID, status *string) ([]entity.SecurityEventElection, error)

	// Trades
	CreateTrade(ctx context.Context, userID uuid.UUID, t entity.Trade) error
	GetTrade(ctx context.Context, userID uuid.UUID, tradeID uuid.UUID) (*entity.Trade, error)
	ListTrades(ctx context.Context, userID uuid.UUID, brokerAccountID uuid.UUID) ([]entity.Trade, error)
	DeleteTrade(ctx context.Context, userID uuid.UUID, tradeID uuid.UUID) error

	// Holdings
	ListHoldings(ctx context.Context, userID uuid.UUID, brokerAccountID uuid.UUID) ([]entity.Holding, error)
	GetHolding(ctx context.Context, userID uuid.UUID, brokerAccountID uuid.UUID, securityID uuid.UUID) (*entity.Holding, error)
	UpsertHolding(ctx context.Context, userID uuid.UUID, h entity.Holding) (*entity.Holding, error)

	// Share lots
	ListShareLots(ctx context.Context, userID uuid.UUID, brokerAccountID uuid.UUID, securityID uuid.UUID) ([]entity.ShareLot, error)
	CreateShareLot(ctx context.Context, userID uuid.UUID, lot entity.ShareLot) error
	UpdateShareLotQuantity(ctx context.Context, userID uuid.UUID, lotID uuid.UUID, quantity string) error
	DeleteShareLotsByTradeID(ctx context.Context, userID uuid.UUID, tradeID uuid.UUID) error
	CreateRealizedTradeLog(ctx context.Context, userID uuid.UUID, log entity.RealizedTradeLog) error
	ListRealizedLogsByTradeID(ctx context.Context, userID uuid.UUID, tradeID uuid.UUID) ([]entity.RealizedTradeLog, error)
	DeleteRealizedLogsByTradeID(ctx context.Context, userID uuid.UUID, tradeID uuid.UUID) error
	ListRealizedLogs(ctx context.Context, userID uuid.UUID, brokerAccountID uuid.UUID) ([]entity.RealizedTradeLog, error)
}

type InvestmentService interface {
	GetInvestmentAccount(ctx context.Context, userID, investmentAccountID uuid.UUID) (*dto.InvestmentAccountResponse, error)
	ListInvestmentAccounts(ctx context.Context, userID uuid.UUID) ([]dto.InvestmentAccountResponse, error)
	UpdateInvestmentAccountSettings(ctx context.Context, userID, investmentAccountID uuid.UUID, req dto.PatchInvestmentAccountRequest) (*dto.InvestmentAccountResponse, error)

	GetSecurity(ctx context.Context, securityID uuid.UUID) (*dto.SecurityResponse, error)
	ListSecurities(ctx context.Context) ([]dto.SecurityResponse, error)

	CreateTrade(ctx context.Context, userID, brokerAccountID uuid.UUID, req dto.CreateTradeRequest) (*dto.TradeResponse, error)
	UpdateTrade(ctx context.Context, userID, brokerAccountID, tradeID uuid.UUID, req dto.CreateTradeRequest) (*dto.TradeResponse, error)
	DeleteTrade(ctx context.Context, userID, brokerAccountID, tradeID uuid.UUID) error
	ListTrades(ctx context.Context, userID, brokerAccountID uuid.UUID) ([]dto.TradeResponse, error)

	ListHoldings(ctx context.Context, userID, brokerAccountID uuid.UUID) ([]dto.HoldingResponse, error)
	ListSecurityPrices(ctx context.Context, securityID uuid.UUID, from, to *string) ([]dto.SecurityPriceDailyResponse, error)
	ListSecurityEvents(ctx context.Context, securityID uuid.UUID, from, to *string) ([]dto.SecurityEventResponse, error)

	ListEligibleCorporateActions(ctx context.Context, userID, brokerAccountID uuid.UUID) ([]dto.EligibleAction, error)
	ClaimCorporateAction(ctx context.Context, userID, brokerAccountID, eventID uuid.UUID, req dto.ClaimCorporateActionRequest) (*dto.TradeResponse, error)
	GetRealizedPNLReport(ctx context.Context, userID, brokerAccountID uuid.UUID) (*dto.RealizedPNLReport, error)
	BackfillTradePrincipalTransactions(ctx context.Context, userID, brokerAccountID uuid.UUID) (*dto.BackfillTradePrincipalResponse, error)
}

