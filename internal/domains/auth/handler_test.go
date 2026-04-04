package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonbn-225/goen-api-v2/internal/core/httpx"
	"github.com/sonbn-225/goen-api-v2/internal/core/security"
	"github.com/sonbn-225/goen-api-v2/internal/domains/auth"
	"github.com/sonbn-225/goen-api-v2/internal/domains/profile"
	"github.com/sonbn-225/goen-api-v2/internal/domains/setting"
	"github.com/sonbn-225/goen-api-v2/internal/infra/postgres"
	repository "github.com/sonbn-225/goen-api-v2/internal/repository"
)

type handlerAuthResponseEnvelope struct {
	Data auth.AuthResponse `json:"data"`
}

type handlerUserResponseEnvelope struct {
	Data auth.User `json:"data"`
}

type handlerErrorResponseEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type authServiceProfileAdapter struct {
	svc auth.Service
}

func (a authServiceProfileAdapter) GetMe(ctx context.Context, userID string) (*auth.User, error) {
	return a.svc.GetMe(ctx, userID)
}

func (a authServiceProfileAdapter) UpdateMyProfile(ctx context.Context, userID string, input auth.UpdateProfileInput) (*auth.User, error) {
	return a.svc.UpdateMyProfile(ctx, userID, input)
}

func (a authServiceProfileAdapter) UploadAvatar(ctx context.Context, userID, _, _ string, _ []byte) (*auth.User, error) {
	return a.svc.UpdateMyAvatar(ctx, userID, "/mock/avatar.jpg")
}

func (a authServiceProfileAdapter) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	return a.svc.ChangePassword(ctx, userID, currentPassword, newPassword)
}

func TestHandlerPatchProfile(t *testing.T) {
	secret := "test-secret"
	repo := newFakeUserRepo()
	_ = repo.CreateUser(context.Background(), auth.UserWithPassword{
		User: auth.User{
			ID:       "u-profile",
			Username: "before",
			Email:    strPtr("before@example.com"),
		},
		PasswordHash: "hashed-Password123",
	})

	mod := auth.NewModule(auth.ModuleDeps{
		UserRepo:         repo,
		Hasher:           fakeHasher{},
		Issuer:           fakeIssuer{},
		AccessTTLMinutes: 60,
	})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	registerProtectedUserRoutes(r, mod)

	payload := map[string]any{"username": "after", "display_name": "After Name"}
	rr := performJSONRequest(t, r, http.MethodPatch, "/profile/me", payload, makeToken(t, secret, "u-profile"))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestHandlerContractSignupEndpoint(t *testing.T) {
	repo := newFakeUserRepo()
	mod := auth.NewModule(auth.ModuleDeps{
		UserRepo:         repo,
		Hasher:           fakeHasher{},
		Issuer:           fakeIssuer{},
		AccessTTLMinutes: 60,
	})

	r := chi.NewRouter()
	mod.RegisterPublicRoutes(r)

	body := map[string]any{
		"email":        "contract-signup@example.com",
		"username":     "contract_signup",
		"display_name": "Contract Signup",
		"password":     "Password123",
	}
	rr := performJSONRequest(t, r, http.MethodPost, "/auth/signup", body, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var resp handlerAuthResponseEnvelope
	decodeJSONResponse(t, rr.Body.Bytes(), &resp)
	if resp.Data.AccessToken == "" {
		t.Fatal("expected access_token")
	}
	if resp.Data.TokenType != "Bearer" {
		t.Fatalf("expected token_type Bearer, got %s", resp.Data.TokenType)
	}
	if resp.Data.ExpiresIn <= 0 {
		t.Fatalf("expected positive expires_in, got %d", resp.Data.ExpiresIn)
	}
	if resp.Data.User.ID == "" {
		t.Fatal("expected user.id")
	}
	if resp.Data.User.Username != "contract_signup" {
		t.Fatalf("expected normalized username contract_signup, got %s", resp.Data.User.Username)
	}
}

func TestHandlerContractSigninEndpoint(t *testing.T) {
	repo := newFakeUserRepo()
	_ = repo.CreateUser(context.Background(), auth.UserWithPassword{
		User: auth.User{
			ID:        "u-contract-signin",
			Username:  "contract_signin",
			Email:     strPtr("contract-signin@example.com"),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
		PasswordHash: "hashed-Password123",
	})

	mod := auth.NewModule(auth.ModuleDeps{
		UserRepo:         repo,
		Hasher:           fakeHasher{},
		Issuer:           fakeIssuer{},
		AccessTTLMinutes: 60,
	})

	r := chi.NewRouter()
	mod.RegisterPublicRoutes(r)

	body := map[string]any{
		"login":    "contract-signin@example.com",
		"password": "Password123",
	}
	rr := performJSONRequest(t, r, http.MethodPost, "/auth/signin", body, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var resp handlerAuthResponseEnvelope
	decodeJSONResponse(t, rr.Body.Bytes(), &resp)
	if resp.Data.User.ID != "u-contract-signin" {
		t.Fatalf("expected user.id u-contract-signin, got %s", resp.Data.User.ID)
	}
	if resp.Data.TokenType != "Bearer" {
		t.Fatalf("expected token_type Bearer, got %s", resp.Data.TokenType)
	}
}

func TestHandlerContractRefreshEndpoint(t *testing.T) {
	secret := "contract-refresh-secret"
	repo := newFakeUserRepo()
	_ = repo.CreateUser(context.Background(), auth.UserWithPassword{
		User: auth.User{
			ID:        "u-contract-refresh",
			Username:  "contract_refresh",
			Email:     strPtr("contract-refresh@example.com"),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
		PasswordHash: "hashed-Password123",
	})

	mod := auth.NewModule(auth.ModuleDeps{
		UserRepo:         repo,
		Hasher:           fakeHasher{},
		Issuer:           fakeIssuer{},
		AccessTTLMinutes: 60,
	})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	registerProtectedUserRoutes(r, mod)

	rr := performJSONRequest(t, r, http.MethodPost, "/auth/refresh", nil, makeToken(t, secret, "u-contract-refresh"))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var resp handlerAuthResponseEnvelope
	decodeJSONResponse(t, rr.Body.Bytes(), &resp)
	if resp.Data.AccessToken != "token-u-contract-refresh" {
		t.Fatalf("expected refreshed access token token-u-contract-refresh, got %s", resp.Data.AccessToken)
	}
	if resp.Data.User.ID != "u-contract-refresh" {
		t.Fatalf("expected user.id u-contract-refresh, got %s", resp.Data.User.ID)
	}
}

func TestHandlerAuthFlowIntegrationPostgresMeProfileSettings(t *testing.T) {
	dbURL := strings.TrimSpace(os.Getenv("GOEN_TEST_DATABASE_URL"))
	if dbURL == "" {
		t.Skip("set GOEN_TEST_DATABASE_URL to run Postgres integration tests")
	}

	pool := mustOpenTestPool(t, dbURL)
	t.Cleanup(func() { pool.Close() })

	ensureUsersSchema(t, pool)

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	email := fmt.Sprintf("it-auth-%s@example.com", suffix)
	username := fmt.Sprintf("it_auth_%s", suffix)
	password := "Password123"

	cleanupTestUsers(t, pool, email, username)
	t.Cleanup(func() { cleanupTestUsers(t, pool, email, username) })

	secret := "integration-secret"
	mod := auth.NewModule(auth.ModuleDeps{
		UserRepo:         repository.NewUserRepository(pool),
		Hasher:           security.NewPasswordHasher(),
		Issuer:           security.NewTokenIssuer(secret, time.Hour),
		AccessTTLMinutes: 60,
	})
	r := authTestRouter(secret, mod)

	signupBody := map[string]any{
		"email":        email,
		"username":     username,
		"display_name": "Integration User",
		"password":     password,
	}
	signupRec := performJSONRequest(t, r, http.MethodPost, "/auth/signup", signupBody, "")
	if signupRec.Code != http.StatusOK {
		t.Fatalf("expected signup status %d, got %d, body=%s", http.StatusOK, signupRec.Code, signupRec.Body.String())
	}

	var signupResp handlerAuthResponseEnvelope
	decodeJSONResponse(t, signupRec.Body.Bytes(), &signupResp)
	if signupResp.Data.AccessToken == "" {
		t.Fatal("expected access token from signup")
	}
	if signupResp.Data.User.ID == "" {
		t.Fatal("expected user id from signup")
	}

	token := signupResp.Data.AccessToken

	meRec := performJSONRequest(t, r, http.MethodGet, "/profile/me", nil, token)
	if meRec.Code != http.StatusOK {
		t.Fatalf("expected me status %d, got %d, body=%s", http.StatusOK, meRec.Code, meRec.Body.String())
	}
	var meResp handlerUserResponseEnvelope
	decodeJSONResponse(t, meRec.Body.Bytes(), &meResp)
	if meResp.Data.Email == nil || *meResp.Data.Email != email {
		t.Fatalf("expected me email %s, got %#v", email, meResp.Data.Email)
	}

	newDisplayName := "Updated Integration User"
	newLocale := "vi-VN"
	profilePatch := map[string]any{
		"display_name": newDisplayName,
	}
	profileRec := performJSONRequest(t, r, http.MethodPatch, "/profile/me", profilePatch, token)
	if profileRec.Code != http.StatusOK {
		t.Fatalf("expected profile patch status %d, got %d, body=%s", http.StatusOK, profileRec.Code, profileRec.Body.String())
	}

	settingsPatch := map[string]any{
		"locale":   newLocale,
		"timezone": "Asia/Ho_Chi_Minh",
	}
	settingsRec := performJSONRequest(t, r, http.MethodPatch, "/settings/me", settingsPatch, token)
	if settingsRec.Code != http.StatusOK {
		t.Fatalf("expected settings patch status %d, got %d, body=%s", http.StatusOK, settingsRec.Code, settingsRec.Body.String())
	}
	var settingsResp handlerUserResponseEnvelope
	decodeJSONResponse(t, settingsRec.Body.Bytes(), &settingsResp)
	if got, _ := settingsResp.Data.Settings["locale"].(string); got != newLocale {
		t.Fatalf("expected updated locale %s, got %#v", newLocale, settingsResp.Data.Settings["locale"])
	}

	meAfterRec := performJSONRequest(t, r, http.MethodGet, "/profile/me", nil, token)
	if meAfterRec.Code != http.StatusOK {
		t.Fatalf("expected me status after patch %d, got %d, body=%s", http.StatusOK, meAfterRec.Code, meAfterRec.Body.String())
	}
	var meAfterResp handlerUserResponseEnvelope
	decodeJSONResponse(t, meAfterRec.Body.Bytes(), &meAfterResp)
	if meAfterResp.Data.DisplayName == nil || *meAfterResp.Data.DisplayName != newDisplayName {
		t.Fatalf("expected display_name %s, got %#v", newDisplayName, meAfterResp.Data.DisplayName)
	}
	if got, _ := meAfterResp.Data.Settings["locale"].(string); got != newLocale {
		t.Fatalf("expected locale %s, got %#v", newLocale, meAfterResp.Data.Settings["locale"])
	}
}

func mustOpenTestPool(t *testing.T, dbURL string) *pgxpool.Pool {
	t.Helper()
	pool, err := postgres.NewPool(dbURL)
	if err != nil {
		t.Fatalf("failed to open postgres pool: %v", err)
	}
	return pool
}

func ensureUsersSchema(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id text PRIMARY KEY,
			username text NOT NULL UNIQUE,
			email text NULL,
			phone text NULL,
			display_name text NULL,
			avatar_url text NULL,
			status text NOT NULL DEFAULT 'active',
			password_hash text NOT NULL,
			settings jsonb NOT NULL DEFAULT '{}'::jsonb,
			created_at timestamptz NOT NULL,
			updated_at timestamptz NOT NULL,
			CONSTRAINT users_email_or_phone_chk CHECK (email IS NOT NULL OR phone IS NOT NULL)
		);
		CREATE UNIQUE INDEX IF NOT EXISTS users_email_uq ON users (lower(email)) WHERE email IS NOT NULL;
		CREATE UNIQUE INDEX IF NOT EXISTS users_phone_uq ON users (phone) WHERE phone IS NOT NULL;
	`)
	if err != nil {
		t.Fatalf("failed to ensure users schema: %v", err)
	}
}

func cleanupTestUsers(t *testing.T, pool *pgxpool.Pool, email, username string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := pool.Exec(ctx, `DELETE FROM users WHERE email = $1 OR username = $2`, email, username)
	if err != nil {
		t.Fatalf("failed to cleanup test users: %v", err)
	}
}

func authTestRouter(secret string, mod *auth.Module) http.Handler {
	r := chi.NewRouter()
	mod.RegisterPublicRoutes(r)
	r.Group(func(r chi.Router) {
		r.Use(httpx.AuthMiddleware(secret))
		registerProtectedUserRoutes(r, mod)
	})
	return r
}

func registerProtectedUserRoutes(r chi.Router, mod *auth.Module) {
	mod.RegisterProtectedRoutes(r)
	profile.NewModule(profile.ModuleDeps{Service: authServiceProfileAdapter{svc: mod.Service}}).RegisterRoutes(r)
	setting.NewModule(setting.ModuleDeps{Service: mod.Service}).RegisterRoutes(r)
}

func TestHandlerSignupInvalidJSON(t *testing.T) {
	repo := newFakeUserRepo()
	mod := auth.NewModule(auth.ModuleDeps{
		UserRepo:         repo,
		Hasher:           fakeHasher{},
		Issuer:           fakeIssuer{},
		AccessTTLMinutes: 60,
	})

	r := chi.NewRouter()
	mod.RegisterPublicRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewBufferString(`{"email":"broken"`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}

	var errResp handlerErrorResponseEnvelope
	decodeJSONResponse(t, rr.Body.Bytes(), &errResp)
	if errResp.Error.Code != "validation_error" {
		t.Fatalf("expected validation_error, got %s", errResp.Error.Code)
	}
}

func TestHandlerRefreshMissingAuthContext(t *testing.T) {
	repo := newFakeUserRepo()
	mod := auth.NewModule(auth.ModuleDeps{
		UserRepo:         repo,
		Hasher:           fakeHasher{},
		Issuer:           fakeIssuer{},
		AccessTTLMinutes: 60,
	})

	r := chi.NewRouter()
	registerProtectedUserRoutes(r, mod)

	rr := performJSONRequest(t, r, http.MethodPost, "/auth/refresh", nil, "")
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusUnauthorized, rr.Code, rr.Body.String())
	}

	var errResp handlerErrorResponseEnvelope
	decodeJSONResponse(t, rr.Body.Bytes(), &errResp)
	if errResp.Error.Code != "unauthorized" {
		t.Fatalf("expected unauthorized, got %s", errResp.Error.Code)
	}
}

func TestHandlerPatchProfileBadPayload(t *testing.T) {
	secret := "test-secret"
	repo := newFakeUserRepo()
	_ = repo.CreateUser(context.Background(), auth.UserWithPassword{
		User: auth.User{
			ID:       "u-bad-payload",
			Username: "before",
			Email:    strPtr("before@example.com"),
		},
		PasswordHash: "hashed-Password123",
	})

	mod := auth.NewModule(auth.ModuleDeps{
		UserRepo:         repo,
		Hasher:           fakeHasher{},
		Issuer:           fakeIssuer{},
		AccessTTLMinutes: 60,
	})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	registerProtectedUserRoutes(r, mod)

	payload := map[string]any{"email": "invalid-email-format"}
	rr := performJSONRequest(t, r, http.MethodPatch, "/profile/me", payload, makeToken(t, secret, "u-bad-payload"))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}

	var errResp handlerErrorResponseEnvelope
	decodeJSONResponse(t, rr.Body.Bytes(), &errResp)
	if errResp.Error.Code != "validation_error" {
		t.Fatalf("expected validation_error, got %s", errResp.Error.Code)
	}
}

func TestHandlerUploadAvatarMissingField(t *testing.T) {
	secret := "test-secret"
	repo := newFakeUserRepo()
	_ = repo.CreateUser(context.Background(), auth.UserWithPassword{
		User: auth.User{
			ID:       "u-avatar-missing",
			Username: "avatar_missing",
			Email:    strPtr("avatar-missing@example.com"),
		},
		PasswordHash: "hashed-Password123",
	})

	mod := auth.NewModule(auth.ModuleDeps{
		UserRepo:         repo,
		Hasher:           fakeHasher{},
		Issuer:           fakeIssuer{},
		AccessTTLMinutes: 60,
	})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	registerProtectedUserRoutes(r, mod)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/profile/me/avatar", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+makeToken(t, secret, "u-avatar-missing"))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
}

func TestHandlerUploadAvatarSuccess(t *testing.T) {
	secret := "test-secret"
	repo := newFakeUserRepo()
	_ = repo.CreateUser(context.Background(), auth.UserWithPassword{
		User: auth.User{
			ID:       "u-avatar-success",
			Username: "avatar_success",
			Email:    strPtr("avatar-success@example.com"),
		},
		PasswordHash: "hashed-Password123",
	})

	mod := auth.NewModule(auth.ModuleDeps{
		UserRepo:         repo,
		Hasher:           fakeHasher{},
		Issuer:           fakeIssuer{},
		AccessTTLMinutes: 60,
	})

	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware(secret))
	registerProtectedUserRoutes(r, mod)

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
	req.Header.Set("Authorization", "Bearer "+makeToken(t, secret, "u-avatar-success"))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var resp handlerUserResponseEnvelope
	decodeJSONResponse(t, rr.Body.Bytes(), &resp)
	if resp.Data.AvatarURL == nil || *resp.Data.AvatarURL == "" {
		t.Fatalf("expected avatar_url, got %#v", resp.Data.AvatarURL)
	}
}

func performJSONRequest(t *testing.T, handler http.Handler, method, path string, body any, bearerToken string) *httptest.ResponseRecorder {
	t.Helper()
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}
		reader = bytes.NewReader(raw)
	}

	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+bearerToken)
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func decodeJSONResponse(t *testing.T, raw []byte, out any) {
	t.Helper()
	if err := json.Unmarshal(raw, out); err != nil {
		t.Fatalf("failed to decode response body %s: %v", string(raw), err)
	}
}

func makeToken(t *testing.T, secret, userID string) string {
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
