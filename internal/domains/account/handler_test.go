package account

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

func makeAccountTestToken(t *testing.T, secret, userID string) string {
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

	repo := &fakeAccountRepo{}
	mod := NewModule(ModuleDeps{Repo: repo})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	body := map[string]any{"name": "Wallet", "type": "cash"}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/accounts/", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+makeAccountTestToken(t, secret, userID))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusCreated, rr.Code, rr.Body.String())
	}
}

func TestHandlerCreateAuthorizedFrontendPayload(t *testing.T) {
	secret := "test-secret"
	userID := "u-test-frontend"

	repo := &fakeAccountRepo{}
	mod := NewModule(ModuleDeps{Repo: repo})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	body := map[string]any{
		"name":              "Main Wallet",
		"account_type":      "wallet",
		"currency":          "VND",
		"account_number":    "123456789",
		"parent_account_id": "",
		"color":             "#1A73E8",
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/accounts/", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+makeAccountTestToken(t, secret, userID))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusCreated, rr.Code, rr.Body.String())
	}
}

func TestHandlerCreateRejectsInvalidHexColor(t *testing.T) {
	secret := "test-secret"
	userID := "u-test-invalid-color"

	repo := &fakeAccountRepo{}
	mod := NewModule(ModuleDeps{Repo: repo})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	body := map[string]any{
		"name":         "Main Wallet",
		"account_type": "wallet",
		"color":        "blue",
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/accounts/", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+makeAccountTestToken(t, secret, userID))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
}

func TestHandlerCreateUnauthorized(t *testing.T) {
	repo := &fakeAccountRepo{}
	mod := NewModule(ModuleDeps{Repo: repo})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware("secret"))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/accounts/", bytes.NewReader([]byte(`{"name":"Wallet","type":"cash"}`)))
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

	repo := &fakeAccountRepo{}
	_ = repo.Create(context.Background(), &Account{ID: "a1", UserID: userID, Name: "Wallet", Type: "cash"})
	_ = repo.Create(context.Background(), &Account{ID: "a2", UserID: "other", Name: "Other", Type: "cash"})

	mod := NewModule(ModuleDeps{Repo: repo})
	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/accounts/", nil)
	req.Header.Set("Authorization", "Bearer "+makeAccountTestToken(t, secret, userID))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestHandlerGetAuthorized(t *testing.T) {
	secret := "test-secret"
	userID := "u-get"

	repo := &fakeAccountRepo{}
	_ = repo.Create(context.Background(), &Account{ID: "a-get", UserID: userID, Name: "Wallet", Type: "cash"})

	mod := NewModule(ModuleDeps{Repo: repo})
	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/accounts/a-get", nil)
	req.Header.Set("Authorization", "Bearer "+makeAccountTestToken(t, secret, userID))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestHandlerDeleteAuthorized(t *testing.T) {
	secret := "test-secret"
	userID := "u-del"

	repo := &fakeAccountRepo{}
	_ = repo.Create(context.Background(), &Account{ID: "a-del", UserID: userID, Name: "Wallet", Type: "wallet"})

	mod := NewModule(ModuleDeps{Repo: repo})
	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodDelete, "/accounts/a-del", nil)
	req.Header.Set("Authorization", "Bearer "+makeAccountTestToken(t, secret, userID))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusNoContent, rr.Code, rr.Body.String())
	}
}

func TestHandlerDeleteNotFound(t *testing.T) {
	secret := "test-secret"
	userID := "u-del"

	repo := &fakeAccountRepo{}
	mod := NewModule(ModuleDeps{Repo: repo})
	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodDelete, "/accounts/missing", nil)
	req.Header.Set("Authorization", "Bearer "+makeAccountTestToken(t, secret, userID))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusNotFound, rr.Code, rr.Body.String())
	}
}

func TestHandlerDeleteUnauthorized(t *testing.T) {
	repo := &fakeAccountRepo{}
	mod := NewModule(ModuleDeps{Repo: repo})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware("secret"))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodDelete, "/accounts/a-del", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}
