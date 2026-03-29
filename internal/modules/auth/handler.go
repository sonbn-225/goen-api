package auth

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/config"
	"github.com/sonbn-225/goen-api/internal/platform/httpx"
	"github.com/sonbn-225/goen-api/internal/response"
	"github.com/sonbn-225/goen-api/internal/storage"
)

// Handler handles HTTP requests for authentication.
type Handler struct {
	svc *Service
	s3  *storage.S3Client
}

// NewHandler creates a new auth handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc, s3: svc.s3}
}

// RegisterRoutes registers all auth routes on the given router.
func (h *Handler) RegisterRoutes(r chi.Router, cfg *config.Config) {
	r.Post("/auth/signup", h.Signup)
	r.Post("/auth/signin", h.Signin)
	r.With(httpx.AuthMiddleware(cfg)).Post("/auth/refresh", h.Refresh)
	r.With(httpx.AuthMiddleware(cfg)).Get("/auth/me", h.Me)
	r.With(httpx.AuthMiddleware(cfg)).Patch("/auth/me/settings", h.PatchMySettings)
	r.With(httpx.AuthMiddleware(cfg)).Post("/auth/me/avatar", h.UploadAvatar)
	r.With(httpx.AuthMiddleware(cfg)).Patch("/auth/me/profile", h.PatchMyProfile)
	r.With(httpx.AuthMiddleware(cfg)).Post("/auth/me/change-password", h.ChangePassword)
	// Public media proxy (no auth — object keys are UUIDs)
	r.Get("/media/{bucket}/*", h.GetMedia)
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
// @Failure 409 {object} response.ErrorEnvelope (conflict: email, phone, or username already exists)
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
		httpx.WriteServiceError(w, err)
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
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, resp)
}

// Refresh handles POST /auth/refresh
// @Summary Refresh access token
// @Description Issue a new access token for the current authenticated user.
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} AuthResponse
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /auth/refresh [post]
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	resp, err := h.svc.Refresh(r.Context(), userID)
	if err != nil {
		httpx.WriteServiceError(w, err)
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
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	user, err := h.svc.GetMe(r.Context(), userID)
	if err != nil {
		httpx.WriteServiceError(w, err)
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
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
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
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, user)
}

// UploadAvatar handles POST /auth/me/avatar
// @Summary Upload profile avatar
// @Description Upload a profile image (multipart/form-data, field "avatar").
// @Tags auth
// @Accept multipart/form-data
// @Produce json
// @Param avatar formData file true "Image file"
// @Success 200 {object} domain.User
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 503 {object} response.ErrorEnvelope
// @Router /auth/me/avatar [post]
func (h *Handler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "failed to parse multipart form", nil)
		return
	}

	file, fh, err := r.FormFile("avatar")
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "field 'avatar' not found", nil)
		return
	}
	defer file.Close()

	user, err := h.svc.UploadAvatar(r.Context(), userID, fh)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, user)
}

// PatchMyProfile handles PATCH /auth/me/profile
// @Summary Update profile (name, email, phone)
// @Tags auth
// @Accept json
// @Produce json
// @Param body body object true "Profile patch (display_name, email, phone, username)"
// @Success 200 {object} domain.User
// @Router /auth/me/profile [patch]
func (h *Handler) PatchMyProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	var body struct {
		DisplayName *string `json:"display_name"`
		Email       *string `json:"email"`
		Phone       *string `json:"phone"`
		Username    *string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil && !errors.Is(err, io.EOF) {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json", nil)
		return
	}

	user, err := h.svc.UpdateMyProfile(r.Context(), userID, body.DisplayName, body.Email, body.Phone, body.Username)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, user)
}

// ChangePassword handles POST /auth/me/change-password
// @Summary Change password
// @Tags auth
// @Accept json
// @Produce json
// @Param body body object true "Password change request"
// @Success 200 {object} response.Envelope
// @Router /auth/me/change-password [post]
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json", nil)
		return
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "required fields missing", nil)
		return
	}

	if err := h.svc.ChangePassword(r.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, response.Envelope{Message: "password updated successfully"})
}

// GetMedia handles GET /media/{bucket}/*
// Streams a file from SeaweedFS without requiring authentication.
// Object keys are unguessable UUIDs so this is safe to keep public.
// @Summary Get media file
// @Tags media
// @Produce application/octet-stream
// @Param bucket path string true "Bucket name"
// @Param key path string true "Object key"
// @Success 200 {file} binary
// @Failure 404 {object} response.ErrorEnvelope
// @Router /media/{bucket}/{key} [get]
func (h *Handler) GetMedia(w http.ResponseWriter, r *http.Request) {
	if h.s3 == nil {
		response.WriteError(w, http.StatusServiceUnavailable, "unavailable", "storage not configured", nil)
		return
	}

	bucket := chi.URLParam(r, "bucket")
	key := strings.TrimPrefix(chi.URLParam(r, "*"), "/")
	if strings.TrimSpace(bucket) == "" || strings.TrimSpace(key) == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid path", nil)
		return
	}

	obj, info, err := h.s3.GetObject(r.Context(), bucket, key)
	if err != nil {
		response.WriteError(w, http.StatusNotFound, "not_found", "media not found", nil)
		return
	}
	defer obj.Close()

	w.Header().Set("Content-Type", info.ContentType)
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, obj)
}

