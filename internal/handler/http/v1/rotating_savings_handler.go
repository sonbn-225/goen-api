package v1

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/handler/middleware"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/response"
)

type RotatingSavingsHandler struct {
	rotatingSvc interfaces.RotatingSavingsService
}

func NewRotatingSavingsHandler(rotatingSvc interfaces.RotatingSavingsService) *RotatingSavingsHandler {
	return &RotatingSavingsHandler{rotatingSvc: rotatingSvc}
}

func (h *RotatingSavingsHandler) RegisterRoutes(r chi.Router, cfg *config.Config) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg))

		r.Route("/rotating-savings/groups", func(r chi.Router) {
			r.Get("/", h.ListRotatingGroups)
			r.Post("/", h.CreateRotatingGroup)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", h.GetRotatingGroupDetail)
				r.Patch("/", h.UpdateRotatingGroup)
				r.Delete("/", h.DeleteRotatingGroup)

				r.Post("/contributions", h.CreateContribution)
				r.Delete("/contributions/{contributionId}", h.DeleteContribution)
			})
		})
	})
}

// ListRotatingGroups godoc
// @Summary List Rotating Savings Groups
// @Description Retrieve a list of Rotating Savings (Hụi) groups
// @Tags RotatingSavings
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.SuccessEnvelope{data=[]dto.RotatingSavingsGroupSummary}
// @Failure 500 {object} response.ErrorEnvelope
// @Router /rotating-savings/groups [get]
func (h *RotatingSavingsHandler) ListRotatingGroups(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}
	groups, err := h.rotatingSvc.ListGroups(r.Context(), userID)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, groups)
}

// CreateRotatingGroup godoc
// @Summary Create Rotating Savings Group
// @Description Define a new rotating savings (Hụi) group
// @Tags RotatingSavings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateRotatingSavingsGroupRequest true "Rotating Group Payload"
// @Success 201 {object} response.SuccessEnvelope{data=dto.RotatingSavingsGroupResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /rotating-savings/groups [post]
func (h *RotatingSavingsHandler) CreateRotatingGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}
	var req dto.CreateRotatingSavingsGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", nil)
		return
	}
	group, err := h.rotatingSvc.CreateGroup(r.Context(), userID, req)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusCreated, group)
}

// GetRotatingGroupDetail godoc
// @Summary Get Rotating Savings Group
// @Description Retrieve detailed aggregation of a Rotating Savings group
// @Tags RotatingSavings
// @Produce json
// @Security BearerAuth
// @Param id path string true "Rotating Savings Group ID"
// @Success 200 {object} response.SuccessEnvelope{data=dto.RotatingSavingsGroupDetailResponse}
// @Failure 500 {object} response.ErrorEnvelope
// @Router /rotating-savings/groups/{id} [get]
func (h *RotatingSavingsHandler) GetRotatingGroupDetail(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}
	id := chi.URLParam(r, "id")
	detail, err := h.rotatingSvc.GetGroupDetail(r.Context(), userID, id)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, detail)
}

// UpdateRotatingGroup godoc
// @Summary Update Rotating Savings Group
// @Description Partially update metadata properties for a group
// @Tags RotatingSavings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Rotating Savings Group ID"
// @Param request body dto.UpdateRotatingSavingsGroupRequest true "Update Group Payload"
// @Success 200 {object} response.SuccessEnvelope{data=dto.RotatingSavingsGroupResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /rotating-savings/groups/{id} [patch]
func (h *RotatingSavingsHandler) UpdateRotatingGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}
	id := chi.URLParam(r, "id")
	var req dto.UpdateRotatingSavingsGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", nil)
		return
	}
	group, err := h.rotatingSvc.UpdateGroup(r.Context(), userID, id, req)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, group)
}

// DeleteRotatingGroup godoc
// @Summary Delete Rotating Savings Group
// @Description Clean up a specific rotating group structure
// @Tags RotatingSavings
// @Produce json
// @Security BearerAuth
// @Param id path string true "Rotating Savings Group ID"
// @Success 200 {object} response.SuccessEnvelope{data=map[string]string}
// @Failure 500 {object} response.ErrorEnvelope
// @Router /rotating-savings/groups/{id} [delete]
func (h *RotatingSavingsHandler) DeleteRotatingGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.rotatingSvc.DeleteGroup(r.Context(), userID, id); err != nil {
		response.WriteInternalError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, map[string]string{"message": "Rotating savings group deleted"})
}

// CreateContribution godoc
// @Summary Create Contribution
// @Description Make a contribution (bidding/paying) within a group cycle
// @Tags RotatingSavings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Rotating Savings Group ID"
// @Param request body dto.RotatingSavingsContributionRequest true "Contribution Payload"
// @Success 201 {object} response.SuccessEnvelope{data=dto.RotatingSavingsContributionResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /rotating-savings/groups/{id}/contributions [post]
func (h *RotatingSavingsHandler) CreateContribution(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}
	id := chi.URLParam(r, "id")
	var req dto.RotatingSavingsContributionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", nil)
		return
	}
	contrib, err := h.rotatingSvc.CreateContribution(r.Context(), userID, id, req)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusCreated, contrib)
}

// DeleteContribution godoc
// @Summary Delete Contribution
// @Description Void a rotating savings contribution history by ID
// @Tags RotatingSavings
// @Produce json
// @Security BearerAuth
// @Param id path string true "Rotating Savings Group ID"
// @Param contributionId path string true "Contribution ID"
// @Success 200 {object} response.SuccessEnvelope{data=map[string]string}
// @Failure 500 {object} response.ErrorEnvelope
// @Router /rotating-savings/groups/{id}/contributions/{contributionId} [delete]
func (h *RotatingSavingsHandler) DeleteContribution(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}
	id := chi.URLParam(r, "id")
	contribID := chi.URLParam(r, "contributionId")
	if err := h.rotatingSvc.DeleteContribution(r.Context(), userID, id, contribID); err != nil {
		response.WriteInternalError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, map[string]string{"message": "Contribution deleted"})
}
