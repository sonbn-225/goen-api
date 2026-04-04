package category

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/core/httpx"
	"github.com/sonbn-225/goen-api-v2/internal/core/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// list godoc
// @Summary List Categories
// @Description List categories accessible to the current authenticated user.
// @Tags categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param type query string false "Filter by transaction type (income, expense, both)"
// @Success 200 {object} response.Envelope{data=[]Category,meta=response.Meta}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /categories [get]
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	txType := r.URL.Query().Get("type")
	items, err := h.service.List(r.Context(), userID, txType)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteList(w, http.StatusOK, items, response.Meta{Total: len(items)})
}

// get godoc
// @Summary Get Category
// @Description Get category details by category id for current authenticated user.
// @Tags categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param categoryId path string true "Category ID"
// @Success 200 {object} response.Envelope{data=Category}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /categories/{categoryId} [get]
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	categoryID := chi.URLParam(r, "categoryId")
	if categoryID == "" {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, "categoryId is required"))
		return
	}

	item, err := h.service.Get(r.Context(), userID, categoryID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteData(w, http.StatusOK, item)
}
