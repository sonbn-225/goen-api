package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/auth"
	"github.com/sonbn-225/goen-api/internal/apierror"
	"github.com/sonbn-225/goen-api/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

type SignupRequest struct {
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
}

type SigninRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken string       `json:"access_token"`
	TokenType   string       `json:"token_type"`
	ExpiresIn   int          `json:"expires_in"`
	User        storage.User `json:"user"`
}

// Signup godoc
// @Summary Signup
// @Description Create a new user account (email or phone required).
// @Tags auth
// @Accept json
// @Produce json
// @Param X-Client-Id header string false "Client instance ID (recommended)"
// @Param body body SignupRequest true "Signup request"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} apierror.Envelope
// @Failure 409 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /auth/signup [post]
func Signup(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if d.DB == nil {
			apierror.Write(w, http.StatusInternalServerError, "internal_error", "DATABASE_URL not set", nil)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if err := storage.EnsureUsersSchema(ctx, d.DB); err != nil {
			apierror.Write(w, http.StatusInternalServerError, "internal_error", "database not ready", map[string]any{"cause": err.Error()})
			return
		}

		var req SignupRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "invalid json", nil)
			return
		}

		email := strings.TrimSpace(req.Email)
		phone := strings.TrimSpace(req.Phone)
		password := req.Password
		if email == "" && phone == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "email or phone is required", map[string]any{"field": "email|phone"})
			return
		}
		if len(password) < 8 {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "password must be at least 8 characters", map[string]any{"field": "password"})
			return
		}

		var emailPtr *string
		if email != "" {
			e := strings.ToLower(email)
			emailPtr = &e
		}
		var phonePtr *string
		if phone != "" {
			p := phone
			phonePtr = &p
		}
		var displayNamePtr *string
		if dn := strings.TrimSpace(req.DisplayName); dn != "" {
			displayNamePtr = &dn
		}

		hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			apierror.Write(w, http.StatusInternalServerError, "internal_error", "failed to hash password", nil)
			return
		}

		now := time.Now().UTC()
		user := storage.User{
			ID:          uuid.NewString(),
			Email:       emailPtr,
			Phone:       phonePtr,
			DisplayName: displayNamePtr,
			Status:      "active",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		created, err := storage.CreateUser(ctx, d.DB, user, string(hashBytes))
		if err != nil {
			if err == storage.ErrUserAlreadyExists {
				apierror.Write(w, http.StatusConflict, "conflict", "email/phone already in use", nil)
				return
			}
			apierror.Write(w, http.StatusInternalServerError, "internal_error", "failed to create user", nil)
			return
		}

		token, expiresIn, err := issueAccessToken(d, created.ID)
		if err != nil {
			apierror.Write(w, http.StatusInternalServerError, "internal_error", "failed to issue token", nil)
			return
		}

		writeJSON(w, http.StatusOK, AuthResponse{
			AccessToken: token,
			TokenType:   "Bearer",
			ExpiresIn:   expiresIn,
			User:        *created,
		})
	}
}

// Signin godoc
// @Summary Signin
// @Description Sign in with email or phone.
// @Tags auth
// @Accept json
// @Produce json
// @Param X-Client-Id header string false "Client instance ID (recommended)"
// @Param body body SigninRequest true "Signin request"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} apierror.Envelope
// @Failure 401 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /auth/signin [post]
func Signin(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if d.DB == nil {
			apierror.Write(w, http.StatusInternalServerError, "internal_error", "DATABASE_URL not set", nil)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if err := storage.EnsureUsersSchema(ctx, d.DB); err != nil {
			apierror.Write(w, http.StatusInternalServerError, "internal_error", "database not ready", map[string]any{"cause": err.Error()})
			return
		}

		var req SigninRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "invalid json", nil)
			return
		}
		login := strings.TrimSpace(req.Login)
		if login == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "login is required", map[string]any{"field": "login"})
			return
		}
		if req.Password == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "password is required", map[string]any{"field": "password"})
			return
		}

		u, err := storage.GetUserWithPasswordByLogin(ctx, d.DB, login)
		if err != nil {
			if err == storage.ErrUserNotFound {
				apierror.Write(w, http.StatusUnauthorized, "unauthorized", "invalid credentials", nil)
				return
			}
			apierror.Write(w, http.StatusInternalServerError, "internal_error", "failed to lookup user", nil)
			return
		}

		if u.Status != "active" {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "user is disabled", nil)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "invalid credentials", nil)
			return
		}

		token, expiresIn, err := issueAccessToken(d, u.ID)
		if err != nil {
			apierror.Write(w, http.StatusInternalServerError, "internal_error", "failed to issue token", nil)
			return
		}

		writeJSON(w, http.StatusOK, AuthResponse{
			AccessToken: token,
			TokenType:   "Bearer",
			ExpiresIn:   expiresIn,
			User:        u.User,
		})
	}
}

// Me godoc
// @Summary Get current user
// @Description Returns the current authenticated user.
// @Tags auth
// @Produce json
// @Success 200 {object} storage.User
// @Failure 401 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Security BearerAuth
// @Router /auth/me [get]
func Me(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if d.DB == nil {
			apierror.Write(w, http.StatusInternalServerError, "internal_error", "DATABASE_URL not set", nil)
			return
		}

		userID, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "missing user", nil)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		user, err := storage.GetUserByID(ctx, d.DB, userID)
		if err != nil {
			if err == storage.ErrUserNotFound {
				apierror.Write(w, http.StatusUnauthorized, "unauthorized", "user not found", nil)
				return
			}
			apierror.Write(w, http.StatusInternalServerError, "internal_error", "failed to load user", nil)
			return
		}

		writeJSON(w, http.StatusOK, user)
	}
}

func issueAccessToken(d Deps, userID string) (token string, expiresInSeconds int, err error) {
	now := time.Now().UTC()
	exp := now.Add(time.Duration(d.Cfg.JWTAccessTTLMinutes) * time.Minute)

	claims := jwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(exp),
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := t.SignedString([]byte(d.Cfg.JWTSecret))
	if err != nil {
		return "", 0, err
	}
	return signed, int(time.Until(exp).Seconds()), nil
}
