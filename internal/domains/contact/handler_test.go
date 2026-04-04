package contact

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

func makeContactTestToken(t *testing.T, secret, userID string) string {
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

type fakeContactService struct {
	items   []Contact
	item    *Contact
	getErr  error
	listErr error
}

func (s *fakeContactService) Create(_ context.Context, _ string, input CreateInput) (*Contact, error) {
	created := &Contact{ID: "c1", UserID: "u1", Name: input.Name}
	return created, nil
}

func (s *fakeContactService) Get(_ context.Context, _ string, _ string) (*Contact, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	if s.item == nil {
		return nil, apperrors.New(apperrors.KindNotFound, "contact not found")
	}
	cloned := *s.item
	return &cloned, nil
}

func (s *fakeContactService) List(_ context.Context, _ string) ([]Contact, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	return s.items, nil
}

func (s *fakeContactService) Update(_ context.Context, _ string, _ string, input UpdateInput) (*Contact, error) {
	updated := &Contact{ID: "c1", UserID: "u1", Name: input.Name}
	return updated, nil
}

func (s *fakeContactService) Delete(_ context.Context, _ string, _ string) error {
	return nil
}

func TestContactHandlerListAuthorized(t *testing.T) {
	secret := "test-secret"
	svc := &fakeContactService{items: []Contact{{ID: "c1", UserID: "u1", Name: "Alice"}}}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/contacts/", nil)
	req.Header.Set("Authorization", "Bearer "+makeContactTestToken(t, secret, "u1"))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestContactHandlerCreateInvalidJSON(t *testing.T) {
	secret := "test-secret"
	svc := &fakeContactService{}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/contacts/", bytes.NewReader([]byte(`{"name":`)))
	req.Header.Set("Authorization", "Bearer "+makeContactTestToken(t, secret, "u1"))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
}

func TestContactHandlerGetNotFound(t *testing.T) {
	secret := "test-secret"
	svc := &fakeContactService{getErr: apperrors.New(apperrors.KindNotFound, "contact not found")}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/contacts/missing", nil)
	req.Header.Set("Authorization", "Bearer "+makeContactTestToken(t, secret, "u1"))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusNotFound, rr.Code, rr.Body.String())
	}
}

func TestContactHandlerDeleteAuthorized(t *testing.T) {
	secret := "test-secret"
	svc := &fakeContactService{}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodDelete, "/contacts/c1", nil)
	req.Header.Set("Authorization", "Bearer "+makeContactTestToken(t, secret, "u1"))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected valid response body, got %v", err)
	}
}
