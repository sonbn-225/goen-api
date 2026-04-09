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

type TradeHandler struct {
	svc interfaces.TradeService
}

func NewTradeHandler(svc interfaces.TradeService) *TradeHandler {
	return &TradeHandler{svc: svc}
}

func (h *TradeHandler) RegisterRoutes(r chi.Router, cfg *config.Config) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg))

		r.Route("/trades/accounts/{id}", func(r chi.Router) {
			r.Get("/", h.ListTrades)
			r.Post("/", h.CreateTrade)
			r.Patch("/{tradeId}", h.UpdateTrade)
			r.Delete("/{tradeId}", h.DeleteTrade)
		})
	})
}

// ListTrades godoc
// @Summary List Trades
// @Description Retrieve trades associated with an investment account
// @Tags Trades
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Success 200 {object} response.SuccessEnvelope{data=[]dto.TradeResponse}
// @Failure 500 {object} response.ErrorEnvelope
// @Router /trades/accounts/{id} [get]
func (h *TradeHandler) ListTrades(w http.ResponseWriter, r *http.Request) {
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
// @Tags Trades
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Param request body dto.CreateTradeRequest true "Trade Payload"
// @Success 201 {object} response.SuccessEnvelope{data=dto.TradeResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /trades/accounts/{id} [post]
func (h *TradeHandler) CreateTrade(w http.ResponseWriter, r *http.Request) {
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
// @Tags Trades
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Param tradeId path string true "Trade ID"
// @Param request body dto.CreateTradeRequest true "Trade Payload"
// @Success 200 {object} response.SuccessEnvelope{data=dto.TradeResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /trades/accounts/{id}/{tradeId} [patch]
func (h *TradeHandler) UpdateTrade(w http.ResponseWriter, r *http.Request) {
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
// @Tags Trades
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Param tradeId path string true "Trade ID"
// @Success 200 {object} response.SuccessEnvelope{data=map[string]string}
// @Failure 500 {object} response.ErrorEnvelope
// @Router /trades/accounts/{id}/{tradeId} [delete]
func (h *TradeHandler) DeleteTrade(w http.ResponseWriter, r *http.Request) {
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
