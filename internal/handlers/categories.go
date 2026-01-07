package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/apierror"
	"github.com/sonbn-225/goen-api/internal/auth"
	"github.com/sonbn-225/goen-api/internal/domain"
	"github.com/sonbn-225/goen-api/internal/services"
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
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}

		items, err := d.CategoryService.List(r.Context(), uid)
		if err != nil {
			apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}

		writeJSON(w, http.StatusOK, items)
	}
}

// CreateCategory godoc
// @Summary Create category
// @Description Create a new category owned by current user.
// @Tags categories
// @Accept json
// @Produce json
// @Param body body services.CreateCategoryRequest true "Create category request"
// @Success 200 {object} domain.Category
// @Failure 400 {object} apierror.Envelope
// @Failure 401 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /categories [post]
func CreateCategory(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}

		var req services.CreateCategoryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "invalid json", nil)
			return
		}

		c, err := d.CategoryService.Create(r.Context(), uid, req)
		if err != nil {
			if errors.Is(err, domain.ErrCategoryNotFound) {
				apierror.Write(w, http.StatusNotFound, "not_found", "category not found", nil)
				return
			}
			apierror.Write(w, http.StatusBadRequest, "validation_error", err.Error(), nil)
			return
		}

		writeJSON(w, http.StatusOK, c)
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
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}

		id := chi.URLParam(r, "categoryId")
		if id == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "categoryId is required", map[string]any{"field": "categoryId"})
			return
		}

		c, err := d.CategoryService.Get(r.Context(), uid, id)
		if err != nil {
			if errors.Is(err, domain.ErrCategoryNotFound) {
				apierror.Write(w, http.StatusNotFound, "not_found", "category not found", nil)
				return
			}
			apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}

		writeJSON(w, http.StatusOK, c)
	}
}
