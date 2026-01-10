package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/apierror"
	"github.com/sonbn-225/goen-api/internal/services"
)

type SecurityMarketDataStatusResponse = services.SecurityMarketDataStatus

// GetSecurityMarketDataStatus godoc
// @Summary Market data status for a security
// @Description Returns last sync timestamps, cooldown to next allowed sync, and current worker rate-limit remaining (best-effort).
// @Tags investments
// @Produce json
// @Param securityId path string true "Security ID"
// @Success 200 {object} SecurityMarketDataStatusResponse
// @Failure 401 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 503 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /securities/{securityId}/market-data/status [get]
func GetSecurityMarketDataStatus(d Deps) http.HandlerFunc {
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

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		resp, err := d.MarketDataService.GetSecurityStatus(ctx, uid, securityID)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		var out SecurityMarketDataStatusResponse = resp
		writeJSON(w, http.StatusOK, out)
	}
}
