package v1

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	_ "github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/handler/middleware"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/response"
	"github.com/sonbn-225/goen-api/internal/pkg/apperr"
)

type SettingsHandler struct {
	svc interfaces.AuthService
}

func NewSettingsHandler(svc interfaces.AuthService) *SettingsHandler {
	return &SettingsHandler{svc: svc}
}

func (h *SettingsHandler) RegisterRoutes(r chi.Router, cfg *config.Config) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg))

		r.Route("/settings/me", func(r chi.Router) {
			r.Patch("/", h.PatchMySettings)
		})
	})
}

// PatchMySettings godoc
// @Summary Update User Settings
// @Description Update the settings for the currently authenticated user
// @Tags Settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]interface{} true "Settings mappings"
// @Success 200 {object} response.SuccessEnvelope{data=dto.UserResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /settings/me [patch]
func (h *SettingsHandler) PatchMySettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("unauthorized", "user not found in context"))
		return
	}

	var patch map[string]any
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "failed to decode request", nil)
		return
	}

	res, err := h.svc.UpdateMySettings(r.Context(), userID, patch)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.WriteSuccess(w, http.StatusOK, res)
}
