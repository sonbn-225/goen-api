package v1

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/response"
)

type SavingsHandler struct {
	savingsSvc interfaces.SavingsService
	rotatingSvc interfaces.RotatingSavingsService
}

func NewSavingsHandler(savingsSvc interfaces.SavingsService, rotatingSvc interfaces.RotatingSavingsService) *SavingsHandler {
	return &SavingsHandler{
		savingsSvc:  savingsSvc,
		rotatingSvc: rotatingSvc,
	}
}

func (h *SavingsHandler) RegisterRoutes(r chi.Router, cfg *config.Config) {
	// Savings
	r.Route("/savings", func(r chi.Router) {
		r.Get("/", h.ListSavings)
		r.Post("/", h.CreateSavings)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetSavings)
			r.Patch("/", h.PatchSavings)
			r.Delete("/", h.DeleteSavings)
		})
	})

	// Rotating Savings
	r.Route("/rotating-savings", func(r chi.Router) {
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
}

// Savings Handlers
// ListSavings godoc
// @Summary List Savings
// @Description Retrieve personal savings goals/accounts for the user
// @Tags Savings
// @Produce json
// @Security BearerAuth
// @Success 200 {array} entity.Savings
// @Failure 500 {object} response.ErrorEnvelope
// @Router /savings [get]
func (h *SavingsHandler) ListSavings(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	items, err := h.savingsSvc.ListSavings(r.Context(), userID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, items)
}

// CreateSavings godoc
// @Summary Create Savings
// @Description Create a personalized savings goal or instrument
// @Tags Savings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateSavingsRequest true "Savings Creation Payload"
// @Success 201 {object} entity.Savings
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /savings [post]
func (h *SavingsHandler) CreateSavings(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	var req dto.CreateSavingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", nil)
		return
	}
	item, err := h.savingsSvc.CreateSavings(r.Context(), userID, req)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusCreated, item)
}

// GetSavings godoc
// @Summary Get Savings
// @Description Retrieve details for a specific savings instance
// @Tags Savings
// @Produce json
// @Security BearerAuth
// @Param id path string true "Savings ID"
// @Success 200 {object} entity.Savings
// @Failure 404 {object} response.ErrorEnvelope
// @Router /savings/{id} [get]
func (h *SavingsHandler) GetSavings(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	id := chi.URLParam(r, "id")
	item, err := h.savingsSvc.GetSavings(r.Context(), userID, id)
	if err != nil {
		response.WriteError(w, http.StatusNotFound, "not_found", "Savings not found", nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, item)
}

// PatchSavings godoc
// @Summary Update Savings
// @Description Partially update specific savings settings
// @Tags Savings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Savings ID"
// @Param request body dto.PatchSavingsRequest true "Savings Update Payload"
// @Success 200 {object} entity.Savings
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /savings/{id} [patch]
func (h *SavingsHandler) PatchSavings(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	id := chi.URLParam(r, "id")
	var req dto.PatchSavingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", nil)
		return
	}
	item, err := h.savingsSvc.PatchSavings(r.Context(), userID, id, req)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, item)
}

// DeleteSavings godoc
// @Summary Delete Savings
// @Description Delete a savings goal by ID
// @Tags Savings
// @Produce json
// @Security BearerAuth
// @Param id path string true "Savings ID"
// @Success 200 {object} map[string]string
// @Failure 500 {object} response.ErrorEnvelope
// @Router /savings/{id} [delete]
func (h *SavingsHandler) DeleteSavings(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	id := chi.URLParam(r, "id")
	if err := h.savingsSvc.DeleteSavings(r.Context(), userID, id); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, map[string]string{"message": "Savings deleted"})
}

// Rotating Savings Handlers
// ListRotatingGroups godoc
// @Summary List Rotating Savings Groups
// @Description Retrieve a list of Rotating Savings (Hụi) groups
// @Tags RotatingSavings
// @Produce json
// @Security BearerAuth
// @Success 200 {array} object
// @Failure 500 {object} response.ErrorEnvelope
// @Router /rotating-savings [get]
func (h *SavingsHandler) ListRotatingGroups(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	groups, err := h.rotatingSvc.ListGroups(r.Context(), userID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, groups)
}

// CreateRotatingGroup godoc
// @Summary Create Rotating Savings Group
// @Description Define a new rotating savings (Hụi) group
// @Tags RotatingSavings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateRotatingSavingsGroupRequest true "Rotating Group Payload"
// @Success 201 {object} object
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /rotating-savings [post]
func (h *SavingsHandler) CreateRotatingGroup(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	var req dto.CreateRotatingSavingsGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", nil)
		return
	}
	group, err := h.rotatingSvc.CreateGroup(r.Context(), userID, req)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusCreated, group)
}

// GetRotatingGroupDetail godoc
// @Summary Get Rotating Savings Group
// @Description Retrieve detailed aggregation of a Rotating Savings group
// @Tags RotatingSavings
// @Produce json
// @Security BearerAuth
// @Param id path string true "Rotating Savings Group ID"
// @Success 200 {object} object
// @Failure 500 {object} response.ErrorEnvelope
// @Router /rotating-savings/{id} [get]
func (h *SavingsHandler) GetRotatingGroupDetail(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	id := chi.URLParam(r, "id")
	detail, err := h.rotatingSvc.GetGroupDetail(r.Context(), userID, id)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, detail)
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
// @Success 200 {object} object
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /rotating-savings/{id} [patch]
func (h *SavingsHandler) UpdateRotatingGroup(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	id := chi.URLParam(r, "id")
	var req dto.UpdateRotatingSavingsGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", nil)
		return
	}
	group, err := h.rotatingSvc.UpdateGroup(r.Context(), userID, id, req)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, group)
}

// DeleteRotatingGroup godoc
// @Summary Delete Rotating Savings Group
// @Description Clean up a specific rotating group structure
// @Tags RotatingSavings
// @Produce json
// @Security BearerAuth
// @Param id path string true "Rotating Savings Group ID"
// @Success 200 {object} map[string]string
// @Failure 500 {object} response.ErrorEnvelope
// @Router /rotating-savings/{id} [delete]
func (h *SavingsHandler) DeleteRotatingGroup(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	id := chi.URLParam(r, "id")
	if err := h.rotatingSvc.DeleteGroup(r.Context(), userID, id); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, map[string]string{"message": "Rotating savings group deleted"})
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
// @Success 201 {object} object
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /rotating-savings/{id}/contributions [post]
func (h *SavingsHandler) CreateContribution(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	id := chi.URLParam(r, "id")
	var req dto.RotatingSavingsContributionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", nil)
		return
	}
	contrib, err := h.rotatingSvc.CreateContribution(r.Context(), userID, id, req)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusCreated, contrib)
}

// DeleteContribution godoc
// @Summary Delete Contribution
// @Description Void a rotating savings contribution history by ID
// @Tags RotatingSavings
// @Produce json
// @Security BearerAuth
// @Param id path string true "Rotating Savings Group ID"
// @Param contributionId path string true "Contribution ID"
// @Success 200 {object} map[string]string
// @Failure 500 {object} response.ErrorEnvelope
// @Router /rotating-savings/{id}/contributions/{contributionId} [delete]
func (h *SavingsHandler) DeleteContribution(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	id := chi.URLParam(r, "id")
	contribID := chi.URLParam(r, "contributionId")
	if err := h.rotatingSvc.DeleteContribution(r.Context(), userID, id, contribID); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, map[string]string{"message": "Contribution deleted"})
}
