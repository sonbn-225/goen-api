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

type BudgetHandler struct {
	svc interfaces.BudgetService
}

func NewBudgetHandler(svc interfaces.BudgetService) *BudgetHandler {
	return &BudgetHandler{svc: svc}
}

func (h *BudgetHandler) RegisterRoutes(r chi.Router, cfg *config.Config) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg))
		r.Get("/budgets", h.List)
		r.Post("/budgets", h.Create)
		r.Get("/budgets/{id}", h.Get)
		r.Patch("/budgets/{id}", h.Update)
		r.Delete("/budgets/{id}", h.Delete)
	})
}

// List godoc
// @Summary List Budgets
// @Description Retrieve a list of budgets for the current user
// @Tags Budgets
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.SuccessEnvelope{data=[]entity.Budget}
// @Failure 401 {object} response.ErrorEnvelope
// @Router /budgets [get]
func (h *BudgetHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	items, err := h.svc.List(r.Context(), userID)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, items)
}

// Create godoc
// @Summary Create Budget
// @Description Create a new budget limit for specific categories
// @Tags Budgets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateBudgetRequest true "Budget Creation Payload"
// @Success 201 {object} response.SuccessEnvelope{data=entity.Budget}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /budgets [post]
func (h *BudgetHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	var req dto.CreateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
		return
	}

	res, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusCreated, res)
}

// Get godoc
// @Summary Get Budget
// @Description Retrieve a specific budget by ID
// @Tags Budgets
// @Produce json
// @Security BearerAuth
// @Param id path string true "Budget ID"
// @Success 200 {object} response.SuccessEnvelope{data=entity.Budget}
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Router /budgets/{id} [get]
func (h *BudgetHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	id := chi.URLParam(r, "id")
	res, err := h.svc.Get(r.Context(), userID, id)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, res)
}

// Update godoc
// @Summary Update Budget
// @Description Partially update budget properties
// @Tags Budgets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Budget ID"
// @Param request body dto.UpdateBudgetRequest true "Budget Update Payload"
// @Success 200 {object} response.SuccessEnvelope{data=entity.Budget}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /budgets/{id} [patch]
func (h *BudgetHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	id := chi.URLParam(r, "id")
	var req dto.UpdateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
		return
	}

	res, err := h.svc.Update(r.Context(), userID, id, req)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, res)
}

// Delete godoc
// @Summary Delete Budget
// @Description Delete a specific budget by ID
// @Tags Budgets
// @Security BearerAuth
// @Param id path string true "Budget ID"
// @Success 204 "No Content"
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Router /budgets/{id} [delete]
func (h *BudgetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), userID, id); err != nil {
		response.WriteInternalError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
