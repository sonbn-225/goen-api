package category

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

func makeCategoryTestToken(t *testing.T, secret, userID string) string {
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

type fakeCategoryService struct {
	items   []Category
	item    *Category
	listErr error
	getErr  error
}

func (s *fakeCategoryService) Get(_ context.Context, _ string, _ string) (*Category, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	if s.item == nil {
		return nil, apperrors.New(apperrors.KindNotFound, "category not found")
	}
	cloned := *s.item
	return &cloned, nil
}

func (s *fakeCategoryService) List(_ context.Context, _ string, _ string) ([]Category, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	return s.items, nil
}

type categoryErrorEnvelope struct {
	Error struct {
		Code string `json:"code"`
	} `json:"error"`
}

func assertCategoryErrorCode(t *testing.T, body []byte, expected string) {
	t.Helper()
	var resp categoryErrorEnvelope
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if resp.Error.Code != expected {
		t.Fatalf("expected error code %q, got %q, body=%s", expected, resp.Error.Code, string(body))
	}
}

func TestHandlerCategoryListAuthorized(t *testing.T) {
	secret := "test-secret"
	svc := &fakeCategoryService{items: []Category{{ID: "cat_food"}}}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/categories/", nil)
	req.Header.Set("Authorization", "Bearer "+makeCategoryTestToken(t, secret, "u-category"))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestHandlerCategoryListUnauthorized(t *testing.T) {
	svc := &fakeCategoryService{}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware("secret"))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/categories/", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestHandlerCategoryListValidationError(t *testing.T) {
	secret := "test-secret"
	svc := &fakeCategoryService{listErr: apperrors.New(apperrors.KindValidation, "type must be one of income, expense, both")}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/categories/?type=invalid", bytes.NewReader(nil))
	req.Header.Set("Authorization", "Bearer "+makeCategoryTestToken(t, secret, "u-category"))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
	assertCategoryErrorCode(t, rr.Body.Bytes(), "validation_error")
}

func TestHandlerCategoryGetAuthorized(t *testing.T) {
	secret := "test-secret"
	svc := &fakeCategoryService{item: &Category{ID: "cat_food"}}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/categories/cat_food", nil)
	req.Header.Set("Authorization", "Bearer "+makeCategoryTestToken(t, secret, "u-category"))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestHandlerCategoryGetNotFound(t *testing.T) {
	secret := "test-secret"
	svc := &fakeCategoryService{getErr: apperrors.New(apperrors.KindNotFound, "category not found")}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/categories/missing", nil)
	req.Header.Set("Authorization", "Bearer "+makeCategoryTestToken(t, secret, "u-category"))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusNotFound, rr.Code, rr.Body.String())
	}
	assertCategoryErrorCode(t, rr.Body.Bytes(), "not_found")
}
