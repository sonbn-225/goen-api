package handlers

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/apierror"
)

type RefreshMarketDataManyResponse struct {
	Stream     string   `json:"stream"`
	Enqueued   int      `json:"enqueued"`
	MessageIDs []string `json:"message_ids"`
	NotFound   []string `json:"not_found_symbols,omitempty"`
}

type VnstockSyncSymbolsRequest struct {
	Symbols       []string `json:"symbols"`
	IncludePrices *bool    `json:"include_prices,omitempty"`
	IncludeEvents *bool    `json:"include_events,omitempty"`
	Force         *bool    `json:"force,omitempty"`
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

// RefreshMarketDataBySymbol godoc
// @Summary Enqueue vnstock refresh for a symbol
// @Description Enqueue vnstock jobs for a ticker symbol (prices/events). Requires the symbol to exist in securities catalog.
// @Tags investments
// @Produce json
// @Param symbol path string true "Ticker symbol (e.g. FPT)"
// @Param include_prices query string false "Include daily prices (default: 1)"
// @Param include_events query string false "Include events (default: 1)"
// @Param force query string false "Bypass caching (1/true)"
// @Success 202 {object} RefreshMarketDataManyResponse
// @Failure 400 {object} apierror.Envelope
// @Failure 401 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 503 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /market-data/vnstock/sync-symbol/{symbol} [post]
func RefreshMarketDataBySymbol(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		symbol := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "symbol")))
		if symbol == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "symbol is required", map[string]any{"field": "symbol"})
			return
		}

		includePrices := parseBoolDefault(r.URL.Query().Get("include_prices"), true)
		includeEvents := parseBoolDefault(r.URL.Query().Get("include_events"), true)
		if !includePrices && !includeEvents {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "include_prices or include_events must be true", nil)
			return
		}
		force := r.URL.Query().Get("force")
		resp, err := d.MarketDataService.EnqueueBySymbol(r.Context(), uid, symbol, includePrices, includeEvents, force)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusAccepted, RefreshMarketDataManyResponse{Stream: resp.Stream, Enqueued: resp.Enqueued, MessageIDs: resp.MessageIDs})
	}
}

// RefreshMarketDataBySymbols godoc
// @Summary Enqueue vnstock refresh for a list of symbols
// @Description Enqueue vnstock jobs for a list of ticker symbols. Requires symbols to exist in securities catalog.
// @Tags investments
// @Accept json
// @Produce json
// @Param force query string false "Bypass caching (1/true)"
// @Param body body VnstockSyncSymbolsRequest true "Symbols request"
// @Success 202 {object} RefreshMarketDataManyResponse
// @Failure 400 {object} apierror.Envelope
// @Failure 401 {object} apierror.Envelope
// @Failure 503 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /market-data/vnstock/sync-symbols [post]
func RefreshMarketDataBySymbols(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		var req VnstockSyncSymbolsRequest
		if ok := decodeJSON(w, r, &req); !ok {
			return
		}

		if len(req.Symbols) == 0 {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "symbols is required", map[string]any{"field": "symbols"})
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
			apierror.Write(w, http.StatusBadRequest, "validation_error", "include_prices or include_events must be true", nil)
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

		resp, err := d.MarketDataService.EnqueueBySymbols(r.Context(), uid, req.Symbols, includePrices, includeEvents, force)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusAccepted, RefreshMarketDataManyResponse{Stream: resp.Stream, Enqueued: resp.Enqueued, MessageIDs: resp.MessageIDs, NotFound: resp.NotFound})
	}
}
