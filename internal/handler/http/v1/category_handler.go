package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/handler/middleware"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/response"
)

type CategoryHandler struct {
	svc interfaces.CategoryService
}

func NewCategoryHandler(svc interfaces.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

func (h *CategoryHandler) RegisterRoutes(r chi.Router, cfg *config.Config) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg))
		r.Get("/categories", h.List)
		r.Get("/categories/{categoryId}", h.Get)
	})
}

// List godoc
// @Summary List Categories
// @Description Retrieve all categories, optionally filtered by transaction type
// @Tags Categories
// @Produce json
// @Security BearerAuth
// @Param type query string false "Transaction type filter (INCOME/EXPENSE)"
// @Success 200 {array} entity.Category
// @Failure 401 {object} response.ErrorEnvelope
// @Router /categories [get]
func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	txType := r.URL.Query().Get("type")
	items, err := h.svc.List(r.Context(), userID, txType)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, items)
}

// Get godoc
// @Summary Get Category
// @Description Retrieve details of a specific category by ID
// @Tags Categories
// @Produce json
// @Security BearerAuth
// @Param categoryId path string true "Category ID"
// @Success 200 {object} entity.Category
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Router /categories/{categoryId} [get]
func (h *CategoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	id := chi.URLParam(r, "categoryId")
	if id == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "category ID is required", nil)
		return
	}

	item, err := h.svc.Get(r.Context(), userID, id)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, item)
}
