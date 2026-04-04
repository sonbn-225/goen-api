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

type SavingsHandler struct {
	savingsSvc interfaces.SavingsService
}

func NewSavingsHandler(savingsSvc interfaces.SavingsService) *SavingsHandler {
	return &SavingsHandler{
		savingsSvc: savingsSvc,
	}
}

func (h *SavingsHandler) RegisterRoutes(r chi.Router, cfg *config.Config) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg))

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
	})
}

// Savings Handlers
// ListSavings godoc
// @Summary List Savings
// @Description Retrieve personal savings goals/accounts for the user
// @Tags Savings
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.SuccessEnvelope{data=[]dto.SavingsResponse}
// @Failure 500 {object} response.ErrorEnvelope
// @Router /savings [get]
func (h *SavingsHandler) ListSavings(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}
	items, err := h.savingsSvc.ListSavings(r.Context(), userID)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, items)
}

// CreateSavings godoc
// @Summary Create Savings
// @Description Create a personalized savings goal or instrument
// @Tags Savings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateSavingsRequest true "Savings Creation Payload"
// @Success 201 {object} response.SuccessEnvelope{data=dto.SavingsResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /savings [post]
func (h *SavingsHandler) CreateSavings(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}
	var req dto.CreateSavingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", nil)
		return
	}
	item, err := h.savingsSvc.CreateSavings(r.Context(), userID, req)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusCreated, item)
}

// GetSavings godoc
// @Summary Get Savings
// @Description Retrieve details for a specific savings instance
// @Tags Savings
// @Produce json
// @Security BearerAuth
// @Param id path string true "Savings ID"
// @Success 200 {object} response.SuccessEnvelope{data=dto.SavingsResponse}
// @Failure 404 {object} response.ErrorEnvelope
// @Router /savings/{id} [get]
func (h *SavingsHandler) GetSavings(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}
	id := chi.URLParam(r, "id")
	item, err := h.savingsSvc.GetSavings(r.Context(), userID, id)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, item)
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
// @Success 200 {object} response.SuccessEnvelope{data=dto.SavingsResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /savings/{id} [patch]
func (h *SavingsHandler) PatchSavings(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}
	id := chi.URLParam(r, "id")
	var req dto.PatchSavingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", nil)
		return
	}
	item, err := h.savingsSvc.PatchSavings(r.Context(), userID, id, req)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, item)
}

// DeleteSavings godoc
// @Summary Delete Savings
// @Description Delete a savings goal by ID
// @Tags Savings
// @Produce json
// @Security BearerAuth
// @Param id path string true "Savings ID"
// @Success 200 {object} response.SuccessEnvelope{data=map[string]string}
// @Failure 500 {object} response.ErrorEnvelope
// @Router /savings/{id} [delete]
func (h *SavingsHandler) DeleteSavings(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.savingsSvc.DeleteSavings(r.Context(), userID, id); err != nil {
		response.WriteInternalError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, map[string]string{"message": "Savings deleted"})
}
