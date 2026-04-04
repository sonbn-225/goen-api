package interfaces

import (
	"context"

	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type InvestmentRepository interface {
	// Investment accounts
	CreateInvestmentAccount(ctx context.Context, userID string, ia entity.InvestmentAccount) error
	GetInvestmentAccount(ctx context.Context, userID string, investmentAccountID string) (*entity.InvestmentAccount, error)
	ListInvestmentAccounts(ctx context.Context, userID string) ([]entity.InvestmentAccount, error)
	UpdateInvestmentAccountSettings(ctx context.Context, userID string, investmentAccountID string, feeSettings any, taxSettings any) (*entity.InvestmentAccount, error)

	// Securities
	GetSecurity(ctx context.Context, securityID string) (*entity.Security, error)
	ListSecurities(ctx context.Context) ([]entity.Security, error)

	// Read-only market data
	ListSecurityPrices(ctx context.Context, securityID string, from *string, to *string) ([]entity.SecurityPriceDaily, error)
	ListSecurityEvents(ctx context.Context, securityID string, from *string, to *string) ([]entity.SecurityEvent, error)
	GetSecurityEvent(ctx context.Context, securityEventID string) (*entity.SecurityEvent, error)

	// Elections
	UpsertSecurityEventElection(ctx context.Context, userID string, e entity.SecurityEventElection) (*entity.SecurityEventElection, error)
	ListSecurityEventElections(ctx context.Context, userID string, brokerAccountID string, status *string) ([]entity.SecurityEventElection, error)

	// Trades
	CreateTrade(ctx context.Context, userID string, t entity.Trade) error
	GetTrade(ctx context.Context, userID string, tradeID string) (*entity.Trade, error)
	ListTrades(ctx context.Context, userID string, brokerAccountID string) ([]entity.Trade, error)
	DeleteTrade(ctx context.Context, userID string, tradeID string) error

	// Holdings
	ListHoldings(ctx context.Context, userID string, brokerAccountID string) ([]entity.Holding, error)
	GetHolding(ctx context.Context, userID string, brokerAccountID string, securityID string) (*entity.Holding, error)
	UpsertHolding(ctx context.Context, userID string, h entity.Holding) (*entity.Holding, error)

	// Share lots
	ListShareLots(ctx context.Context, userID string, brokerAccountID string, securityID string) ([]entity.ShareLot, error)
	CreateShareLot(ctx context.Context, userID string, lot entity.ShareLot) error
	UpdateShareLotQuantity(ctx context.Context, userID string, lotID string, quantity string) error
	DeleteShareLotsByTradeID(ctx context.Context, userID string, tradeID string) error
	CreateRealizedTradeLog(ctx context.Context, userID string, log entity.RealizedTradeLog) error
	ListRealizedLogsByTradeID(ctx context.Context, userID string, tradeID string) ([]entity.RealizedTradeLog, error)
	DeleteRealizedLogsByTradeID(ctx context.Context, userID string, tradeID string) error
	ListRealizedLogs(ctx context.Context, userID string, brokerAccountID string) ([]entity.RealizedTradeLog, error)
}

type InvestmentService interface {
	GetInvestmentAccount(ctx context.Context, userID, investmentAccountID string) (*dto.InvestmentAccountResponse, error)
	ListInvestmentAccounts(ctx context.Context, userID string) ([]dto.InvestmentAccountResponse, error)
	UpdateInvestmentAccountSettings(ctx context.Context, userID, investmentAccountID string, req dto.PatchInvestmentAccountRequest) (*dto.InvestmentAccountResponse, error)

	GetSecurity(ctx context.Context, securityID string) (*dto.SecurityResponse, error)
	ListSecurities(ctx context.Context) ([]dto.SecurityResponse, error)

	CreateTrade(ctx context.Context, userID, brokerAccountID string, req dto.CreateTradeRequest) (*dto.TradeResponse, error)
	UpdateTrade(ctx context.Context, userID, brokerAccountID, tradeID string, req dto.CreateTradeRequest) (*dto.TradeResponse, error)
	DeleteTrade(ctx context.Context, userID, brokerAccountID, tradeID string) error
	ListTrades(ctx context.Context, userID, brokerAccountID string) ([]dto.TradeResponse, error)

	ListHoldings(ctx context.Context, userID, brokerAccountID string) ([]dto.HoldingResponse, error)
	ListSecurityPrices(ctx context.Context, securityID string, from, to *string) ([]dto.SecurityPriceDailyResponse, error)
	ListSecurityEvents(ctx context.Context, securityID string, from, to *string) ([]dto.SecurityEventResponse, error)

	ListEligibleCorporateActions(ctx context.Context, userID, brokerAccountID string) ([]dto.EligibleAction, error)
	ClaimCorporateAction(ctx context.Context, userID, brokerAccountID, eventID string, req dto.ClaimCorporateActionRequest) (*dto.TradeResponse, error)
	GetRealizedPNLReport(ctx context.Context, userID, brokerAccountID string) (*dto.RealizedPNLReport, error)
	BackfillTradePrincipalTransactions(ctx context.Context, userID, brokerAccountID string) (*dto.BackfillTradePrincipalResponse, error)
}
