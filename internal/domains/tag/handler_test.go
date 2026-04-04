package tag

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

func makeTagTestToken(t *testing.T, secret, userID string) string {
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

type fakeTagService struct {
	items     []Tag
	item      *Tag
	createErr error
	listErr   error
	getErr    error
}

func (s *fakeTagService) Create(_ context.Context, _ string, input CreateInput) (*Tag, error) {
	if s.createErr != nil {
		return nil, s.createErr
	}
	created := Tag{ID: "t1", UserID: "u1", NameVI: input.NameVI, NameEN: input.NameEN, Color: input.Color, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	return &created, nil
}

func (s *fakeTagService) Get(_ context.Context, _ string, _ string) (*Tag, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	if s.item == nil {
		return nil, apperrors.New(apperrors.KindNotFound, "tag not found")
	}
	cloned := *s.item
	return &cloned, nil
}

func (s *fakeTagService) List(_ context.Context, _ string) ([]Tag, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	return s.items, nil
}

func (s *fakeTagService) GetOrCreateByName(_ context.Context, _ string, _ string, _ string) (string, error) {
	return "t1", nil
}

func TestTagHandlerListAuthorized(t *testing.T) {
	secret := "test-secret"
	svc := &fakeTagService{items: []Tag{{ID: "t1", UserID: "u1"}}}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/tags/", nil)
	req.Header.Set("Authorization", "Bearer "+makeTagTestToken(t, secret, "u1"))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestTagHandlerCreateAuthorized(t *testing.T) {
	secret := "test-secret"
	svc := &fakeTagService{}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	body := map[string]any{"name_en": "Food", "color": "#00ff00"}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/tags/", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+makeTagTestToken(t, secret, "u1"))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusCreated, rr.Code, rr.Body.String())
	}
}

func TestTagHandlerCreateInvalidJSON(t *testing.T) {
	secret := "test-secret"
	svc := &fakeTagService{}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/tags/", bytes.NewReader([]byte(`{"name_en":`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+makeTagTestToken(t, secret, "u1"))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
}

func TestTagHandlerGetNotFound(t *testing.T) {
	secret := "test-secret"
	svc := &fakeTagService{getErr: apperrors.New(apperrors.KindNotFound, "tag not found")}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/tags/missing", nil)
	req.Header.Set("Authorization", "Bearer "+makeTagTestToken(t, secret, "u1"))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusNotFound, rr.Code, rr.Body.String())
	}
}
