package setting

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/core/httpx"
	"github.com/sonbn-225/goen-api-v2/internal/domains/auth"
)

func makeSettingTestToken(t *testing.T, secret, userID string) string {
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

type settingUserEnvelope struct {
	Data auth.User `json:"data"`
}

type settingErrorEnvelope struct {
	Error struct {
		Code string `json:"code"`
	} `json:"error"`
}

func assertSettingErrorCode(t *testing.T, body []byte, expected string) {
	t.Helper()
	var resp settingErrorEnvelope
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if resp.Error.Code != expected {
		t.Fatalf("expected error code %q, got %q, body=%s", expected, resp.Error.Code, string(body))
	}
}

func TestHandlerPatchSettingsAuthorized(t *testing.T) {
	secret := "test-secret"
	svc := &fakeSettingService{user: auth.User{ID: "u-setting", Username: "setting_user"}}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	body := map[string]any{"locale": "vi-VN", "timezone": "Asia/Ho_Chi_Minh"}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPatch, "/settings/me", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+makeSettingTestToken(t, secret, "u-setting"))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var resp settingUserEnvelope
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got, _ := resp.Data.Settings["locale"].(string); got != "vi-VN" {
		t.Fatalf("expected locale vi-VN, got %#v", resp.Data.Settings["locale"])
	}
}

func TestHandlerPatchSettingsUnauthorized(t *testing.T) {
	svc := &fakeSettingService{}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware("secret"))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPatch, "/settings/me", bytes.NewReader([]byte(`{"locale":"vi-VN"}`)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestHandlerPatchSettingsInvalidJSON(t *testing.T) {
	secret := "test-secret"
	svc := &fakeSettingService{user: auth.User{ID: "u-setting-invalid-json", Username: "setting_user"}}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPatch, "/settings/me", strings.NewReader(`{"locale":`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+makeSettingTestToken(t, secret, "u-setting-invalid-json"))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
	assertSettingErrorCode(t, rr.Body.Bytes(), "validation_error")
}

func TestHandlerPatchSettingsServiceNotFound(t *testing.T) {
	secret := "test-secret"
	svc := &fakeSettingService{
		user:      auth.User{ID: "u-setting-notfound", Username: "setting_user"},
		updateErr: apperrors.New(apperrors.KindNotFound, "user not found"),
	}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	body := map[string]any{"locale": "vi-VN"}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPatch, "/settings/me", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+makeSettingTestToken(t, secret, "u-setting-notfound"))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusNotFound, rr.Code, rr.Body.String())
	}
	assertSettingErrorCode(t, rr.Body.Bytes(), "not_found")
}

func TestHandlerPatchSettingsServiceConflict(t *testing.T) {
	secret := "test-secret"
	svc := &fakeSettingService{
		user:      auth.User{ID: "u-setting-conflict", Username: "setting_user"},
		updateErr: apperrors.New(apperrors.KindConflict, "setting key is immutable"),
	}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	body := map[string]any{"timezone": "UTC"}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPatch, "/settings/me", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+makeSettingTestToken(t, secret, "u-setting-conflict"))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusConflict, rr.Code, rr.Body.String())
	}
	assertSettingErrorCode(t, rr.Body.Bytes(), "conflict")
}
