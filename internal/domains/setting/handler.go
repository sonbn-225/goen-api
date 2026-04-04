package setting

import (
	"net/http"

	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/core/httpx"
	"github.com/sonbn-225/goen-api-v2/internal/core/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// patchSettings godoc
// @Summary Update Settings
// @Description Partially update current user settings map.
// @Tags settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body PatchSettingsRequest true "Settings patch request"
// @Success 200 {object} response.Envelope{data=UserResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /settings/me [patch]
func (h *Handler) patchSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	patch := map[string]any{}
	if err := httpx.DecodeJSONAllowEmpty(r, &patch); err != nil {
		response.WriteError(w, apperrors.Wrap(apperrors.KindValidation, "invalid request body", err))
		return
	}

	updated, err := h.service.UpdateMySettings(r.Context(), userID, patch)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteData(w, http.StatusOK, updated)
}
