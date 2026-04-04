package v1

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/response"
)

type MarketDataHandler struct {
	svc interfaces.MarketDataService
}

func NewMarketDataHandler(svc interfaces.MarketDataService) *MarketDataHandler {
	return &MarketDataHandler{svc: svc}
}

func (h *MarketDataHandler) RegisterRoutes(r chi.Router, cfg *config.Config) {
	r.Get("/market-data/status", h.GetGlobalStatus)
	r.Post("/market-data/sync", h.MarketSync)
	r.Post("/market-data/refresh-symbols", h.RefreshSymbols)

	r.Get("/securities/{id}/status", h.GetSecurityStatus)
	r.Post("/securities/{id}/refresh-prices", h.RefreshPrices)
	r.Post("/securities/{id}/refresh-events", h.RefreshEvents)
}

func (h *MarketDataHandler) GetGlobalStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.svc.GetGlobalStatus(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, status)
}

func (h *MarketDataHandler) MarketSync(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	var req dto.MarketSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", nil)
		return
	}

	res, err := h.svc.EnqueueMarketSync(r.Context(), userID, req)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusAccepted, res)
}

func (h *MarketDataHandler) RefreshSymbols(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	var req dto.RefreshSymbolsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", nil)
		return
	}

	res, err := h.svc.EnqueueBySymbols(r.Context(), userID, req)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusAccepted, res)
}

func (h *MarketDataHandler) GetSecurityStatus(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	id := chi.URLParam(r, "id")
	status, err := h.svc.GetSecurityStatus(r.Context(), userID, id)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, status)
}

func (h *MarketDataHandler) RefreshPrices(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	id := chi.URLParam(r, "id")
	var req dto.RefreshPriceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", nil)
		return
	}
	req.SecurityID = id

	res, err := h.svc.EnqueueSecurityPricesDaily(r.Context(), userID, req)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusAccepted, res)
}

func (h *MarketDataHandler) RefreshEvents(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	id := chi.URLParam(r, "id")
	var req dto.RefreshEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", nil)
		return
	}
	req.SecurityID = id

	res, err := h.svc.EnqueueSecurityEvents(r.Context(), userID, req)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusAccepted, res)
}
