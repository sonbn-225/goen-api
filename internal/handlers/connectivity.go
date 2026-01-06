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

		resp := ConnectivityResponse{}

		if d.DB == nil {
			resp.Postgres = ConnectivityItem{OK: false, Error: "DATABASE_URL not set"}
		} else if details, err := d.DB.Probe(ctx); err != nil {
			resp.Postgres = ConnectivityItem{OK: false, Error: err.Error()}
		} else {
			resp.Postgres = ConnectivityItem{OK: true, Details: details}
		}

		if d.Redis == nil {
			resp.Redis = ConnectivityItem{OK: false, Error: "REDIS_URL not set or invalid"}
		} else if details, err := d.Redis.Probe(ctx); err != nil {
			resp.Redis = ConnectivityItem{OK: false, Error: err.Error()}
		} else {
			resp.Redis = ConnectivityItem{OK: true, Details: details}
		}

		status := http.StatusOK
		if !resp.Postgres.OK || !resp.Redis.OK {
			status = http.StatusServiceUnavailable
		}

		writeJSON(w, status, resp)
	}
}
