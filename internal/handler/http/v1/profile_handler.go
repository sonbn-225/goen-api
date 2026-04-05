package v1

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/handler/middleware"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/response"
)

type ProfileHandler struct {
	svc interfaces.AuthService
}

func NewProfileHandler(svc interfaces.AuthService) *ProfileHandler {
	return &ProfileHandler{svc: svc}
}

func (h *ProfileHandler) RegisterRoutes(r chi.Router, cfg *config.Config) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg))

		r.Route("/profile/me", func(r chi.Router) {
			r.Get("/", h.Me)
			r.Patch("/", h.PatchMyProfile)
			r.Route("/avatar", func(r chi.Router) {
				r.Get("/", h.GetMyAvatars)
				r.Post("/", h.UploadAvatar)
			})
			r.Post("/change-password", h.ChangePassword)
		})
	})
}

// Me godoc
// @Summary Get Current User
// @Description Retrieve the profile of the currently authenticated user
// @Tags Profile
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.SuccessEnvelope{data=dto.UserResponse}
// @Failure 401 {object} response.ErrorEnvelope
// @Router /profile/me [get]
func (h *ProfileHandler) Me(w http.ResponseWriter, r *http.Request) {
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

	response.WriteSuccess(w, http.StatusOK, res)
}

// PatchMyProfile godoc
// @Summary Update User Profile
// @Description Update user information such as email, phone, displayName, or username
// @Tags Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object true "Profile Update Body"
// @Success 200 {object} response.SuccessEnvelope{data=dto.UserResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /profile/me [patch]
func (h *ProfileHandler) PatchMyProfile(w http.ResponseWriter, r *http.Request) {
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

	response.WriteSuccess(w, http.StatusOK, res)
}

// UploadAvatar godoc
// @Summary Upload Avatar
// @Description Upload a new avatar image for the user
// @Tags Profile
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param avatar formData file true "Avatar image file"
// @Success 200 {object} response.SuccessEnvelope{data=dto.UserResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /profile/me/avatar [post]
func (h *ProfileHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
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

	response.WriteSuccess(w, http.StatusOK, res)
}

// ChangePassword godoc
// @Summary Change Password
// @Description Change the password for the current user
// @Tags Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object true "Password Change Body"
// @Success 200 {object} response.SuccessEnvelope{data=map[string]string}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /profile/me/change-password [post]
func (h *ProfileHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
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

	response.WriteSuccess(w, http.StatusOK, map[string]string{"message": "password updated successfully"})
}

// GetMyAvatars godoc
// @Summary List My Avatars
// @Description Retrieve a list of all previously uploaded avatars for the current user
// @Tags Profile
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.SuccessEnvelope{data=[]dto.MediaResponse}
// @Failure 401 {object} response.ErrorEnvelope
// @Router /profile/me/avatar [get]
func (h *ProfileHandler) GetMyAvatars(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	res, err := h.svc.GetMyAvatars(r.Context(), userID)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteSuccess(w, http.StatusOK, res)
}
