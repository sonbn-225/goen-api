package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/sonbn-225/goen-api/internal/services"
)

type GlobalMarketDataStatusResponse = services.GlobalMarketDataStatus

// GetGlobalMarketDataStatus godoc
// @Summary Global market-data worker status
// @Description Returns global worker rate-limit remaining and market sync last timestamps.
// @Tags investments
// @Produce json
// @Success 200 {object} GlobalMarketDataStatusResponse
// @Failure 401 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /market-data/vnstock/status [get]
func GetGlobalMarketDataStatus(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		resp, err := d.MarketDataService.GetGlobalStatus(ctx, uid)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		var out GlobalMarketDataStatusResponse = resp
		writeJSON(w, http.StatusOK, out)
	}
}
