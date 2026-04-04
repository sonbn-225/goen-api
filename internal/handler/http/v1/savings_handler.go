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
	// Savings Instruments
	r.Route("/savings-instruments", func(r chi.Router) {
		r.Get("/", h.ListSavingsInstruments)
		r.Post("/", h.CreateSavingsInstrument)
		r.Delete("/{id}", h.DeleteSavingsInstrument)
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

// Savings Instruments Handlers
func (h *SavingsHandler) ListSavingsInstruments(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	instruments, err := h.savingsSvc.ListSavingsInstruments(r.Context(), userID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, instruments)
}

func (h *SavingsHandler) CreateSavingsInstrument(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	var req dto.CreateSavingsInstrumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", nil)
		return
	}
	instr, err := h.savingsSvc.CreateSavingsInstrument(r.Context(), userID, req)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusCreated, instr)
}

func (h *SavingsHandler) DeleteSavingsInstrument(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	id := chi.URLParam(r, "id")
	if err := h.savingsSvc.DeleteSavingsInstrument(r.Context(), userID, id); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, map[string]string{"message": "Savings instrument deleted"})
}

// Rotating Savings Handlers
func (h *SavingsHandler) ListRotatingGroups(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	groups, err := h.rotatingSvc.ListGroups(r.Context(), userID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, groups)
}

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

func (h *SavingsHandler) DeleteRotatingGroup(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	id := chi.URLParam(r, "id")
	if err := h.rotatingSvc.DeleteGroup(r.Context(), userID, id); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, map[string]string{"message": "Rotating savings group deleted"})
}

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
