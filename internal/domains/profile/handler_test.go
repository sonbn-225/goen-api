package profile

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
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

func makeProfileTestToken(t *testing.T, secret, userID string) string {
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

type profileUserEnvelope struct {
	Data auth.User `json:"data"`
}

type profileErrorEnvelope struct {
	Error struct {
		Code string `json:"code"`
	} `json:"error"`
}

func assertProfileErrorCode(t *testing.T, body []byte, expected string) {
	t.Helper()
	var resp profileErrorEnvelope
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if resp.Error.Code != expected {
		t.Fatalf("expected error code %q, got %q, body=%s", expected, resp.Error.Code, string(body))
	}
}

func TestHandlerMeAuthorized(t *testing.T) {
	secret := "test-secret"
	svc := &fakeProfileService{user: auth.User{ID: "u-profile", Username: "profile_user"}}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/profile/me", nil)
	req.Header.Set("Authorization", "Bearer "+makeProfileTestToken(t, secret, "u-profile"))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestHandlerMeUnauthorized(t *testing.T) {
	svc := &fakeProfileService{}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware("secret"))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/profile/me", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestHandlerPatchProfileAuthorized(t *testing.T) {
	secret := "test-secret"
	svc := &fakeProfileService{user: auth.User{ID: "u-profile", Username: "before"}}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	body := map[string]any{"username": "after", "display_name": "After Name"}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPatch, "/profile/me", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+makeProfileTestToken(t, secret, "u-profile"))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestHandlerPatchProfileInvalidJSON(t *testing.T) {
	secret := "test-secret"
	svc := &fakeProfileService{user: auth.User{ID: "u-profile-invalid-json", Username: "before"}}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPatch, "/profile/me", strings.NewReader(`{"username":`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+makeProfileTestToken(t, secret, "u-profile-invalid-json"))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
	assertProfileErrorCode(t, rr.Body.Bytes(), "validation_error")
}

func TestHandlerPatchProfileServiceNotFound(t *testing.T) {
	secret := "test-secret"
	svc := &fakeProfileService{
		user:             auth.User{ID: "u-profile-notfound", Username: "before"},
		updateProfileErr: apperrors.New(apperrors.KindNotFound, "user not found"),
	}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	body := map[string]any{"display_name": "After Name"}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPatch, "/profile/me", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+makeProfileTestToken(t, secret, "u-profile-notfound"))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusNotFound, rr.Code, rr.Body.String())
	}
	assertProfileErrorCode(t, rr.Body.Bytes(), "not_found")
}

func TestHandlerPatchProfileServiceConflict(t *testing.T) {
	secret := "test-secret"
	svc := &fakeProfileService{
		user:             auth.User{ID: "u-profile-conflict", Username: "before"},
		updateProfileErr: apperrors.New(apperrors.KindConflict, "username already exists"),
	}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	body := map[string]any{"username": "existing"}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPatch, "/profile/me", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+makeProfileTestToken(t, secret, "u-profile-conflict"))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusConflict, rr.Code, rr.Body.String())
	}
	assertProfileErrorCode(t, rr.Body.Bytes(), "conflict")
}

func TestHandlerUploadAvatarMissingField(t *testing.T) {
	secret := "test-secret"
	svc := &fakeProfileService{user: auth.User{ID: "u-avatar", Username: "avatar_user"}}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/profile/me/avatar", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+makeProfileTestToken(t, secret, "u-avatar"))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
}

func TestHandlerUploadAvatarUnsupportedMime(t *testing.T) {
	secret := "test-secret"
	svc := &fakeProfileService{user: auth.User{ID: "u-avatar-unsupported", Username: "avatar_user"}}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("avatar", "avatar.txt")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := part.Write([]byte("plain-text-avatar")); err != nil {
		t.Fatalf("failed to write avatar bytes: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/profile/me/avatar", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+makeProfileTestToken(t, secret, "u-avatar-unsupported"))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
	assertProfileErrorCode(t, rr.Body.Bytes(), "validation_error")
}

func TestHandlerUploadAvatarTooLarge(t *testing.T) {
	secret := "test-secret"
	svc := &fakeProfileService{user: auth.User{ID: "u-avatar-too-large", Username: "avatar_user"}}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("avatar", "avatar.png")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	pngHeader := []byte{137, 80, 78, 71, 13, 10, 26, 10}
	large := append(pngHeader, bytes.Repeat([]byte("a"), int(maxAvatarSizeBytes)+1)...)
	if _, err := part.Write(large); err != nil {
		t.Fatalf("failed to write avatar bytes: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/profile/me/avatar", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+makeProfileTestToken(t, secret, "u-avatar-too-large"))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
	assertProfileErrorCode(t, rr.Body.Bytes(), "validation_error")
}

func TestHandlerUploadAvatarSuccess(t *testing.T) {
	secret := "test-secret"
	svc := &fakeProfileService{user: auth.User{ID: "u-avatar-success", Username: "avatar_success"}}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("avatar", "avatar.png")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	pngHeader := []byte{137, 80, 78, 71, 13, 10, 26, 10}
	if _, err := part.Write(append(pngHeader, []byte("test-avatar")...)); err != nil {
		t.Fatalf("failed to write avatar bytes: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/profile/me/avatar", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+makeProfileTestToken(t, secret, "u-avatar-success"))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var resp profileUserEnvelope
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Data.AvatarURL == nil || *resp.Data.AvatarURL == "" {
		t.Fatalf("expected avatar_url, got %#v", resp.Data.AvatarURL)
	}

}

func TestHandlerChangePasswordAuthorized(t *testing.T) {
	secret := "test-secret"
	svc := &fakeProfileService{user: auth.User{ID: "u-password", Username: "password_user"}}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	body := map[string]any{"current_password": "Password123", "new_password": "NewPassword9"}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/profile/me/change-password", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+makeProfileTestToken(t, secret, "u-password"))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestHandlerChangePasswordInvalidJSON(t *testing.T) {
	secret := "test-secret"
	svc := &fakeProfileService{user: auth.User{ID: "u-password-invalid-json", Username: "password_user"}}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/profile/me/change-password", strings.NewReader(`{"current_password":"Password123"`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+makeProfileTestToken(t, secret, "u-password-invalid-json"))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
	assertProfileErrorCode(t, rr.Body.Bytes(), "validation_error")
}

func TestHandlerChangePasswordServiceNotFound(t *testing.T) {
	secret := "test-secret"
	svc := &fakeProfileService{
		user:              auth.User{ID: "u-password-notfound", Username: "password_user"},
		changePasswordErr: apperrors.New(apperrors.KindNotFound, "user not found"),
	}
	mod := NewModule(ModuleDeps{Service: svc})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	mod.RegisterRoutes(r)

	body := map[string]any{"current_password": "Password123", "new_password": "NewPassword9"}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/profile/me/change-password", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+makeProfileTestToken(t, secret, "u-password-notfound"))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusNotFound, rr.Code, rr.Body.String())
	}
	assertProfileErrorCode(t, rr.Body.Bytes(), "not_found")
}
