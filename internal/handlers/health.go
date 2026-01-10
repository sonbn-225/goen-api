package handlers

import (
	"context"
	"net/http"
	"time"
)

type HealthResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks,omitempty"`
}

// Healthz godoc
// @Summary Health check
// @Description Liveness check for goen-api.
// @Tags meta
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /healthz [get]
func Healthz(_ Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
	}
}

// Readyz godoc
// @Summary Readiness check
// @Description Readiness check; includes Postgres/Redis status when configured.
// @Tags meta
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 503 {object} HealthResponse
// @Router /readyz [get]
func Readyz(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		checks, ready := d.DiagnosticsService.Readiness(ctx)
		statusCode := http.StatusOK
		if !ready {
			statusCode = http.StatusServiceUnavailable
		}

		resp := HealthResponse{Status: "ok", Checks: checks}
		if !ready {
			resp.Status = "unready"
		}

		writeJSON(w, statusCode, resp)
	}
}
