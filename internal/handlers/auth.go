package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/sonbn-225/goen-api/internal/auth"
	"github.com/sonbn-225/goen-api/internal/apierror"
	"github.com/sonbn-225/goen-api/internal/domain"
	"github.com/sonbn-225/goen-api/internal/services"
)

// Signup godoc
// @Summary Signup
// @Description Create a new user account (email or phone required).
// @Tags auth
// @Accept json
// @Produce json
// @Param X-Client-Id header string false "Client instance ID (recommended)"
// @Param body body services.SignupRequest true "Signup request"
// @Success 200 {object} services.AuthResponse
// @Failure 400 {object} apierror.Envelope
// @Failure 409 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /auth/signup [post]
func Signup(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req services.SignupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "invalid json", nil)
			return
		}

		resp, err := d.AuthService.Signup(r.Context(), req)
		if err != nil {
			if errors.Is(err, domain.ErrUserAlreadyExists) {
				apierror.Write(w, http.StatusConflict, "conflict", "user already exists", nil)
				return
			}
			apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// Signin godoc
// @Summary Signin
// @Description Sign in with email/phone and password.
// @Tags auth
// @Accept json
// @Produce json
// @Param X-Client-Id header string false "Client instance ID (recommended)"
// @Param body body services.SigninRequest true "Signin request"
// @Success 200 {object} services.AuthResponse
// @Failure 400 {object} apierror.Envelope
// @Failure 401 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /auth/signin [post]
func Signin(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req services.SigninRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "invalid json", nil)
			return
		}

		resp, err := d.AuthService.Signin(r.Context(), req)
		if err != nil {
			// In a real app we might verify if it's "invalid credentials" or internal error
			// The service returns "invalid credentials" for not found or bad password.
			if err.Error() == "invalid credentials" {
				apierror.Write(w, http.StatusUnauthorized, "unauthorized", "invalid credentials", nil)
				return
			}
			apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// Me godoc
// @Summary Get current user
// @Description Get current logged in user information.
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} domain.User
// @Failure 401 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /auth/me [get]
func Me(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}

		user, err := d.AuthService.GetMe(r.Context(), uid)
		if err != nil {
			if errors.Is(err, domain.ErrUserNotFound) {
				apierror.Write(w, http.StatusUnauthorized, "unauthorized", "user not found", nil)
				return
			}
			apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}
}

// PatchMySettings godoc
// @Summary Update current user settings
// @Description Merge-patch current user's settings (JSON object) into stored settings.
// @Tags auth
// @Accept json
// @Produce json
// @Param body body object true "Settings patch (JSON object)"
// @Success 200 {object} domain.User
// @Failure 400 {object} apierror.Envelope
// @Failure 401 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /auth/me/settings [patch]
func PatchMySettings(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}

		patch := map[string]any{}
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&patch); err != nil {
			if errors.Is(err, io.EOF) {
				patch = map[string]any{}
			} else {
				apierror.Write(w, http.StatusBadRequest, "validation_error", "invalid json", nil)
				return
			}
		}

		user, err := d.AuthService.UpdateMySettings(r.Context(), uid, patch)
		if err != nil {
			if errors.Is(err, domain.ErrUserNotFound) {
				apierror.Write(w, http.StatusUnauthorized, "unauthorized", "user not found", nil)
				return
			}
			apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}
}
