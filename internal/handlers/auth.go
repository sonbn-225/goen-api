package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/sonbn-225/goen-api/internal/apierror"
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
		if ok := decodeJSON(w, r, &req); !ok {
			return
		}

		resp, err := d.AuthService.Signup(r.Context(), req)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, resp)
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
		if ok := decodeJSON(w, r, &req); !ok {
			return
		}

		resp, err := d.AuthService.Signin(r.Context(), req)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, resp)
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
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		user, err := d.AuthService.GetMe(r.Context(), uid)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, user)
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
		uid, ok := requireUserID(w, r)
		if !ok {
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
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, user)
	}
}
