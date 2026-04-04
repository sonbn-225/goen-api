package debt

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sonbn-225/goen-api-v2/internal/core/httpx"
)

func makeDebtTestToken(t *testing.T, secret, userID string) string {
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

func TestDebtHandlerListAuthorized(t *testing.T) {
	secret := "test-secret"
	svc := &fakeDebtService{}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/debts/", nil)
	req.Header.Set("Authorization", "Bearer "+makeDebtTestToken(t, secret, "u1"))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestDebtHandlerCreateInvalidJSON(t *testing.T) {
	secret := "test-secret"
	svc := &fakeDebtService{}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/debts/", bytes.NewReader([]byte(`{"account_id":`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+makeDebtTestToken(t, secret, "u1"))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
}

func TestDebtHandlerCreatePaymentAuthorized(t *testing.T) {
	secret := "test-secret"
	svc := &fakeDebtService{}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	body := map[string]any{"transaction_id": "t1"}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/debts/d1/payments", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+makeDebtTestToken(t, secret, "u1"))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusCreated, rr.Code, rr.Body.String())
	}
}

func TestDebtHandlerListDebtLinksForTransactionAuthorized(t *testing.T) {
	secret := "test-secret"
	svc := &fakeDebtService{}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/transactions/t1/debt-links", nil)
	req.Header.Set("Authorization", "Bearer "+makeDebtTestToken(t, secret, "u1"))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}
}
