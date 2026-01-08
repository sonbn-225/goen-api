package handlers

import (
	"net/http"

	"github.com/sonbn-225/goen-api/internal/apierror"
	"github.com/sonbn-225/goen-api/internal/auth"
)

func boolTo01(v bool) string {
	if v {
		return "1"
	}
	return "0"
}

type RefreshMarketDataResponse struct {
	Stream    string `json:"stream"`
	MessageID string `json:"message_id"`
}

// RefreshMarketDataAll godoc
// @Summary Enqueue market-wide sync (vnstock)
// @Description Enqueue a vnstock job; worker will sync securities catalog then fan-out daily prices and events jobs.
// @Tags investments
// @Produce json
// @Param force query string false "Bypass caching (1/true)"
// @Param full query string false "Full price history (only applies when include_prices=1)"
// @Param include_prices query string false "Include daily prices (default: 1)"
// @Param include_events query string false "Include events (default: 1)"
// @Success 202 {object} RefreshMarketDataResponse
// @Failure 401 {object} apierror.Envelope
// @Failure 503 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /market-data/vnstock/sync-all [post]
func RefreshMarketDataAll(d Deps) http.HandlerFunc {
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

		force := r.URL.Query().Get("force")
		full := parseBoolDefault(r.URL.Query().Get("full"), false)
		includePrices := parseBoolDefault(r.URL.Query().Get("include_prices"), true)
		includeEvents := parseBoolDefault(r.URL.Query().Get("include_events"), true)
		if !includePrices && !includeEvents {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "include_prices or include_events must be true", nil)
			return
		}

		stream := "goen:market_data:jobs"
		values := map[string]any{
			"job_type":             "vnstock.market_sync",
			"include_prices":       boolTo01(includePrices),
			"include_events":       boolTo01(includeEvents),
			"requested_by_user_id": uid,
		}
		if force != "" {
			values["force"] = force
		}
		if full && includePrices {
			values["full"] = "1"
		}

		id, err := d.Redis.XAdd(r.Context(), stream, values)
		if err != nil {
			apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}

		writeJSON(w, http.StatusAccepted, RefreshMarketDataResponse{Stream: stream, MessageID: id})
	}
}
