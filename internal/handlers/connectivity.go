package handlers

import (
	"context"
	"net/http"
	"time"
)

type ConnectivityItem struct {
	OK      bool           `json:"ok"`
	Details map[string]any `json:"details,omitempty"`
	Error   string         `json:"error,omitempty"`
}

type ConnectivityResponse struct {
	Postgres ConnectivityItem `json:"postgres"`
	Redis    ConnectivityItem `json:"redis"`
}

// Connectivity godoc
// @Summary Connectivity probe
// @Description Probes Postgres and Redis and returns diagnostic info.
// @Tags meta
// @Produce json
// @Success 200 {object} ConnectivityResponse
// @Failure 503 {object} ConnectivityResponse
// @Router /connectivity [get]
func Connectivity(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		svcResp := d.DiagnosticsService.Connectivity(ctx)
		resp := ConnectivityResponse{
			Postgres: ConnectivityItem{OK: svcResp.Postgres.OK, Details: svcResp.Postgres.Details, Error: svcResp.Postgres.Error},
			Redis:    ConnectivityItem{OK: svcResp.Redis.OK, Details: svcResp.Redis.Details, Error: svcResp.Redis.Error},
		}

		status := http.StatusOK
		if !resp.Postgres.OK || !resp.Redis.OK {
			status = http.StatusServiceUnavailable
		}

		writeJSON(w, status, resp)
	}
}
