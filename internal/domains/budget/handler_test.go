package budget

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
	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/core/httpx"
)

func makeBudgetTestToken(t *testing.T, secret, userID string) string {
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

type fakeBudgetService struct {
	items     []WithStats
	item      *WithStats
	createErr error
	listErr   error
	getErr    error
}

func (s *fakeBudgetService) Create(_ context.Context, _ string, input CreateInput) (*WithStats, error) {
	if s.createErr != nil {
		return nil, s.createErr
	}
	created := &WithStats{Budget: Budget{ID: "b1", UserID: "u1", Amount: input.Amount}, Spent: "0", Remaining: input.Amount, PercentUsed: 0}
	return created, nil
}

func (s *fakeBudgetService) Get(_ context.Context, _ string, _ string) (*WithStats, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	if s.item == nil {
		return nil, apperrors.New(apperrors.KindNotFound, "budget not found")
	}
	cloned := *s.item
	return &cloned, nil
}

func (s *fakeBudgetService) List(_ context.Context, _ string) ([]WithStats, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	return s.items, nil
}

func TestBudgetHandlerListAuthorized(t *testing.T) {
	secret := "test-secret"
	svc := &fakeBudgetService{items: []WithStats{{Budget: Budget{ID: "b1", UserID: "u1"}}}}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/budgets/", nil)
	req.Header.Set("Authorization", "Bearer "+makeBudgetTestToken(t, secret, "u1"))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestBudgetHandlerCreateAuthorized(t *testing.T) {
	secret := "test-secret"
	svc := &fakeBudgetService{}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	body := map[string]any{
		"period":       "month",
		"period_start": "2026-04-01",
		"amount":       "1000.00",
		"category_id":  "cat_def_food",
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/budgets/", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+makeBudgetTestToken(t, secret, "u1"))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusCreated, rr.Code, rr.Body.String())
	}
}

func TestBudgetHandlerCreateInvalidJSON(t *testing.T) {
	secret := "test-secret"
	svc := &fakeBudgetService{}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/budgets/", bytes.NewReader([]byte(`{"period":`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+makeBudgetTestToken(t, secret, "u1"))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
}

func TestBudgetHandlerGetNotFound(t *testing.T) {
	secret := "test-secret"
	svc := &fakeBudgetService{getErr: apperrors.New(apperrors.KindNotFound, "budget not found")}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/budgets/missing", nil)
	req.Header.Set("Authorization", "Bearer "+makeBudgetTestToken(t, secret, "u1"))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusNotFound, rr.Code, rr.Body.String())
	}
}
