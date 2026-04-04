package rotatingsavings

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sonbn-225/goen-api-v2/internal/core/httpx"
)

func makeRotatingTestToken(t *testing.T, secret, userID string) string {
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

func TestRotatingSavingsHandlerListGroupsAuthorized(t *testing.T) {
	secret := "test-secret"
	mod := NewModule(ModuleDeps{Service: &fakeRotatingService{}})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/rotating-savings/groups/", nil)
	req.Header.Set("Authorization", "Bearer "+makeRotatingTestToken(t, secret, "u1"))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var payload struct {
		Data []GroupSummary `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
}

func TestRotatingSavingsHandlerUnauthorized(t *testing.T) {
	mod := NewModule(ModuleDeps{Service: &fakeRotatingService{}})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware("secret"))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/rotating-savings/groups/", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestRotatingSavingsHandlerCreateGroupBadPayload(t *testing.T) {
	secret := "test-secret"
	mod := NewModule(ModuleDeps{Service: &fakeRotatingService{}})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/rotating-savings/groups/", strings.NewReader("{bad-json"))
	req.Header.Set("Authorization", "Bearer "+makeRotatingTestToken(t, secret, "u1"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}
