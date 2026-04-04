package savings

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

type fakeSavingsService struct {
	listItems []SavingsInstrument
	err       error
}

func (s *fakeSavingsService) Create(_ context.Context, _ string, input CreateInput) (*SavingsInstrument, error) {
	item := SavingsInstrument{ID: "ins_created", Principal: input.Principal, SavingsAccountID: "acc_sav", ParentAccountID: "acc_parent", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	return &item, s.err
}

func (s *fakeSavingsService) Get(_ context.Context, _ string, instrumentID string) (*SavingsInstrument, error) {
	item := SavingsInstrument{ID: instrumentID, Principal: "1000", SavingsAccountID: "acc_sav", ParentAccountID: "acc_parent", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	return &item, s.err
}

func (s *fakeSavingsService) List(_ context.Context, _ string) ([]SavingsInstrument, error) {
	if s.listItems == nil {
		return []SavingsInstrument{}, s.err
	}
	return s.listItems, s.err
}

func (s *fakeSavingsService) Patch(_ context.Context, _ string, instrumentID string, _ PatchInput) (*SavingsInstrument, error) {
	item := SavingsInstrument{ID: instrumentID, Principal: "1200", SavingsAccountID: "acc_sav", ParentAccountID: "acc_parent", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	return &item, s.err
}

func (s *fakeSavingsService) Delete(_ context.Context, _, _ string) error {
	return s.err
}

func makeSavingsTestToken(t *testing.T, secret, userID string) string {
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

func TestSavingsHandlerListAuthorized(t *testing.T) {
	secret := "test-secret"
	mod := NewModule(ModuleDeps{Service: &fakeSavingsService{listItems: []SavingsInstrument{{ID: "ins_1", Principal: "1000", SavingsAccountID: "acc_sav", ParentAccountID: "acc_parent"}}}})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/savings/instruments/", nil)
	req.Header.Set("Authorization", "Bearer "+makeSavingsTestToken(t, secret, "u1"))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var payload struct {
		Data []SavingsInstrument `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(payload.Data) != 1 {
		t.Fatalf("expected one item, got %#v", payload.Data)
	}
}

func TestSavingsHandlerListUnauthorized(t *testing.T) {
	mod := NewModule(ModuleDeps{Service: &fakeSavingsService{}})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware("secret"))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/savings/instruments/", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestSavingsHandlerCreateBadPayload(t *testing.T) {
	secret := "test-secret"
	mod := NewModule(ModuleDeps{Service: &fakeSavingsService{}})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/savings/instruments/", bytes.NewBufferString("{bad-json"))
	req.Header.Set("Authorization", "Bearer "+makeSavingsTestToken(t, secret, "u1"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
}
