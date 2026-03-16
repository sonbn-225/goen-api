package investment

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/sonbn-225/goen-api/internal/domain"
	"github.com/sonbn-225/goen-api/internal/modules/transaction"
)

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) CreateInvestmentAccount(ctx context.Context, userID string, ia domain.InvestmentAccount) error {
	args := m.Called(ctx, userID, ia)
	return args.Error(0)
}

func (m *MockRepo) GetInvestmentAccount(ctx context.Context, userID, id string) (*domain.InvestmentAccount, error) {
	args := m.Called(ctx, userID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.InvestmentAccount), args.Error(1)
}

func (m *MockRepo) ListInvestmentAccounts(ctx context.Context, userID string) ([]domain.InvestmentAccount, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.InvestmentAccount), args.Error(1)
}

func (m *MockRepo) GetSecurity(ctx context.Context, id string) (*domain.Security, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Security), args.Error(1)
}

func (m *MockRepo) ListSecurities(ctx context.Context) ([]domain.Security, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Security), args.Error(1)
}

func (m *MockRepo) ListTrades(ctx context.Context, userID, brokerAccountID string) ([]domain.Trade, error) {
	args := m.Called(ctx, userID, brokerAccountID)
	return args.Get(0).([]domain.Trade), args.Error(1)
}

func (m *MockRepo) GetTrade(ctx context.Context, userID, tradeID string) (*domain.Trade, error) {
	args := m.Called(ctx, userID, tradeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Trade), args.Error(1)
}

func (m *MockRepo) CreateTrade(ctx context.Context, userID string, tr domain.Trade) error {
	args := m.Called(ctx, userID, tr)
	return args.Error(0)
}

func (m *MockRepo) DeleteTrade(ctx context.Context, userID, tradeID string) error {
	args := m.Called(ctx, userID, tradeID)
	return args.Error(0)
}

func (m *MockRepo) ListHoldings(ctx context.Context, userID, brokerAccountID string) ([]domain.Holding, error) {
	args := m.Called(ctx, userID, brokerAccountID)
	return args.Get(0).([]domain.Holding), args.Error(1)
}

func (m *MockRepo) GetHolding(ctx context.Context, userID, brokerAccountID, securityID string) (*domain.Holding, error) {
	args := m.Called(ctx, userID, brokerAccountID, securityID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Holding), args.Error(1)
}

func (m *MockRepo) UpsertHolding(ctx context.Context, userID string, h domain.Holding) (*domain.Holding, error) {
	args := m.Called(ctx, userID, h)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Holding), args.Error(1)
}

func (m *MockRepo) ListSecurityPrices(ctx context.Context, securityID string, from, to *string) ([]domain.SecurityPriceDaily, error) {
	args := m.Called(ctx, securityID, from, to)
	return args.Get(0).([]domain.SecurityPriceDaily), args.Error(1)
}

func (m *MockRepo) ListSecurityEvents(ctx context.Context, securityID string, from, to *string) ([]domain.SecurityEvent, error) {
	args := m.Called(ctx, securityID, from, to)
	return args.Get(0).([]domain.SecurityEvent), args.Error(1)
}

func (m *MockRepo) GetSecurityEvent(ctx context.Context, id string) (*domain.SecurityEvent, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SecurityEvent), args.Error(1)
}

func (m *MockRepo) ListSecurityEventElections(ctx context.Context, userID, brokerAccountID string, status *string) ([]domain.SecurityEventElection, error) {
	args := m.Called(ctx, userID, brokerAccountID, status)
	return args.Get(0).([]domain.SecurityEventElection), args.Error(1)
}

func (m *MockRepo) UpsertSecurityEventElection(ctx context.Context, userID string, el domain.SecurityEventElection) (*domain.SecurityEventElection, error) {
	args := m.Called(ctx, userID, el)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SecurityEventElection), args.Error(1)
}

func (m *MockRepo) ListShareLots(ctx context.Context, userID, brokerAccountID, securityID string) ([]domain.ShareLot, error) {
	args := m.Called(ctx, userID, brokerAccountID, securityID)
	return args.Get(0).([]domain.ShareLot), args.Error(1)
}

func (m *MockRepo) CreateShareLot(ctx context.Context, userID string, lot domain.ShareLot) error {
	args := m.Called(ctx, userID, lot)
	return args.Error(0)
}

func (m *MockRepo) UpdateShareLotQuantity(ctx context.Context, userID, lotID, quantity string) error {
	args := m.Called(ctx, userID, lotID, quantity)
	return args.Error(0)
}

func (m *MockRepo) DeleteShareLotsByTradeID(ctx context.Context, userID, tradeID string) error {
	args := m.Called(ctx, userID, tradeID)
	return args.Error(0)
}

func (m *MockRepo) ListRealizedLogsByTradeID(ctx context.Context, userID, tradeID string) ([]domain.RealizedTradeLog, error) {
	args := m.Called(ctx, userID, tradeID)
	return args.Get(0).([]domain.RealizedTradeLog), args.Error(1)
}

func (m *MockRepo) CreateRealizedTradeLog(ctx context.Context, userID string, log domain.RealizedTradeLog) error {
	args := m.Called(ctx, userID, log)
	return args.Error(0)
}

func (m *MockRepo) DeleteRealizedLogsByTradeID(ctx context.Context, userID, tradeID string) error {
	args := m.Called(ctx, userID, tradeID)
	return args.Error(0)
}

func (m *MockRepo) ListRealizedLogs(ctx context.Context, userID string, brokerAccountID string) ([]domain.RealizedTradeLog, error) {
	args := m.Called(ctx, userID, brokerAccountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.RealizedTradeLog), args.Error(1)
}

func (m *MockRepo) ListDividends(ctx context.Context, userID string, brokerAccountID string) ([]domain.Transaction, error) {
	args := m.Called(ctx, userID, brokerAccountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Transaction), args.Error(1)
}

func (m *MockRepo) UpdateInvestmentAccountSettings(ctx context.Context, userID, id string, feeSettings, taxSettings any) (*domain.InvestmentAccount, error) {
	args := m.Called(ctx, userID, id, feeSettings, taxSettings)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.InvestmentAccount), args.Error(1)
}

func (m *MockRepo) ListEligibleSecuritiesForEvents(ctx context.Context, userID, brokerAccountID string) ([]domain.Security, error) {
	args := m.Called(ctx, userID, brokerAccountID)
	return args.Get(0).([]domain.Security), args.Error(1)
}

type MockTxService struct {
	mock.Mock
}

func (m *MockTxService) Create(ctx context.Context, userID string, req transaction.CreateRequest) (*domain.Transaction, error) {
	args := m.Called(ctx, userID, req)
	return args.Get(0).(*domain.Transaction), args.Error(1)
}

func (m *MockTxService) Delete(ctx context.Context, userID, transactionID string) error {
	args := m.Called(ctx, userID, transactionID)
	return args.Error(0)
}

type MockAccountService struct {
	mock.Mock
}

func (m *MockAccountService) GetAccountByID(ctx context.Context, userID, accountID string) (*domain.Account, error) {
	args := m.Called(ctx, userID, accountID)
	return args.Get(0).(*domain.Account), args.Error(1)
}

func TestClaimCorporateAction_Residuals(t *testing.T) {
	repo := new(MockRepo)
	txSvc := new(MockTxService)
	accSvc := new(MockAccountService)
	svc := NewService(repo, accSvc, txSvc, nil, nil)

	ctx := context.Background()
	userID := "user1"
	bid := "account1"
	evID := "event1"
	secID := "VNM"

	// Mocking data
	ev := &domain.SecurityEvent{
		ID:               evID,
		SecurityID:       secID,
		EventType:        "bonus_issue",
		RatioNumerator:   ptr("1"),
		RatioDenominator: ptr("10"),
		EffectiveDate:    ptr("2026-03-16"),
	}

	h := &domain.Holding{
		SecurityID: secID,
		Quantity:   "123", // 123 * 1/10 = 12.3 entitlement
	}

	repo.On("GetSecurityEvent", ctx, evID).Return(ev, nil)
	repo.On("GetSecurity", ctx, secID).Return(&domain.Security{ID: secID, Symbol: secID}, nil)
	repo.On("GetHolding", ctx, userID, bid, secID).Return(h, nil).Once() // First call for entitlement
	repo.On("ListSecurityEventElections", ctx, userID, bid, mock.Anything).Return([]domain.SecurityEventElection{}, nil)
	repo.On("UpsertSecurityEventElection", ctx, userID, mock.Anything).Return(&domain.SecurityEventElection{}, nil)
	
	// Expectations for CreateTrade (service method) -> repo.CreateTrade
	repo.On("CreateTrade", ctx, userID, mock.MatchedBy(func(tr domain.Trade) bool {
		return tr.Quantity == "12.00000000" && tr.Price == "0"
	})).Return(nil)

	// Expectations for CreateTrade (service method) -> repo.CreateShareLot
	repo.On("CreateShareLot", ctx, userID, mock.Anything).Return(nil)
	
	// Expectations for upsertHoldingFromLots
	repo.On("ListShareLots", ctx, userID, bid, secID).Return([]domain.ShareLot{
		{
			Quantity:     "12.00000000",
			CostBasisPer: "0",
			Status:       "active",
		},
	}, nil)
	repo.On("GetHolding", ctx, userID, bid, secID).Return(h, nil).Once() // Second call inside upsertHoldingFromLots
	repo.On("UpsertHolding", ctx, userID, mock.Anything).Return(&domain.Holding{}, nil)

	ia := &domain.InvestmentAccount{ID: bid, AccountID: "cash_acc"}
	repo.On("GetInvestmentAccount", ctx, userID, bid).Return(ia, nil)

	// Expect Income transaction for 0.3 * 10000 = 3000 VND residual
	txSvc.On("Create", ctx, userID, mock.MatchedBy(func(req transaction.CreateRequest) bool {
		// We expect "3000.00" because formatRatDecimalScale(big.NewRat(3000, 1), 2) is "3000.00"
		return req.Type == "income" && req.Amount == "3000.00"
	})).Return(&domain.Transaction{}, nil)

	_, err := svc.ClaimCorporateAction(ctx, userID, bid, evID, ClaimCorporateActionRequest{})
	assert.NoError(t, err)

	repo.AssertExpectations(t)
	txSvc.AssertExpectations(t)
}

func TestClaimCorporateAction_Split_Residuals(t *testing.T) {
	repo := new(MockRepo)
	txSvc := new(MockTxService)
	accSvc := new(MockAccountService)
	svc := NewService(repo, accSvc, txSvc, nil, nil)

	ctx := context.Background()
	userID := "user1"
	bid := "account1"
	evID := "event1"
	secID := "VNM"

	// Mocking data: 3:2 split
	ev := &domain.SecurityEvent{
		ID:               evID,
		SecurityID:       secID,
		EventType:        "split",
		RatioNumerator:   ptr("3"),
		RatioDenominator: ptr("2"),
		EffectiveDate:    ptr("2026-03-16"),
	}

	h := &domain.Holding{
		SecurityID: secID,
		Quantity:   "101", // 101 * 3/2 = 151.5 entitlement
	}

	repo.On("GetSecurityEvent", ctx, evID).Return(ev, nil)
	repo.On("GetSecurity", ctx, secID).Return(&domain.Security{ID: secID, Symbol: secID}, nil)
	repo.On("GetHolding", ctx, userID, bid, secID).Return(h, nil).Once() // First call for entitlement
	repo.On("ListSecurityEventElections", ctx, userID, bid, mock.Anything).Return([]domain.SecurityEventElection{}, nil)
	repo.On("UpsertSecurityEventElection", ctx, userID, mock.Anything).Return(&domain.SecurityEventElection{}, nil)
	
	// Expect Trade for 151 shares
	repo.On("CreateTrade", ctx, userID, mock.MatchedBy(func(tr domain.Trade) bool {
		return tr.Quantity == "151.00000000" && tr.Price == "0"
	})).Return(nil)

	repo.On("CreateShareLot", ctx, userID, mock.Anything).Return(nil)
	repo.On("ListShareLots", ctx, userID, bid, secID).Return([]domain.ShareLot{{Quantity: "151", Status: "active"}}, nil)
	repo.On("GetHolding", ctx, userID, bid, secID).Return(h, nil).Once()
	repo.On("UpsertHolding", ctx, userID, mock.Anything).Return(&domain.Holding{}, nil)

	ia := &domain.InvestmentAccount{ID: bid, AccountID: "cash_acc"}
	repo.On("GetInvestmentAccount", ctx, userID, bid).Return(ia, nil)

	// Expect Income transaction for 0.5 * 10000 = 5000 VND residual
	txSvc.On("Create", ctx, userID, mock.MatchedBy(func(req transaction.CreateRequest) bool {
		return req.Type == "income" && req.Amount == "5000.00"
	})).Return(&domain.Transaction{}, nil)

	_, err := svc.ClaimCorporateAction(ctx, userID, bid, evID, ClaimCorporateActionRequest{})
	assert.NoError(t, err)

	repo.AssertExpectations(t)
	txSvc.AssertExpectations(t)
}

