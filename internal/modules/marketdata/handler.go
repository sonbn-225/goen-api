package marketdata

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/platform/httpx"
	"github.com/sonbn-225/goen-api/internal/response"
)

// Handler handles HTTP requests for market data.
type Handler struct {
	svc *Service
}

// NewHandler creates a new market data handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes registers all market data routes.
func (h *Handler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	// Security-level endpoints
	r.With(authMiddleware).Post("/securities/{securityId}/prices-daily/refresh", h.RefreshSecurityPricesDaily)
	r.With(authMiddleware).Post("/securities/{securityId}/events/refresh", h.RefreshSecurityEvents)
	r.With(authMiddleware).Get("/securities/{securityId}/market-data/status", h.GetSecurityStatus)

	// Global endpoints
	r.With(authMiddleware).Post("/market-data/vnstock/sync-all", h.RefreshMarketDataAll)
	r.With(authMiddleware).Post("/market-data/vnstock/sync-symbol/{symbol}", h.RefreshMarketDataBySymbol)
	r.With(authMiddleware).Post("/market-data/vnstock/sync-symbols", h.RefreshMarketDataBySymbols)
	r.With(authMiddleware).Get("/market-data/vnstock/status", h.GetGlobalStatus)
}

// RefreshSecurityPricesDaily handles POST /securities/{securityId}/prices-daily/refresh
// @Summary Enqueue refresh daily prices for security
// @Description Enqueue a vnstock job; worker will fetch OHLCV and upsert into security_price_dailies.
// @Tags investments
// @Produce json
// @Param securityId path string true "Security ID"
// @Param from query string false "From date (YYYY-MM-DD)"
// @Param to query string false "To date (YYYY-MM-DD)"
// @Param full query string false "Fetch full history (1/true)"
// @Success 202 {object} RefreshOneResponse
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 503 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /securities/{securityId}/prices-daily/refresh [post]
func (h *Handler) RefreshSecurityPricesDaily(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	securityID := chi.URLParam(r, "securityId")
	if securityID == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "securityId is required", map[string]any{"field": "securityId"})
		return
	}

	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	force := r.URL.Query().Get("force")
	full := r.URL.Query().Get("full")

	if from != "" {
		if _, err := time.Parse("2006-01-02", from); err != nil {
			response.WriteError(w, http.StatusBadRequest, "validation_error", "from is invalid", map[string]any{"field": "from"})
			return
		}
	}
	if to != "" {
		if _, err := time.Parse("2006-01-02", to); err != nil {
			response.WriteError(w, http.StatusBadRequest, "validation_error", "to is invalid", map[string]any{"field": "to"})
			return
		}
	}

	resp, err := h.svc.EnqueueSecurityPricesDaily(r.Context(), userID, securityID, force, full, from, to)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusAccepted, resp)
}

// RefreshSecurityEvents handles POST /securities/{securityId}/events/refresh
// @Summary Enqueue refresh security events
// @Description Enqueue a vnstock job; worker will fetch corporate actions/events.
// @Tags investments
// @Produce json
// @Param securityId path string true "Security ID"
// @Success 202 {object} RefreshOneResponse
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 503 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /securities/{securityId}/events/refresh [post]
func (h *Handler) RefreshSecurityEvents(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	securityID := chi.URLParam(r, "securityId")
	if securityID == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "securityId is required", map[string]any{"field": "securityId"})
		return
	}

	force := r.URL.Query().Get("force")
	resp, err := h.svc.EnqueueSecurityEvents(r.Context(), userID, securityID, force)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusAccepted, resp)
}

// GetSecurityStatus handles GET /securities/{securityId}/market-data/status
// @Summary Market data status for a security
// @Description Returns last sync timestamps, cooldown, and rate-limit (best-effort).
// @Tags investments
// @Produce json
// @Param securityId path string true "Security ID"
// @Success 200 {object} SecurityStatus
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 503 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /securities/{securityId}/market-data/status [get]
func (h *Handler) GetSecurityStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	securityID := chi.URLParam(r, "securityId")
	if securityID == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "securityId is required", map[string]any{"field": "securityId"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	resp, err := h.svc.GetSecurityStatus(ctx, userID, securityID)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, resp)
}

// RefreshMarketDataAll handles POST /market-data/vnstock/sync-all
// @Summary Enqueue market-wide sync (vnstock)
// @Description Enqueue a vnstock job to sync securities catalog then (optionally) fan-out daily prices and events jobs.
// @Tags investments
// @Produce json
// @Param force query string false "Bypass caching (1/true)"
// @Param full query string false "Full price history"
// @Param include_prices query string false "Include daily prices (default: 1)"
// @Param include_events query string false "Include events (default: 1)"
// @Success 202 {object} RefreshOneResponse
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 503 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /market-data/vnstock/sync-all [post]
func (h *Handler) RefreshMarketDataAll(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	force := r.URL.Query().Get("force")
	full := parseBoolDefault(r.URL.Query().Get("full"), false)
	includePrices := parseBoolDefault(r.URL.Query().Get("include_prices"), true)
	includeEvents := parseBoolDefault(r.URL.Query().Get("include_events"), true)

	resp, err := h.svc.EnqueueMarketSync(r.Context(), userID, includePrices, includeEvents, force, full)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusAccepted, resp)
}

// SyncSymbolsRequest for POST body.
type SyncSymbolsRequest struct {
	Symbols       []string `json:"symbols"`
	IncludePrices *bool    `json:"include_prices,omitempty"`
	IncludeEvents *bool    `json:"include_events,omitempty"`
	Force         *bool    `json:"force,omitempty"`
}

// RefreshMarketDataBySymbol handles POST /market-data/vnstock/sync-symbol/{symbol}
// @Summary Enqueue vnstock refresh for a symbol
// @Description Enqueue vnstock jobs for a ticker symbol.
// @Tags investments
// @Produce json
// @Param symbol path string true "Ticker symbol (e.g. FPT)"
// @Param include_prices query string false "Include daily prices (default: 1)"
// @Param include_events query string false "Include events (default: 1)"
// @Param force query string false "Bypass caching (1/true)"
// @Success 202 {object} RefreshManyResponse
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 503 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /market-data/vnstock/sync-symbol/{symbol} [post]
func (h *Handler) RefreshMarketDataBySymbol(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	symbol := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "symbol")))
	if symbol == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "symbol is required", map[string]any{"field": "symbol"})
		return
	}

	includePrices := parseBoolDefault(r.URL.Query().Get("include_prices"), true)
	includeEvents := parseBoolDefault(r.URL.Query().Get("include_events"), true)
	if !includePrices && !includeEvents {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "include_prices or include_events must be true", nil)
		return
	}

	force := r.URL.Query().Get("force")
	resp, err := h.svc.EnqueueBySymbol(r.Context(), userID, symbol, includePrices, includeEvents, force)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusAccepted, resp)
}

// RefreshMarketDataBySymbols handles POST /market-data/vnstock/sync-symbols
// @Summary Enqueue vnstock refresh for a list of symbols
// @Description Enqueue vnstock jobs for a list of ticker symbols.
// @Tags investments
// @Accept json
// @Produce json
// @Param force query string false "Bypass caching (1/true)"
// @Param body body SyncSymbolsRequest true "Symbols request"
// @Success 202 {object} RefreshManyResponse
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 503 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /market-data/vnstock/sync-symbols [post]
func (h *Handler) RefreshMarketDataBySymbols(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	var req SyncSymbolsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}

	if len(req.Symbols) == 0 {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "symbols is required", map[string]any{"field": "symbols"})
		return
	}

	includePrices := true
	includeEvents := true
	if req.IncludePrices != nil {
		includePrices = *req.IncludePrices
	}
	if req.IncludeEvents != nil {
		includeEvents = *req.IncludeEvents
	}
	if !includePrices && !includeEvents {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "include_prices or include_events must be true", nil)
		return
	}

	force := r.URL.Query().Get("force")
	if force == "" && req.Force != nil {
		if *req.Force {
			force = "1"
		} else {
			force = "0"
		}
	}

	resp, err := h.svc.EnqueueBySymbols(r.Context(), userID, req.Symbols, includePrices, includeEvents, force)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusAccepted, resp)
}

// GetGlobalStatus handles GET /market-data/vnstock/status
// @Summary Global market-data worker status
// @Description Returns global worker rate-limit remaining and market sync last timestamps.
// @Tags investments
// @Produce json
// @Success 200 {object} GlobalStatus
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /market-data/vnstock/status [get]
func (h *Handler) GetGlobalStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	resp, err := h.svc.GetGlobalStatus(ctx, userID)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, resp)
}

func parseBoolDefault(raw string, def bool) bool {
	v := strings.TrimSpace(strings.ToLower(raw))
	if v == "" {
		return def
	}
	switch v {
	case "1", "true", "yes", "y", "force":
		return true
	case "0", "false", "no", "n":
		return false
	default:
		return def
	}
}

