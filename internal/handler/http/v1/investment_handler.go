package v1

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/handler/middleware"
	"github.com/sonbn-225/goen-api/internal/pkg/apperr"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/response"
)

type InvestmentHandler struct {
	svc interfaces.InvestmentService
}

func NewInvestmentHandler(svc interfaces.InvestmentService) *InvestmentHandler {
	return &InvestmentHandler{svc: svc}
}

func (h *InvestmentHandler) RegisterRoutes(r chi.Router, cfg *config.Config) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg))

		r.Route("/investments/accounts/{id}", func(r chi.Router) {
			r.Get("/holdings", h.ListHoldings)
			r.Get("/reports/realized-pnl", h.GetRealizedPNLReport)
			r.Get("/eligible-actions", h.ListEligibleActions)
			r.Post("/actions/{eventId}/claim", h.ClaimAction)
			r.Post("/backfill-cash", h.BackfillCash)
		})

	})
}

// Endpoints for Listing and Getting accounts are now part of the standard Account API.

// ListHoldings godoc
// @Summary List Holdings
// @Description Retrieve current holdings for an investment account
// @Tags Investments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Success 200 {object} response.SuccessEnvelope{data=[]dto.HoldingResponse}
// @Failure 500 {object} response.ErrorEnvelope
// @Router /investments/accounts/{id}/holdings [get]
func (h *InvestmentHandler) ListHoldings(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("unauthorized", "user not found in context"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid account id format", nil)
		return
	}
	holdings, err := h.svc.ListHoldings(r.Context(), userID, id)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, holdings)
}

// GetRealizedPNLReport godoc
// @Summary Get Realized P&L Report
// @Description Realized profit/loss report for an investment account
// @Tags Investments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Success 200 {object} response.SuccessEnvelope{data=dto.RealizedPNLReport}
// @Failure 500 {object} response.ErrorEnvelope
// @Router /investments/accounts/{id}/reports/realized-pnl [get]
func (h *InvestmentHandler) GetRealizedPNLReport(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("unauthorized", "user not found in context"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid account id format", nil)
		return
	}
	report, err := h.svc.GetRealizedPNLReport(r.Context(), userID, id)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, report)
}

func (h *InvestmentHandler) ListEligibleActions(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("unauthorized", "user not found in context"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid account id format", nil)
		return
	}
	items, err := h.svc.ListEligibleCorporateActions(r.Context(), userID, id)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, items)
}

func (h *InvestmentHandler) ClaimAction(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("unauthorized", "user not found in context"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid account id format", nil)
		return
	}
	eventID, err := uuid.Parse(chi.URLParam(r, "eventId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid event id format", nil)
		return
	}

	var req dto.ClaimCorporateActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", nil)
		return
	}

	trade, err := h.svc.ClaimCorporateAction(r.Context(), userID, id, eventID, req)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, trade)
}

func (h *InvestmentHandler) BackfillCash(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("unauthorized", "user not found in context"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid account id format", nil)
		return
	}
	result, err := h.svc.BackfillTradePrincipalTransactions(r.Context(), userID, id)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, result)
}
