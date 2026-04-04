package transaction

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
	"github.com/sonbn-225/goen-api-v2/internal/core/money"
)

func makeTransactionTestToken(t *testing.T, secret, userID string) string {
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

func TestHandlerCreateAuthorized(t *testing.T) {
	secret := "test-secret"
	userID := "u-test"

	repo := &fakeTransactionRepo{}
	mod := NewModule(ModuleDeps{Repo: repo})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	body := map[string]any{
		"account_id": "acc-1",
		"type":       "expense",
		"amount":     50,
		"line_items": []map[string]any{{
			"category_id": "cat-1",
			"amount":      50,
		}},
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/transactions/", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+makeTransactionTestToken(t, secret, userID))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusCreated, rr.Code, rr.Body.String())
	}
}

func TestHandlerCreateAuthorizedWithGroupParticipants(t *testing.T) {
	secret := "test-secret"
	userID := "u-test"

	repo := &fakeTransactionRepo{}
	mod := NewModule(ModuleDeps{Repo: repo})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	body := map[string]any{
		"account_id":            "acc-1",
		"type":                  "expense",
		"amount":                "100",
		"line_items":            []map[string]any{{"category_id": "cat-1", "amount": "100"}},
		"owner_original_amount": "50",
		"group_participants": []map[string]any{
			{"participant_name": "Alice", "original_amount": "20"},
			{"participant_name": "Bob", "original_amount": "30"},
		},
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/transactions/", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+makeTransactionTestToken(t, secret, userID))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusCreated, rr.Code, rr.Body.String())
	}
	if len(repo.createOptions) != 1 {
		t.Fatalf("expected create options to be captured once, got %d", len(repo.createOptions))
	}
	if len(repo.createOptions[0].GroupParticipants) != 2 {
		t.Fatalf("expected 2 group participants, got %d", len(repo.createOptions[0].GroupParticipants))
	}
}

func TestHandlerCreateUnauthorized(t *testing.T) {
	repo := &fakeTransactionRepo{}
	mod := NewModule(ModuleDeps{Repo: repo})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware("secret"))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/transactions/", bytes.NewReader([]byte(`{"account_id":"acc-1","type":"expense","amount":50,"line_items":[{"category_id":"cat-1","amount":50}]}`)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestHandlerListAuthorized(t *testing.T) {
	secret := "test-secret"
	userID := "u-list"

	repo := &fakeTransactionRepo{}
	_ = repo.Create(context.Background(), &Transaction{ID: "t1", UserID: userID, Type: "expense", Amount: money.MustFromString("10")}, CreateOptions{})
	_ = repo.Create(context.Background(), &Transaction{ID: "t2", UserID: "other", Type: "expense", Amount: money.MustFromString("20")}, CreateOptions{})

	mod := NewModule(ModuleDeps{Repo: repo})
	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/transactions/", nil)
	req.Header.Set("Authorization", "Bearer "+makeTransactionTestToken(t, secret, userID))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestHandlerGetAuthorized(t *testing.T) {
	secret := "test-secret"
	userID := "u-get"

	repo := &fakeTransactionRepo{}
	_ = repo.Create(context.Background(), &Transaction{ID: "t-get", UserID: userID, Type: "expense", Status: "pending", Amount: money.MustFromString("10")}, CreateOptions{})

	mod := NewModule(ModuleDeps{Repo: repo})
	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/transactions/t-get", nil)
	req.Header.Set("Authorization", "Bearer "+makeTransactionTestToken(t, secret, userID))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestHandlerUpdateAuthorized(t *testing.T) {
	secret := "test-secret"
	userID := "u-update"

	repo := &fakeTransactionRepo{}
	_ = repo.Create(context.Background(), &Transaction{ID: "t-update", UserID: userID, Type: "expense", Status: "pending", Amount: money.MustFromString("10")}, CreateOptions{})

	mod := NewModule(ModuleDeps{Repo: repo})
	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	body := map[string]any{
		"note": "Updated via patch",
		"line_items": []map[string]any{{
			"category_id": "cat-1",
			"amount":      "30",
		}},
		"group_participants": []map[string]any{{
			"participant_name": "Alice",
			"original_amount":  "10",
			"share_amount":     "10",
		}},
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPatch, "/transactions/t-update", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+makeTransactionTestToken(t, secret, userID))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestHandlerBatchPatchStatusAuthorized(t *testing.T) {
	secret := "test-secret"
	userID := "u-batch"

	repo := &fakeTransactionRepo{batchPatchUpdated: []string{"t1"}}
	mod := NewModule(ModuleDeps{Repo: repo})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	body := map[string]any{
		"transaction_ids": []string{"t1", "t2"},
		"patch": map[string]any{
			"status": "posted",
		},
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPatch, "/transactions/batch", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+makeTransactionTestToken(t, secret, userID))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestHandlerListGroupParticipantsAuthorized(t *testing.T) {
	secret := "test-secret"
	userID := "u-group"

	repo := &fakeTransactionRepo{
		items:        []Transaction{{ID: "t1", UserID: userID, Type: "expense", Status: "pending", Amount: money.MustFromString("10")}},
		participants: []GroupExpenseParticipant{{ID: "p1", TransactionID: "t1", ParticipantName: "Alice", OriginalAmount: "10", ShareAmount: "10"}},
	}
	mod := NewModule(ModuleDeps{Repo: repo})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/transactions/t1/group-expense-participants", nil)
	req.Header.Set("Authorization", "Bearer "+makeTransactionTestToken(t, secret, userID))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}
}
