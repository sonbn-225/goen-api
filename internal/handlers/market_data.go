package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/apierror"
)

type RefreshPricesDailyResponse struct {
	Stream    string `json:"stream"`
	MessageID string `json:"message_id"`
}

// RefreshSecurityPricesDaily godoc
// @Summary Enqueue refresh daily prices for security
// @Description Enqueue a vnstock job; worker will fetch OHLCV and upsert into security_price_dailies.
// @Tags investments
// @Produce json
// @Param securityId path string true "Security ID"
// @Param from query string false "From date (YYYY-MM-DD)"
// @Param to query string false "To date (YYYY-MM-DD)"
// @Param full query string false "Fetch full history (1/true). If set and from is empty, worker uses VNSTOCK_FULL_HISTORY_START_DATE."
// @Success 202 {object} RefreshPricesDailyResponse
// @Failure 400 {object} apierror.Envelope
// @Failure 401 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 503 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /securities/{securityId}/prices-daily/refresh [post]
func RefreshSecurityPricesDaily(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		securityID := chi.URLParam(r, "securityId")
		if securityID == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "securityId is required", map[string]any{"field": "securityId"})
			return
		}

		from := r.URL.Query().Get("from")
		to := r.URL.Query().Get("to")
		force := r.URL.Query().Get("force")
		full := r.URL.Query().Get("full")

		if from != "" {
			if _, err := time.Parse("2006-01-02", from); err != nil {
				apierror.Write(w, http.StatusBadRequest, "validation_error", "from is invalid", map[string]any{"field": "from"})
				return
			}
		}
		if to != "" {
			if _, err := time.Parse("2006-01-02", to); err != nil {
				apierror.Write(w, http.StatusBadRequest, "validation_error", "to is invalid", map[string]any{"field": "to"})
				return
			}
		}

		resp, err := d.MarketDataService.EnqueueSecurityPricesDaily(r.Context(), uid, securityID, force, full, from, to)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusAccepted, RefreshPricesDailyResponse{Stream: resp.Stream, MessageID: resp.MessageID})
	}
}
