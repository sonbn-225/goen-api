package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/apierror"
	"github.com/sonbn-225/goen-api/internal/auth"
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

func loadSecurityIDsBySymbols(d Deps, r *http.Request, symbols []string) (map[string]string, error) {
	if d.DB == nil {
		return nil, nil
	}
	pool, err := d.DB.Pool(r.Context())
	if err != nil {
		return nil, err
	}
	if pool == nil {
		return nil, nil
	}

	cleaned := make([]string, 0, len(symbols))
	seen := map[string]struct{}{}
	for _, s := range symbols {
		sym := strings.ToUpper(strings.TrimSpace(s))
		if sym == "" {
			continue
		}
		if _, ok := seen[sym]; ok {
			continue
		}
		seen[sym] = struct{}{}
		cleaned = append(cleaned, sym)
	}
	if len(cleaned) == 0 {
		return map[string]string{}, nil
	}

	rows, err := pool.Query(r.Context(), `
		SELECT symbol, id
		FROM securities
		WHERE symbol = ANY($1)
	`, cleaned)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string]string{}
	for rows.Next() {
		var sym, id string
		if err := rows.Scan(&sym, &id); err != nil {
			return nil, err
		}
		out[strings.ToUpper(strings.TrimSpace(sym))] = id
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
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
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}
		if d.Redis == nil {
			apierror.Write(w, http.StatusServiceUnavailable, "dependency_unavailable", "redis is not configured", nil)
			return
		}
		if d.DB == nil {
			apierror.Write(w, http.StatusServiceUnavailable, "dependency_unavailable", "postgres is not configured", nil)
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

		idsBySymbol, err := loadSecurityIDsBySymbols(d, r, []string{symbol})
		if err != nil {
			apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}
		securityID, found := idsBySymbol[symbol]
		if !found || securityID == "" {
			apierror.Write(w, http.StatusNotFound, "not_found", "security symbol not found", map[string]any{"symbol": symbol})
			return
		}

		stream := "goen:market_data:jobs"
		messageIDs := []string{}
		enqueued := 0

		if includePrices {
			values := map[string]any{
				"job_type":             "vnstock.prices_daily",
				"security_id":          securityID,
				"requested_by_user_id": uid,
			}
			if force != "" {
				values["force"] = force
			}
			id, err := d.Redis.XAdd(r.Context(), stream, values)
			if err != nil {
				apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
				return
			}
			messageIDs = append(messageIDs, id)
			enqueued++
		}

		if includeEvents {
			values := map[string]any{
				"job_type":             "vnstock.security_events",
				"security_id":          securityID,
				"requested_by_user_id": uid,
			}
			if force != "" {
				values["force"] = force
			}
			id, err := d.Redis.XAdd(r.Context(), stream, values)
			if err != nil {
				apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
				return
			}
			messageIDs = append(messageIDs, id)
			enqueued++
		}

		writeJSON(w, http.StatusAccepted, RefreshMarketDataManyResponse{Stream: stream, Enqueued: enqueued, MessageIDs: messageIDs})
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
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}
		if d.Redis == nil {
			apierror.Write(w, http.StatusServiceUnavailable, "dependency_unavailable", "redis is not configured", nil)
			return
		}
		if d.DB == nil {
			apierror.Write(w, http.StatusServiceUnavailable, "dependency_unavailable", "postgres is not configured", nil)
			return
		}

		var req VnstockSyncSymbolsRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
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

		// Resolve security IDs from symbols
		cleaned := make([]string, 0, len(req.Symbols))
		seen := map[string]struct{}{}
		for _, s := range req.Symbols {
			sym := strings.ToUpper(strings.TrimSpace(s))
			if sym == "" {
				continue
			}
			if _, ok := seen[sym]; ok {
				continue
			}
			seen[sym] = struct{}{}
			cleaned = append(cleaned, sym)
		}
		if len(cleaned) == 0 {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "symbols is required", map[string]any{"field": "symbols"})
			return
		}

		idsBySymbol, err := loadSecurityIDsBySymbols(d, r, cleaned)
		if err != nil {
			apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}

		notFound := []string{}
		stream := "goen:market_data:jobs"
		messageIDs := []string{}
		enqueued := 0

		for _, sym := range cleaned {
			securityID, ok := idsBySymbol[sym]
			if !ok || securityID == "" {
				notFound = append(notFound, sym)
				continue
			}

			if includePrices {
				values := map[string]any{
					"job_type":             "vnstock.prices_daily",
					"security_id":          securityID,
					"requested_by_user_id": uid,
				}
				if force != "" {
					values["force"] = force
				}
				id, err := d.Redis.XAdd(r.Context(), stream, values)
				if err != nil {
					apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
					return
				}
				messageIDs = append(messageIDs, id)
				enqueued++
			}

			if includeEvents {
				values := map[string]any{
					"job_type":             "vnstock.security_events",
					"security_id":          securityID,
					"requested_by_user_id": uid,
				}
				if force != "" {
					values["force"] = force
				}
				id, err := d.Redis.XAdd(r.Context(), stream, values)
				if err != nil {
					apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
					return
				}
				messageIDs = append(messageIDs, id)
				enqueued++
			}
		}

		writeJSON(w, http.StatusAccepted, RefreshMarketDataManyResponse{Stream: stream, Enqueued: enqueued, MessageIDs: messageIDs, NotFound: notFound})
	}
}
