package profile

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/core/httpx"
	"github.com/sonbn-225/goen-api-v2/internal/core/response"
	"github.com/sonbn-225/goen-api-v2/internal/domains/auth"
)

const maxAvatarSizeBytes int64 = 5 << 20

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// me godoc
// @Summary Get Current User
// @Description Return current authenticated user profile.
// @Tags profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Envelope{data=auth.User}
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /profile/me [get]
func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	user, err := h.service.GetMe(r.Context(), userID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteData(w, http.StatusOK, user)
}

// patchProfile godoc
// @Summary Update Profile
// @Description Partially update current user profile.
// @Tags profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body PatchProfileRequest true "Profile patch request"
// @Success 200 {object} response.Envelope{data=auth.User}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 409 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /profile/me [patch]
func (h *Handler) patchProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	var body PatchProfileRequest
	if err := httpx.DecodeJSONAllowEmpty(r, &body); err != nil {
		response.WriteError(w, apperrors.Wrap(apperrors.KindValidation, "invalid request body", err))
		return
	}

	updated, err := h.service.UpdateMyProfile(r.Context(), userID, auth.UpdateProfileInput{
		DisplayName: body.DisplayName,
		Email:       body.Email,
		Phone:       body.Phone,
		Username:    body.Username,
	})
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteData(w, http.StatusOK, updated)
}

// uploadAvatar godoc
// @Summary Upload Profile Avatar
// @Description Upload current user's profile image as multipart/form-data field "avatar".
// @Tags profile
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param avatar formData file true "Avatar image file"
// @Success 200 {object} response.Envelope{data=auth.User}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /profile/me/avatar [post]
func (h *Handler) uploadAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	if err := r.ParseMultipartForm(maxAvatarSizeBytes); err != nil {
		response.WriteError(w, apperrors.Wrap(apperrors.KindValidation, "failed to parse multipart form", err))
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			response.WriteError(w, apperrors.New(apperrors.KindValidation, "field 'avatar' is required"))
			return
		}
		response.WriteError(w, apperrors.Wrap(apperrors.KindValidation, "failed to read avatar file", err))
		return
	}
	defer file.Close()

	raw, err := io.ReadAll(io.LimitReader(file, maxAvatarSizeBytes+1))
	if err != nil {
		response.WriteError(w, apperrors.Wrap(apperrors.KindValidation, "failed to read avatar content", err))
		return
	}
	if int64(len(raw)) > maxAvatarSizeBytes {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, "avatar file size must be <= 5MB"))
		return
	}
	if len(raw) == 0 {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, "avatar file is empty"))
		return
	}

	contentType := http.DetectContentType(raw)
	ext, ok := avatarExtension(contentType)
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, "unsupported avatar file type"))
		return
	}
	fileName := "avatar" + ext
	if header != nil && strings.TrimSpace(header.Filename) != "" {
		fileName = header.Filename
	}

	updated, err := h.service.UploadAvatar(r.Context(), userID, fileName, contentType, raw)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteData(w, http.StatusOK, updated)
}

// changePassword godoc
// @Summary Change Password
// @Description Change current user's password.
// @Tags profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ChangePasswordRequest true "Change password request"
// @Success 200 {object} response.Envelope{data=ChangePasswordResult}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /profile/me/change-password [post]
func (h *Handler) changePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	var req ChangePasswordRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.WriteError(w, apperrors.Wrap(apperrors.KindValidation, "invalid request body", err))
		return
	}

	if err := h.service.ChangePassword(r.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteData(w, http.StatusOK, ChangePasswordResult{Success: true})
}

func avatarExtension(contentType string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "image/jpeg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	case "image/webp":
		return ".webp", true
	case "image/gif":
		return ".gif", true
	default:
		return "", false
	}
}
