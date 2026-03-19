package category

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/platform/httpx"
	"github.com/sonbn-225/goen-api/internal/response"
)

// Handler handles HTTP requests for categories.
type Handler struct {
	svc *Service
}

// NewHandler creates a new category handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes registers all category routes.
func (h *Handler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.With(authMiddleware).Get("/categories", h.List)
	r.With(authMiddleware).Get("/categories/", h.List)
	r.With(authMiddleware).Get("/categories/{categoryId}", h.Get)
}

// List handles GET /categories
// @Summary List categories
// @Description List categories accessible to current user. Optionally filter by transaction type.
// @Tags categories
// @Produce json
// @Param type query string false "Filter by transaction type (income, expense, both)"
// @Success 200 {array} domain.Category
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /categories [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	txType := r.URL.Query().Get("type")
	items, err := h.svc.List(r.Context(), userID, txType)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, items)
}

// Get handles GET /categories/{categoryId}
// @Summary Get category
// @Description Get a single category.
// @Tags categories
// @Produce json
// @Param categoryId path string true "Category ID"
// @Success 200 {object} domain.Category
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /categories/{categoryId} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	id := chi.URLParam(r, "categoryId")
	if id == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "categoryId is required", map[string]any{"field": "categoryId"})
		return
	}

	item, err := h.svc.Get(r.Context(), userID, id)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, item)
}
