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

// Signup godoc
// @Summary User Signup
// @Description Register a new user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.SignupRequest true "Signup request"
// @Success 201 {object} dto.AuthResponse
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /auth/signup [post]
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

// Signin godoc
// @Summary User Signin
// @Description Authenticate a user and return tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.SigninRequest true "Signin request"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /auth/signin [post]
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

// Refresh godoc
// @Summary Refresh Token
// @Description Get a new access token using an existing session
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.AuthResponse
// @Failure 401 {object} response.ErrorEnvelope
// @Router /auth/refresh [post]
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

// Me godoc
// @Summary Get Current User
// @Description Retrieve the profile of the currently authenticated user
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} entity.User
// @Failure 401 {object} response.ErrorEnvelope
// @Router /auth/me [get]
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

// PatchMySettings godoc
// @Summary Update User Settings
// @Description Update the settings for the currently authenticated user
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]interface{} true "Settings mappings"
// @Success 200 {object} entity.User
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /auth/me/settings [patch]
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

// UploadAvatar godoc
// @Summary Upload Avatar
// @Description Upload a new avatar image for the user
// @Tags Auth
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param avatar formData file true "Avatar image file"
// @Success 200 {object} entity.User
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /auth/me/avatar [post]
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

// PatchMyProfile godoc
// @Summary Update User Profile
// @Description Update user information such as email, phone, displayName, or username
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object true "Profile Update Body"
// @Success 200 {object} entity.User
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /auth/me/profile [patch]
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

// ChangePassword godoc
// @Summary Change Password
// @Description Change the password for the current user
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object true "Password Change Body"
// @Success 200 {object} map[string]string
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /auth/me/change-password [post]
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

// GetMedia godoc
// @Summary Proxy Media Images
// @Description Retrieve images via minio reverse proxy
// @Tags Public
// @Produce image/*
// @Param bucket path string true "Bucket Name"
// @Param key path string true "Object Key"
// @Success 200 {file} file
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 503 {object} response.ErrorEnvelope
// @Router /media/{bucket}/{key} [get]
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
