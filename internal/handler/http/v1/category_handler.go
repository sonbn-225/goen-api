package v1
 
import (
	"net/http"
 
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/handler/middleware"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/response"
	"github.com/sonbn-225/goen-api/internal/pkg/apperr"
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
 
		r.Route("/categories", func(r chi.Router) {
			r.Get("/", h.List)
			r.Get("/{categoryId}", h.Get)
		})
	})
}
 
// List godoc
// @Summary List Categories
// @Description Retrieve all categories, optionally filtered by transaction type
// @Tags Categories
// @Produce json
// @Security BearerAuth
// @Param type query string false "Transaction type filter (INCOME/EXPENSE)"
// @Success 200 {object} response.SuccessEnvelope{data=[]dto.CategoryResponse}
// @Failure 401 {object} response.ErrorEnvelope
// @Router /categories [get]
func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}
 
	txType := r.URL.Query().Get("type")
	items, err := h.svc.List(r.Context(), userID, txType)
	if err != nil {
		response.HandleError(w, err)
		return
	}
 
	response.WriteSuccess(w, http.StatusOK, items)
}
 
// Get godoc
// @Summary Get Category
// @Description Retrieve details of a specific category by ID
// @Tags Categories
// @Produce json
// @Security BearerAuth
// @Param categoryId path string true "Category ID"
// @Success 200 {object} response.SuccessEnvelope{data=dto.CategoryResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Router /categories/{categoryId} [get]
func (h *CategoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}
 
	id, err := uuid.Parse(chi.URLParam(r, "categoryId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid category id format", nil)
		return
	}
 
	item, err := h.svc.Get(r.Context(), userID, id)
	if err != nil {
		response.HandleError(w, err)
		return
	}
 
	if item == nil {
		response.WriteError(w, http.StatusNotFound, "not_found", "category not found", nil)
		return
	}
 
	response.WriteSuccess(w, http.StatusOK, item)
}
