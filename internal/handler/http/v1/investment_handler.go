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
			r.Route("/trades", func(r chi.Router) {
				r.Get("/", h.ListTrades)
				r.Post("/", h.CreateTrade)
				r.Patch("/{tradeId}", h.UpdateTrade)
				r.Delete("/{tradeId}", h.DeleteTrade)
			})

			r.Get("/holdings", h.ListHoldings)
			r.Get("/reports/realized-pnl", h.GetRealizedPNLReport)
			r.Get("/eligible-actions", h.ListEligibleActions)
			r.Post("/actions/{eventId}/claim", h.ClaimAction)
			r.Post("/backfill-cash", h.BackfillCash)
		})

	})
}

// Endpoints for Listing and Getting accounts are now part of the standard Account API.

// ListTrades godoc
// @Summary List Trades
// @Description Retrieve trades associated with an investment account
// @Tags Investments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Success 200 {object} response.SuccessEnvelope{data=[]dto.TradeResponse}
// @Failure 500 {object} response.ErrorEnvelope
// @Router /investments/accounts/{id}/trades [get]
func (h *InvestmentHandler) ListTrades(w http.ResponseWriter, r *http.Request) {
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
	trades, err := h.svc.ListTrades(r.Context(), userID, id)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, trades)
}

// CreateTrade godoc
// @Summary Create Trade
// @Description Create a new trade for an investment account
// @Tags Investments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Param request body dto.CreateTradeRequest true "Trade Payload"
// @Success 201 {object} response.SuccessEnvelope{data=dto.TradeResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /investments/accounts/{id}/trades [post]
func (h *InvestmentHandler) CreateTrade(w http.ResponseWriter, r *http.Request) {
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
	var req dto.CreateTradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", nil)
		return
	}
	trade, err := h.svc.CreateTrade(r.Context(), userID, id, req)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusCreated, trade)
}

// UpdateTrade godoc
// @Summary Update Trade
// @Description Update an existing trade record
// @Tags Investments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Param tradeId path string true "Trade ID"
// @Param request body dto.CreateTradeRequest true "Trade Payload"
// @Success 200 {object} response.SuccessEnvelope{data=dto.TradeResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /investments/accounts/{id}/trades/{tradeId} [patch]
func (h *InvestmentHandler) UpdateTrade(w http.ResponseWriter, r *http.Request) {
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
	tradeID, err := uuid.Parse(chi.URLParam(r, "tradeId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid trade id format", nil)
		return
	}
	var req dto.CreateTradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", nil)
		return
	}
	trade, err := h.svc.UpdateTrade(r.Context(), userID, id, tradeID, req)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	if trade == nil {
		response.WriteError(w, http.StatusNotFound, "not_found", "trade not found", nil)
		return
	}
	response.WriteSuccess(w, http.StatusOK, trade)
}

// DeleteTrade godoc
// @Summary Delete Trade
// @Description Delete a trade from an investment account
// @Tags Investments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Param tradeId path string true "Trade ID"
// @Success 200 {object} response.SuccessEnvelope{data=map[string]string}
// @Failure 500 {object} response.ErrorEnvelope
// @Router /investments/accounts/{id}/trades/{tradeId} [delete]
func (h *InvestmentHandler) DeleteTrade(w http.ResponseWriter, r *http.Request) {
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
	tradeID, err := uuid.Parse(chi.URLParam(r, "tradeId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid trade id format", nil)
		return
	}
	err = h.svc.DeleteTrade(r.Context(), userID, id, tradeID)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, map[string]string{"message": "Trade deleted"})
}

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

