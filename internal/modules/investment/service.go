package investment

import (
	"context"

	"github.com/sonbn-225/goen-api/internal/config"
	"github.com/sonbn-225/goen-api/internal/domain"
	"github.com/sonbn-225/goen-api/internal/storage"
)

// Service handles investment business logic.
type Service struct {
	repo       domain.InvestmentRepository
	accountSvc AccountServiceDep
	txSvc      TransactionServiceDep
	cfg        *config.Config
	redis      *storage.Redis
}

// NewService creates a new investment service.
func NewService(
	repo domain.InvestmentRepository,
	accountSvc AccountServiceDep,
	txSvc TransactionServiceDep,
	cfg *config.Config,
	redis *storage.Redis,
) *Service {
	return &Service{
		repo:       repo,
		accountSvc: accountSvc,
		txSvc:      txSvc,
		cfg:        cfg,
		redis:      redis,
	}
}

// GetInvestmentAccount retrieves an investment account by ID.
func (s *Service) GetInvestmentAccount(ctx context.Context, userID, investmentAccountID string) (*domain.InvestmentAccount, error) {
	return s.repo.GetInvestmentAccount(ctx, userID, investmentAccountID)
}

// ListInvestmentAccounts lists all investment accounts for a user.
func (s *Service) ListInvestmentAccounts(ctx context.Context, userID string) ([]domain.InvestmentAccount, error) {
	return s.repo.ListInvestmentAccounts(ctx, userID)
}

// GetSecurity retrieves a security by ID.
func (s *Service) GetSecurity(ctx context.Context, securityID string) (*domain.Security, error) {
	return s.repo.GetSecurity(ctx, securityID)
}

// ListSecurities lists all securities.
func (s *Service) ListSecurities(ctx context.Context) ([]domain.Security, error) {
	return s.repo.ListSecurities(ctx)
}

// ListTrades lists trades for an investment account.
func (s *Service) ListTrades(ctx context.Context, userID, brokerAccountID string) ([]domain.Trade, error) {
	return s.repo.ListTrades(ctx, userID, brokerAccountID)
}

// ListHoldings lists holdings for an investment account.
func (s *Service) ListHoldings(ctx context.Context, userID, brokerAccountID string) ([]domain.Holding, error) {
	return s.repo.ListHoldings(ctx, userID, brokerAccountID)
}

// ListSecurityPrices lists price history for a security.
func (s *Service) ListSecurityPrices(ctx context.Context, securityID string, from, to *string) ([]domain.SecurityPriceDaily, error) {
	return s.repo.ListSecurityPrices(ctx, securityID, from, to)
}

// ListSecurityEvents lists events for a security.
func (s *Service) ListSecurityEvents(ctx context.Context, securityID string, from, to *string) ([]domain.SecurityEvent, error) {
	return s.repo.ListSecurityEvents(ctx, securityID, from, to)
}
