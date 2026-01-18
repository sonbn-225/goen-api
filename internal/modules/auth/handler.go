package auth

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/apperrors"
	"github.com/sonbn-225/goen-api/internal/config"
	"github.com/sonbn-225/goen-api/internal/httpapi"
	"github.com/sonbn-225/goen-api/internal/response"
)

// Handler handles HTTP requests for authentication.
type Handler struct {
	svc *Service
}

// NewHandler creates a new auth handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes registers all auth routes on the given router.
func (h *Handler) RegisterRoutes(r chi.Router, cfg *config.Config) {
	r.Post("/auth/signup", h.Signup)
	r.Post("/auth/signin", h.Signin)
	r.With(httpapi.AuthMiddleware(cfg)).Get("/auth/me", h.Me)
	r.With(httpapi.AuthMiddleware(cfg)).Patch("/auth/me/settings", h.PatchMySettings)
}

// Signup handles POST /auth/signup
// @Summary Signup
// @Description Create a new user account (email or phone required).
// @Tags auth
// @Accept json
// @Produce json
// @Param body body SignupRequest true "Signup request"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 409 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /auth/signup [post]
func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}

	resp, err := h.svc.Signup(r.Context(), req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, resp)
}

// Signin handles POST /auth/signin
// @Summary Signin
// @Description Sign in with email/phone and password.
// @Tags auth
// @Accept json
// @Produce json
// @Param body body SigninRequest true "Signin request"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /auth/signin [post]
func (h *Handler) Signin(w http.ResponseWriter, r *http.Request) {
	var req SigninRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}

	resp, err := h.svc.Signin(r.Context(), req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, resp)
}

// Me handles GET /auth/me
// @Summary Get current user
// @Description Get current logged in user information.
// @Tags auth
// @Produce json
// @Success 200 {object} domain.User
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /auth/me [get]
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	user, err := h.svc.GetMe(r.Context(), userID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, user)
}

// PatchMySettings handles PATCH /auth/me/settings
// @Summary Update current user settings
// @Description Merge-patch current user's settings.
// @Tags auth
// @Accept json
// @Produce json
// @Param body body object true "Settings patch"
// @Success 200 {object} domain.User
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /auth/me/settings [patch]
func (h *Handler) PatchMySettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	patch := map[string]any{}
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		if !errors.Is(err, io.EOF) {
			response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json", nil)
			return
		}
	}

	user, err := h.svc.UpdateMySettings(r.Context(), userID, patch)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, user)
}

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	var se *apperrors.Error
	if errors.As(err, &se) {
		response.WriteError(w, se.HTTPStatus(), string(se.Kind), se.Message, se.Details)
		return
	}
	response.WriteInternalError(w, err)
}
