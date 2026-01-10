package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/apierror"
)

// ListCategories godoc
// @Summary List categories
// @Description List categories accessible to current user (includes global default categories).
// @Tags categories
// @Produce json
// @Success 200 {array} domain.Category
// @Failure 401 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /categories [get]
func ListCategories(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		items, err := d.CategoryService.List(r.Context(), uid)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, items)
	}
}

// GetCategory godoc
// @Summary Get category
// @Description Get a single category owned by current user.
// @Tags categories
// @Produce json
// @Param categoryId path string true "Category ID"
// @Success 200 {object} domain.Category
// @Failure 401 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /categories/{categoryId} [get]
func GetCategory(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		id := chi.URLParam(r, "categoryId")
		if id == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "categoryId is required", map[string]any{"field": "categoryId"})
			return
		}

		c, err := d.CategoryService.Get(r.Context(), uid, id)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, c)
	}
}
