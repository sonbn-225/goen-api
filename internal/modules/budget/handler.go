package budget

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/platform/httpx"
	"github.com/sonbn-225/goen-api/internal/response"
)

// Handler handles HTTP requests for budgets.
type Handler struct {
	svc *Service
}

// NewHandler creates a new budget handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes registers all budget routes.
func (h *Handler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.With(authMiddleware).Get("/budgets", h.List)
	r.With(authMiddleware).Get("/budgets/", h.List)
	r.With(authMiddleware).Post("/budgets", h.Create)
	r.With(authMiddleware).Post("/budgets/", h.Create)
	r.With(authMiddleware).Get("/budgets/{budgetId}", h.Get)
}

// List handles GET /budgets
// @Summary List budgets
// @Description List budgets owned by current user.
// @Tags budgets
// @Produce json
// @Success 200 {array} WithStats
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /budgets [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	items, err := h.svc.List(r.Context(), userID)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, items)
}

// Create handles POST /budgets
// @Summary Create budget
// @Description Create a new budget.
// @Tags budgets
// @Accept json
// @Produce json
// @Param body body CreateRequest true "Create budget request"
// @Success 200 {object} WithStats
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /budgets [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}

	b, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, b)
}

// Get handles GET /budgets/{budgetId}
// @Summary Get budget
// @Description Get a single budget.
// @Tags budgets
// @Produce json
// @Param budgetId path string true "Budget ID"
// @Success 200 {object} WithStats
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /budgets/{budgetId} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	id := chi.URLParam(r, "budgetId")
	if id == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "budgetId is required", map[string]any{"field": "budgetId"})
		return
	}

	b, err := h.svc.Get(r.Context(), userID, id)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, b)
}

