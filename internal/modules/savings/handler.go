package savings

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/httpapi"
	"github.com/sonbn-225/goen-api/internal/response"
	"github.com/sonbn-225/goen-api/internal/apperrors"
)

// Handler handles HTTP requests for savings.
type Handler struct {
	svc *Service
}

// NewHandler creates a new savings handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes registers all savings routes.
func (h *Handler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.With(authMiddleware).Get("/savings/instruments", h.List)
	r.With(authMiddleware).Get("/savings/instruments/", h.List)
	r.With(authMiddleware).Post("/savings/instruments", h.Create)
	r.With(authMiddleware).Post("/savings/instruments/", h.Create)
	r.With(authMiddleware).Get("/savings/instruments/{instrumentId}", h.Get)
	r.With(authMiddleware).Patch("/savings/instruments/{instrumentId}", h.Patch)
	r.With(authMiddleware).Delete("/savings/instruments/{instrumentId}", h.Delete)
}

// List handles GET /savings/instruments
// @Summary List savings instruments
// @Description List savings instruments accessible to current user.
// @Tags savings
// @Produce json
// @Success 200 {array} domain.SavingsInstrument
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /savings/instruments [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	items, err := h.svc.List(r.Context(), userID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, items)
}

// Create handles POST /savings/instruments
// @Summary Create savings instrument
// @Description Create a new savings instrument.
// @Tags savings
// @Accept json
// @Produce json
// @Param body body CreateRequest true "Create savings instrument request"
// @Success 200 {object} domain.SavingsInstrument
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /savings/instruments [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}

	item, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, item)
}

// Get handles GET /savings/instruments/{instrumentId}
// @Summary Get savings instrument
// @Description Get a single savings instrument.
// @Tags savings
// @Produce json
// @Param instrumentId path string true "Savings instrument ID"
// @Success 200 {object} domain.SavingsInstrument
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /savings/instruments/{instrumentId} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	id := chi.URLParam(r, "instrumentId")
	if id == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "instrumentId is required", map[string]any{"field": "instrumentId"})
		return
	}

	item, err := h.svc.Get(r.Context(), userID, id)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, item)
}

// Patch handles PATCH /savings/instruments/{instrumentId}
// @Summary Patch savings instrument
// @Description Patch a savings instrument.
// @Tags savings
// @Accept json
// @Produce json
// @Param instrumentId path string true "Savings instrument ID"
// @Param body body PatchRequest true "Patch savings instrument request"
// @Success 200 {object} domain.SavingsInstrument
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /savings/instruments/{instrumentId} [patch]
func (h *Handler) Patch(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	id := chi.URLParam(r, "instrumentId")
	if id == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "instrumentId is required", map[string]any{"field": "instrumentId"})
		return
	}

	var req PatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}

	item, err := h.svc.Patch(r.Context(), userID, id, req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, item)
}

// Delete handles DELETE /savings/instruments/{instrumentId}
// @Summary Delete savings instrument
// @Description Delete a savings instrument.
// @Tags savings
// @Param instrumentId path string true "Savings instrument ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /savings/instruments/{instrumentId} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	id := chi.URLParam(r, "instrumentId")
	if id == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "instrumentId is required", map[string]any{"field": "instrumentId"})
		return
	}

	if err := h.svc.Delete(r.Context(), userID, id); err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]string{"deleted": id})
}

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	var se *apperrors.Error
	if errors.As(err, &se) {
		response.WriteError(w, se.HTTPStatus(), string(se.Kind), se.Message, se.Details)
		return
	}
	response.WriteInternalError(w, err)
}
