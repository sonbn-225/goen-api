package v1

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/handler/middleware"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/response"
	"github.com/sonbn-225/goen-api/internal/pkg/storage"
)

type AuthHandler struct {
	svc interfaces.AuthService
	s3  *storage.S3Client
}

func NewAuthHandler(svc interfaces.AuthService, s3 *storage.S3Client) *AuthHandler {
	return &AuthHandler{
		svc: svc,
		s3:  s3,
	}
}

func (h *AuthHandler) RegisterRoutes(r chi.Router, cfg *config.Config) {
	r.Post("/auth/signup", h.Signup)
	r.Post("/auth/signin", h.Signin)

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg))
		r.Post("/auth/refresh", h.Refresh)
		r.Get("/auth/me", h.Me)
		r.Patch("/auth/me/settings", h.PatchMySettings)
		r.Post("/auth/me/avatar", h.UploadAvatar)
		r.Patch("/auth/me/profile", h.PatchMyProfile)
		r.Post("/auth/me/change-password", h.ChangePassword)
	})

	// Public media proxy
	r.Get("/media/{bucket}/*", h.GetMedia)
}

func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req dto.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "failed to decode request", nil)
		return
	}

	res, err := h.svc.Signup(r.Context(), req)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}

	response.WriteJSON(w, http.StatusCreated, res)
}

func (h *AuthHandler) Signin(w http.ResponseWriter, r *http.Request) {
	var req dto.SigninRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "failed to decode request", nil)
		return
	}

	res, err := h.svc.Signin(r.Context(), req)
	if err != nil {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", err.Error(), nil)
		return
	}

	response.WriteJSON(w, http.StatusOK, res)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	res, err := h.svc.Refresh(r.Context(), userID)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, res)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	res, err := h.svc.GetMe(r.Context(), userID)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, res)
}

func (h *AuthHandler) PatchMySettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	var patch map[string]any
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "failed to decode request", nil)
		return
	}

	res, err := h.svc.UpdateMySettings(r.Context(), userID, patch)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, res)
}

func (h *AuthHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "failed to parse multipart form", nil)
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "missing avatar file", nil)
		return
	}
	defer file.Close()

	res, err := h.svc.UploadAvatar(r.Context(), userID, header)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, res)
}

func (h *AuthHandler) PatchMyProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	var body struct {
		DisplayName *string `json:"display_name"`
		Email       *string `json:"email"`
		Phone       *string `json:"phone"`
		Username    *string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "failed to decode request", nil)
		return
	}

	res, err := h.svc.UpdateMyProfile(r.Context(), userID, body.DisplayName, body.Email, body.Phone, body.Username)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, res)
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "failed to decode request", nil)
		return
	}

	if err := h.svc.ChangePassword(r.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		response.WriteError(w, http.StatusBadRequest, "bad_request", err.Error(), nil)
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]string{"message": "password updated successfully"})
}

func (h *AuthHandler) GetMedia(w http.ResponseWriter, r *http.Request) {
	if h.s3 == nil {
		response.WriteError(w, http.StatusServiceUnavailable, "unavailable", "storage not configured", nil)
		return
	}

	bucket := chi.URLParam(r, "bucket")
	key := strings.TrimPrefix(chi.URLParam(r, "*"), "/")
	if bucket == "" || key == "" {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "missing bucket or key", nil)
		return
	}

	obj, info, err := h.s3.GetObject(r.Context(), bucket, key)
	if err != nil {
		response.WriteError(w, http.StatusNotFound, "not_found", "media not found", nil)
		return
	}
	defer obj.Close()

	w.Header().Set("Content-Type", info.ContentType)
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, obj)
}
