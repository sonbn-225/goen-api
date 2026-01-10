package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/apierror"
	"github.com/sonbn-225/goen-api/internal/services"
)

// ListBudgets godoc
// @Summary List budgets
// @Description List budgets owned by current user; includes computed spent/remaining.
// @Tags budgets
// @Produce json
// @Success 200 {array} services.BudgetWithStats
// @Failure 401 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /budgets [get]
func ListBudgets(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		items, err := d.BudgetService.List(r.Context(), uid)
		if err != nil {
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, items)
	}
}

// CreateBudget godoc
// @Summary Create budget
// @Description Create a new budget owned by current user.
// @Tags budgets
// @Accept json
// @Produce json
// @Param body body services.CreateBudgetRequest true "Create budget request"
// @Success 200 {object} services.BudgetWithStats
// @Failure 400 {object} apierror.Envelope
// @Failure 401 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /budgets [post]
func CreateBudget(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		var req services.CreateBudgetRequest
		if ok := decodeJSON(w, r, &req); !ok {
			return
		}

		b, err := d.BudgetService.Create(r.Context(), uid, req)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, b)
	}
}

// GetBudget godoc
// @Summary Get budget
// @Description Get a single budget owned by current user; includes computed spent/remaining.
// @Tags budgets
// @Produce json
// @Param budgetId path string true "Budget ID"
// @Success 200 {object} services.BudgetWithStats
// @Failure 401 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /budgets/{budgetId} [get]
func GetBudget(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		id := chi.URLParam(r, "budgetId")
		if id == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "budgetId is required", map[string]any{"field": "budgetId"})
			return
		}

		b, err := d.BudgetService.Get(r.Context(), uid, id)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, b)
	}
}
