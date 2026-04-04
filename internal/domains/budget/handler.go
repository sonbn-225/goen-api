package budget

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
// @Summary List Budgets
// @Description List budgets for current authenticated user.
// @Tags budgets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Envelope{data=[]WithStats,meta=response.Meta}
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /budgets [get]
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	items, err := h.service.List(r.Context(), userID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteList(w, http.StatusOK, items, response.Meta{Total: len(items)})
}

// create godoc
// @Summary Create Budget
// @Description Create a new budget for current authenticated user.
// @Tags budgets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateBudgetRequest true "Create budget request"
// @Success 201 {object} response.Envelope{data=WithStats}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /budgets [post]
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	var req CreateBudgetRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.WriteError(w, apperrors.Wrap(apperrors.KindValidation, "invalid request body", err))
		return
	}

	created, err := h.service.Create(r.Context(), userID, CreateInput{
		Name:                  req.Name,
		Period:                req.Period,
		PeriodStart:           req.PeriodStart,
		PeriodEnd:             req.PeriodEnd,
		Amount:                req.Amount,
		AlertThresholdPercent: req.AlertThresholdPercent,
		RolloverMode:          req.RolloverMode,
		CategoryID:            req.CategoryID,
	})
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteData(w, http.StatusCreated, created)
}

// get godoc
// @Summary Get Budget
// @Description Get budget details by budget id for current authenticated user.
// @Tags budgets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param budgetId path string true "Budget ID"
// @Success 200 {object} response.Envelope{data=WithStats}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /budgets/{budgetId} [get]
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	budgetID := chi.URLParam(r, "budgetId")
	if budgetID == "" {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, "budgetId is required"))
		return
	}

	item, err := h.service.Get(r.Context(), userID, budgetID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteData(w, http.StatusOK, item)
}
