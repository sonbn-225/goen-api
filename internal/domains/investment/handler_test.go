package investment

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sonbn-225/goen-api-v2/internal/core/httpx"
)

type fakeInvestmentService struct{}

func (s *fakeInvestmentService) ListInvestmentAccounts(_ context.Context, _ string) ([]InvestmentAccount, error) {
	return []InvestmentAccount{{ID: "ia1", AccountID: "a1", Currency: "VND"}}, nil
}

func (s *fakeInvestmentService) GetInvestmentAccount(_ context.Context, _ string, _ string) (*InvestmentAccount, error) {
	return &InvestmentAccount{ID: "ia1", AccountID: "a1", Currency: "VND"}, nil
}

func (s *fakeInvestmentService) UpdateInvestmentAccountSettings(_ context.Context, _ string, _ string, _ UpdateInvestmentAccountSettingsInput) (*InvestmentAccount, error) {
	return &InvestmentAccount{ID: "ia1", AccountID: "a1", Currency: "VND"}, nil
}

func (s *fakeInvestmentService) ListSecurities(_ context.Context, _ string) ([]Security, error) {
	return []Security{{ID: "sec1", Symbol: "AAA"}}, nil
}

func (s *fakeInvestmentService) GetSecurity(_ context.Context, _ string, _ string) (*Security, error) {
	return &Security{ID: "sec1", Symbol: "AAA"}, nil
}

func (s *fakeInvestmentService) ListSecurityPrices(_ context.Context, _ string, _ string, _ *string, _ *string) ([]SecurityPriceDaily, error) {
	return []SecurityPriceDaily{{ID: "p1", SecurityID: "sec1", PriceDate: "2026-04-03", Close: "10"}}, nil
}

func (s *fakeInvestmentService) ListSecurityEvents(_ context.Context, _ string, _ string, _ *string, _ *string) ([]SecurityEvent, error) {
	return []SecurityEvent{{ID: "e1", SecurityID: "sec1", EventType: "cash_dividend"}}, nil
}

func (s *fakeInvestmentService) CreateTrade(_ context.Context, _ string, investmentAccountID string, input CreateTradeInput) (*Trade, error) {
	return &Trade{ID: "t1", BrokerAccountID: investmentAccountID, SecurityID: input.SecurityID, Side: input.Side, Quantity: "1.00000000", Price: "10.00000000", Fees: "0.00", Taxes: "0.00", OccurredAt: time.Now().UTC(), CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}, nil
}

func (s *fakeInvestmentService) ListTrades(_ context.Context, _ string, investmentAccountID string) ([]Trade, error) {
	return []Trade{{ID: "t1", BrokerAccountID: investmentAccountID, SecurityID: "sec1", Side: "buy", Quantity: "1.00000000", Price: "10.00000000", Fees: "0.00", Taxes: "0.00", OccurredAt: time.Now().UTC(), CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}}, nil
}

func (s *fakeInvestmentService) ListHoldings(_ context.Context, _ string, investmentAccountID string) ([]Holding, error) {
	return []Holding{{ID: "h1", BrokerAccountID: investmentAccountID, SecurityID: "sec1", Quantity: "1.00000000", SourceOfTruth: "trades", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}}, nil
}

func makeInvestmentTestToken(t *testing.T, secret, userID string) string {
	t.Helper()
	claims := jwt.MapClaims{
		"sub": userID,
		"iat": time.Now().Add(-1 * time.Minute).Unix(),
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return signed
}

func TestInvestmentHandlerListAccountsAuthorized(t *testing.T) {
	secret := "test-secret"
	mod := NewModule(ModuleDeps{Service: &fakeInvestmentService{}})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/investment-accounts/", nil)
	req.Header.Set("Authorization", "Bearer "+makeInvestmentTestToken(t, secret, "u1"))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestInvestmentHandlerCreateTradeAuthorized(t *testing.T) {
	secret := "test-secret"
	mod := NewModule(ModuleDeps{Service: &fakeInvestmentService{}})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	body := map[string]any{
		"security_id": "sec1",
		"side":        "buy",
		"quantity":    "1",
		"price":       "10",
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/investment-accounts/ia1/trades", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+makeInvestmentTestToken(t, secret, "u1"))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusCreated, rr.Code, rr.Body.String())
	}
}

func TestInvestmentHandlerListSecuritiesAuthorized(t *testing.T) {
	secret := "test-secret"
	mod := NewModule(ModuleDeps{Service: &fakeInvestmentService{}})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/securities/", nil)
	req.Header.Set("Authorization", "Bearer "+makeInvestmentTestToken(t, secret, "u1"))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}
}
