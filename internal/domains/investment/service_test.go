package investment

import (
	"context"
	"testing"
	"time"

	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
)

type fakeInvestmentRepo struct {
	accounts       []InvestmentAccount
	securities     []Security
	holdings       map[string]Holding
	trades         []Trade
	prices         []SecurityPriceDaily
	events         []SecurityEvent
	updateErr      error
	createTradeErr error
}

func (r *fakeInvestmentRepo) GetInvestmentAccount(_ context.Context, _ string, investmentAccountID string) (*InvestmentAccount, error) {
	for _, item := range r.accounts {
		if item.ID == investmentAccountID {
			cloned := item
			return &cloned, nil
		}
	}
	return nil, nil
}

func (r *fakeInvestmentRepo) ListInvestmentAccounts(_ context.Context, _ string) ([]InvestmentAccount, error) {
	return r.accounts, nil
}

func (r *fakeInvestmentRepo) UpdateInvestmentAccountSettings(_ context.Context, _ string, investmentAccountID string, feeSettings any, taxSettings any) (*InvestmentAccount, error) {
	if r.updateErr != nil {
		return nil, r.updateErr
	}
	for i, item := range r.accounts {
		if item.ID != investmentAccountID {
			continue
		}
		if feeSettings != nil {
			item.FeeSettings = feeSettings
		}
		if taxSettings != nil {
			item.TaxSettings = taxSettings
		}
		item.UpdatedAt = time.Now().UTC()
		r.accounts[i] = item
		cloned := item
		return &cloned, nil
	}
	return nil, nil
}

func (r *fakeInvestmentRepo) GetSecurity(_ context.Context, securityID string) (*Security, error) {
	for _, item := range r.securities {
		if item.ID == securityID {
			cloned := item
			return &cloned, nil
		}
	}
	return nil, nil
}

func (r *fakeInvestmentRepo) ListSecurities(_ context.Context) ([]Security, error) {
	return r.securities, nil
}

func (r *fakeInvestmentRepo) ListSecurityPrices(_ context.Context, _ string, _ *string, _ *string) ([]SecurityPriceDaily, error) {
	return r.prices, nil
}

func (r *fakeInvestmentRepo) ListSecurityEvents(_ context.Context, _ string, _ *string, _ *string) ([]SecurityEvent, error) {
	return r.events, nil
}

func (r *fakeInvestmentRepo) CreateTrade(_ context.Context, _ string, trade Trade) error {
	if r.createTradeErr != nil {
		return r.createTradeErr
	}
	r.trades = append(r.trades, trade)
	return nil
}

func (r *fakeInvestmentRepo) ListTrades(_ context.Context, _ string, brokerAccountID string) ([]Trade, error) {
	out := make([]Trade, 0)
	for _, item := range r.trades {
		if item.BrokerAccountID == brokerAccountID {
			out = append(out, item)
		}
	}
	return out, nil
}

func (r *fakeInvestmentRepo) ListHoldings(_ context.Context, _ string, brokerAccountID string) ([]Holding, error) {
	out := make([]Holding, 0)
	for _, item := range r.holdings {
		if item.BrokerAccountID == brokerAccountID {
			out = append(out, item)
		}
	}
	return out, nil
}

func (r *fakeInvestmentRepo) GetHolding(_ context.Context, _ string, brokerAccountID, securityID string) (*Holding, error) {
	item, ok := r.holdings[holdingKey(brokerAccountID, securityID)]
	if !ok {
		return nil, nil
	}
	cloned := item
	return &cloned, nil
}

func (r *fakeInvestmentRepo) UpsertHolding(_ context.Context, _ string, holding Holding) (*Holding, error) {
	if r.holdings == nil {
		r.holdings = make(map[string]Holding)
	}
	r.holdings[holdingKey(holding.BrokerAccountID, holding.SecurityID)] = holding
	cloned := holding
	return &cloned, nil
}

func holdingKey(brokerAccountID, securityID string) string {
	return brokerAccountID + ":" + securityID
}

func TestServiceCreateTradeBuyUpdatesHolding(t *testing.T) {
	repo := &fakeInvestmentRepo{
		accounts:   []InvestmentAccount{{ID: "ia1", AccountID: "a1", Currency: "VND"}},
		securities: []Security{{ID: "sec1", Symbol: "AAA"}},
		holdings:   map[string]Holding{},
	}
	svc := NewService(repo)

	created, err := svc.CreateTrade(context.Background(), "u1", "ia1", CreateTradeInput{
		SecurityID: "sec1",
		Side:       "buy",
		Quantity:   "10",
		Price:      "5",
		Fees:       strPtr("1"),
		Taxes:      strPtr("0.5"),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if created == nil {
		t.Fatalf("expected created trade, got nil")
	}
	if len(repo.trades) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(repo.trades))
	}

	h, ok := repo.holdings[holdingKey("ia1", "sec1")]
	if !ok {
		t.Fatalf("expected holding to be upserted")
	}
	if h.Quantity != "10.00000000" {
		t.Fatalf("expected quantity 10.00000000, got %s", h.Quantity)
	}
	if h.CostBasisTotal == nil || *h.CostBasisTotal != "51.50" {
		t.Fatalf("expected cost basis 51.50, got %#v", h.CostBasisTotal)
	}
	if h.AvgCost == nil || *h.AvgCost != "5.15000000" {
		t.Fatalf("expected avg cost 5.15000000, got %#v", h.AvgCost)
	}
}

func TestServiceCreateTradeSellRejectsInsufficientHolding(t *testing.T) {
	repo := &fakeInvestmentRepo{
		accounts:   []InvestmentAccount{{ID: "ia1", AccountID: "a1", Currency: "VND"}},
		securities: []Security{{ID: "sec1", Symbol: "AAA"}},
		holdings: map[string]Holding{
			holdingKey("ia1", "sec1"): {BrokerAccountID: "ia1", SecurityID: "sec1", Quantity: "2.00000000", CostBasisTotal: strPtr("20.00")},
		},
	}
	svc := NewService(repo)

	_, err := svc.CreateTrade(context.Background(), "u1", "ia1", CreateTradeInput{
		SecurityID: "sec1",
		Side:       "sell",
		Quantity:   "3",
		Price:      "12",
	})
	if apperrors.KindOf(err) != apperrors.KindValidation {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestServiceCreateTradeSellUpdatesHolding(t *testing.T) {
	repo := &fakeInvestmentRepo{
		accounts:   []InvestmentAccount{{ID: "ia1", AccountID: "a1", Currency: "VND"}},
		securities: []Security{{ID: "sec1", Symbol: "AAA"}},
		holdings: map[string]Holding{
			holdingKey("ia1", "sec1"): {BrokerAccountID: "ia1", SecurityID: "sec1", Quantity: "10.00000000", CostBasisTotal: strPtr("100.00")},
		},
	}
	svc := NewService(repo)

	_, err := svc.CreateTrade(context.Background(), "u1", "ia1", CreateTradeInput{
		SecurityID: "sec1",
		Side:       "sell",
		Quantity:   "4",
		Price:      "12",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	h := repo.holdings[holdingKey("ia1", "sec1")]
	if h.Quantity != "6.00000000" {
		t.Fatalf("expected quantity 6.00000000, got %s", h.Quantity)
	}
	if h.CostBasisTotal == nil || *h.CostBasisTotal != "60.00" {
		t.Fatalf("expected cost basis 60.00, got %#v", h.CostBasisTotal)
	}
	if h.AvgCost == nil || *h.AvgCost != "10.00000000" {
		t.Fatalf("expected avg cost 10.00000000, got %#v", h.AvgCost)
	}
}

func TestServiceUpdateInvestmentAccountSettingsRequiresFields(t *testing.T) {
	repo := &fakeInvestmentRepo{accounts: []InvestmentAccount{{ID: "ia1", AccountID: "a1", Currency: "VND"}}}
	svc := NewService(repo)

	_, err := svc.UpdateInvestmentAccountSettings(context.Background(), "u1", "ia1", UpdateInvestmentAccountSettingsInput{})
	if apperrors.KindOf(err) != apperrors.KindValidation {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func strPtr(v string) *string {
	return &v
}
