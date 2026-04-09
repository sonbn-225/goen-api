package v1
 
import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
 
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/handler/middleware"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/response"
	"github.com/sonbn-225/goen-api/internal/pkg/apperr"
)
 
type MarketDataHandler struct {
	svc interfaces.MarketDataService
}
 
func NewMarketDataHandler(svc interfaces.MarketDataService) *MarketDataHandler {
	return &MarketDataHandler{svc: svc}
}
 
func (h *MarketDataHandler) RegisterRoutes(r chi.Router, cfg *config.Config) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg))
 
		r.Route("/investments/market-data", func(r chi.Router) {
			r.Get("/status", h.GetGlobalStatus)
			r.Post("/sync", h.MarketSync)
			r.Post("/sync-catalog", h.SyncCatalog)
			r.Post("/sync-symbols", h.RefreshSymbols)
		})
 
		r.Route("/investments/securities/{id}", func(r chi.Router) {
			r.Get("/status", h.GetSecurityStatus)
			r.Post("/prices-daily/refresh", h.RefreshPrices)
			r.Post("/events/refresh", h.RefreshEvents)
		})
	})
}
 
// GetGlobalStatus godoc
// @Summary Get Global Market Data Status
// @Description Retrieve the status of the global background market data sync routines
// @Tags MarketData
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.SuccessEnvelope{data=object}
// @Failure 500 {object} response.ErrorEnvelope
// @Router /investments/market-data/status [get]
func (h *MarketDataHandler) GetGlobalStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.svc.GetGlobalStatus(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteSuccess(w, http.StatusOK, status)
}
 
// MarketSync godoc
// @Summary Enqueue Global Market Sync
// @Description Trigger a global sync mapping all active investments to real-time prices
// @Tags MarketData
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.MarketSyncRequest true "Market Sync Payload"
// @Success 202 {object} response.SuccessEnvelope{data=object}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /investments/market-data/sync [post]
func (h *MarketDataHandler) MarketSync(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("unauthorized", "user not found in context"))
		return
	}
	var req dto.MarketSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", nil)
		return
	}
 
	if includePrices, ok := queryBool(r, "prices"); ok {
		req.IncludePrices = includePrices
	}
	if includeEvents, ok := queryBool(r, "events"); ok {
		req.IncludeEvents = includeEvents
	}
	if !req.IncludePrices && !req.IncludeEvents {
		req.IncludePrices = true
		req.IncludeEvents = true
	}
	if force := strings.TrimSpace(r.URL.Query().Get("force")); force != "" {
		req.Force = &force
	}
	if full, ok := queryBool(r, "full"); ok {
		req.Full = full
	}
 
	res, err := h.svc.EnqueueMarketSync(r.Context(), userID, req)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteSuccess(w, http.StatusAccepted, res)
}
 
func (h *MarketDataHandler) SyncCatalog(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("unauthorized", "user not found in context"))
		return
	}
 
	req := dto.MarketSyncRequest{
		IncludePrices: true,
		IncludeEvents: true,
	}
	if force := strings.TrimSpace(r.URL.Query().Get("force")); force != "" {
		req.Force = &force
	}
 
	res, err := h.svc.EnqueueMarketSync(r.Context(), userID, req)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteSuccess(w, http.StatusAccepted, res)
}
 
// RefreshSymbols godoc
// @Summary Refresh Specific Symbols
// @Description Trigger a refresh for distinct market symbols immediately
// @Tags MarketData
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.RefreshSymbolsRequest true "Symbols payload"
// @Success 202 {object} response.SuccessEnvelope{data=object}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /investments/market-data/sync-symbols [post]
func (h *MarketDataHandler) RefreshSymbols(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("unauthorized", "user not found in context"))
		return
	}
	var req dto.RefreshSymbolsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", nil)
		return
	}
	if includePrices, ok := queryBool(r, "prices"); ok {
		req.IncludePrices = includePrices
	}
	if includeEvents, ok := queryBool(r, "events"); ok {
		req.IncludeEvents = includeEvents
	}
	if !req.IncludePrices && !req.IncludeEvents {
		req.IncludePrices = true
		req.IncludeEvents = true
	}
	if force := strings.TrimSpace(r.URL.Query().Get("force")); force != "" {
		req.Force = &force
	}
 
	res, err := h.svc.EnqueueBySymbols(r.Context(), userID, req)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteSuccess(w, http.StatusAccepted, res)
}
 
// GetSecurityStatus godoc
// @Summary Get Security Sync Status
// @Description Real-time sync status for a distinct security ID
// @Tags MarketData
// @Produce json
// @Security BearerAuth
// @Param id path string true "Security ID"
// @Success 200 {object} response.SuccessEnvelope{data=object}
// @Failure 500 {object} response.ErrorEnvelope
// @Router /investments/securities/{id}/status [get]
func (h *MarketDataHandler) GetSecurityStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("unauthorized", "user not found in context"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid security id format", nil)
		return
	}
	status, err := h.svc.GetSecurityStatus(r.Context(), userID, id)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteSuccess(w, http.StatusOK, status)
}
 
// RefreshPrices godoc
// @Summary Refresh Security Prices
// @Description Manually queue a price refresh from provider for a specific security
// @Tags MarketData
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Security ID"
// @Param request body dto.RefreshPriceRequest true "Refresh Payload"
// @Success 202 {object} response.SuccessEnvelope{data=object}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /investments/securities/{id}/prices-daily/refresh [post]
func (h *MarketDataHandler) RefreshPrices(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("unauthorized", "user not found in context"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid security id format", nil)
		return
	}
	var req dto.RefreshPriceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", nil)
		return
	}
	req.SecurityID = id
	if force := strings.TrimSpace(r.URL.Query().Get("force")); force != "" {
		req.Force = &force
	}
	if full := strings.TrimSpace(r.URL.Query().Get("full")); full != "" {
		req.Full = &full
	}
	if from := strings.TrimSpace(r.URL.Query().Get("from")); from != "" {
		req.From = &from
	}
	if to := strings.TrimSpace(r.URL.Query().Get("to")); to != "" {
		req.To = &to
	}
 
	res, err := h.svc.EnqueueSecurityPricesDaily(r.Context(), userID, req)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteSuccess(w, http.StatusAccepted, res)
}
 
// RefreshEvents godoc
// @Summary Refresh Security Events
// @Description Manually queue an event updates fetch (dividends, splits) for a security
// @Tags MarketData
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Security ID"
// @Param request body dto.RefreshEventRequest true "Refresh Event Payload"
// @Success 202 {object} response.SuccessEnvelope{data=object}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /investments/securities/{id}/events/refresh [post]
func (h *MarketDataHandler) RefreshEvents(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("unauthorized", "user not found in context"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid security id format", nil)
		return
	}
	var req dto.RefreshEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", nil)
		return
	}
	req.SecurityID = id
	if force := strings.TrimSpace(r.URL.Query().Get("force")); force != "" {
		req.Force = &force
	}
 
	res, err := h.svc.EnqueueSecurityEvents(r.Context(), userID, req)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteSuccess(w, http.StatusAccepted, res)
}
 
func queryBool(r *http.Request, key string) (bool, bool) {
	v := strings.TrimSpace(strings.ToLower(r.URL.Query().Get(key)))
	if v == "" {
		return false, false
	}
	return v == "1" || v == "true" || v == "yes" || v == "on", true
}
