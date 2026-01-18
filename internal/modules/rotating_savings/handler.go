package rotatingsavings

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/apperrors"
	"github.com/sonbn-225/goen-api/internal/httpapi"
	"github.com/sonbn-225/goen-api/internal/response"
)

// Handler handles HTTP requests for rotating savings.
type Handler struct {
	svc *Service
}

// NewHandler creates a new rotating savings handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes registers all rotating savings routes.
func (h *Handler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.With(authMiddleware).Get("/rotating-savings/groups", h.ListGroups)
	r.With(authMiddleware).Get("/rotating-savings/groups/", h.ListGroups)
	r.With(authMiddleware).Post("/rotating-savings/groups", h.CreateGroup)
	r.With(authMiddleware).Post("/rotating-savings/groups/", h.CreateGroup)
	r.With(authMiddleware).Get("/rotating-savings/groups/{groupId}", h.GetGroup)
	r.With(authMiddleware).Get("/rotating-savings/groups/{groupId}/contributions", h.ListContributions)
	r.With(authMiddleware).Post("/rotating-savings/groups/{groupId}/contributions", h.CreateContribution)
}

// ListGroups handles GET /rotating-savings/groups
// @Summary List rotating savings groups
// @Description List rotating savings groups owned by current user.
// @Tags rotating_savings
// @Produce json
// @Success 200 {array} domain.RotatingSavingsGroup
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /rotating-savings/groups [get]
func (h *Handler) ListGroups(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	items, err := h.svc.ListGroups(r.Context(), userID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, items)
}

// CreateGroup handles POST /rotating-savings/groups
// @Summary Create rotating savings group
// @Description Create a new rotating savings group.
// @Tags rotating_savings
// @Accept json
// @Produce json
// @Param body body CreateGroupRequest true "Create group request"
// @Success 200 {object} domain.RotatingSavingsGroup
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /rotating-savings/groups [post]
func (h *Handler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	var req CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}

	item, err := h.svc.CreateGroup(r.Context(), userID, req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, item)
}

// GetGroup handles GET /rotating-savings/groups/{groupId}
// @Summary Get rotating savings group
// @Description Get a rotating savings group.
// @Tags rotating_savings
// @Produce json
// @Param groupId path string true "Group ID"
// @Success 200 {object} domain.RotatingSavingsGroup
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /rotating-savings/groups/{groupId} [get]
func (h *Handler) GetGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	id := chi.URLParam(r, "groupId")
	if id == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "groupId is required", map[string]any{"field": "groupId"})
		return
	}

	item, err := h.svc.GetGroup(r.Context(), userID, id)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, item)
}

// ListContributions handles GET /rotating-savings/groups/{groupId}/contributions
// @Summary List rotating savings contributions
// @Description List contributions/payouts for a group.
// @Tags rotating_savings
// @Produce json
// @Param groupId path string true "Group ID"
// @Success 200 {array} domain.RotatingSavingsContribution
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /rotating-savings/groups/{groupId}/contributions [get]
func (h *Handler) ListContributions(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	groupID := chi.URLParam(r, "groupId")
	if groupID == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "groupId is required", map[string]any{"field": "groupId"})
		return
	}

	items, err := h.svc.ListContributions(r.Context(), userID, groupID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, items)
}

// CreateContribution handles POST /rotating-savings/groups/{groupId}/contributions
// @Summary Create rotating savings contribution
// @Description Create a contribution/payout.
// @Tags rotating_savings
// @Accept json
// @Produce json
// @Param groupId path string true "Group ID"
// @Param body body CreateContributionRequest true "Create contribution request"
// @Success 200 {object} domain.RotatingSavingsContribution
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /rotating-savings/groups/{groupId}/contributions [post]
func (h *Handler) CreateContribution(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	groupID := chi.URLParam(r, "groupId")
	if groupID == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "groupId is required", map[string]any{"field": "groupId"})
		return
	}

	var req CreateContributionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}

	item, err := h.svc.CreateContribution(r.Context(), userID, groupID, req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	var se *apperrors.Error
	if errors.As(err, &se) {
		response.WriteError(w, se.HTTPStatus(), string(se.Kind), se.Message, se.Details)
		return
	}
	response.WriteInternalError(w, err)
}
