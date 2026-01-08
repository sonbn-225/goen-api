package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/apierror"
	"github.com/sonbn-225/goen-api/internal/auth"
	"github.com/sonbn-225/goen-api/internal/domain"
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
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}
		if d.Redis == nil {
			apierror.Write(w, http.StatusServiceUnavailable, "dependency_unavailable", "redis is not configured", nil)
			return
		}

		securityID := chi.URLParam(r, "securityId")
		if securityID == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "securityId is required", map[string]any{"field": "securityId"})
			return
		}

		// Validate security exists (also normalizes errors)
		if _, err := d.InvestmentService.GetSecurity(r.Context(), uid, securityID); err != nil {
			if err == domain.ErrSecurityNotFound {
				apierror.Write(w, http.StatusNotFound, "not_found", "security not found", nil)
				return
			}
			apierror.Write(w, http.StatusBadRequest, "validation_error", err.Error(), nil)
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

		stream := "goen:market_data:jobs"
		values := map[string]any{
			"job_type":             "vnstock.prices_daily",
			"security_id":          securityID,
			"requested_by_user_id": uid,
		}
		if force != "" {
			values["force"] = force
		}
		if full != "" {
			values["full"] = full
		}
		if from != "" {
			values["from"] = from
		}
		if to != "" {
			values["to"] = to
		}

		id, err := d.Redis.XAdd(r.Context(), stream, values)
		if err != nil {
			apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}

		writeJSON(w, http.StatusAccepted, RefreshPricesDailyResponse{Stream: stream, MessageID: id})
	}
}
