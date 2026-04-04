package report

import (
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

type fakeReportService struct {
	err error
}

func (s *fakeReportService) GetDashboardReport(_ context.Context, _ string) (*DashboardReport, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &DashboardReport{
		TotalBalances:    []AccountBalance{{AccountID: "a1", Currency: "VND", Balance: "1000"}},
		Cashflow6Months:  []CashflowStat{{Month: "2026-04", Income: "1500", Expense: "300"}},
		TopExpensesMonth: []CategoryExpenseStat{{CategoryID: "cat_food", Amount: "300"}},
	}, nil
}

func makeReportTestToken(t *testing.T, secret, userID string) string {
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

func TestReportHandlerGetDashboardAuthorized(t *testing.T) {
	secret := "test-secret"
	mod := NewModule(ModuleDeps{Service: &fakeReportService{}})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/reports/dashboard", nil)
	req.Header.Set("Authorization", "Bearer "+makeReportTestToken(t, secret, "u1"))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var payload struct {
		Data DashboardReport `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(payload.Data.TotalBalances) != 1 {
		t.Fatalf("expected one balance row, got %#v", payload.Data)
	}
}

func TestReportHandlerGetDashboardUnauthorized(t *testing.T) {
	mod := NewModule(ModuleDeps{Service: &fakeReportService{}})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware("secret"))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/reports/dashboard", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}
