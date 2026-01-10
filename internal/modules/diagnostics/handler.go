package diagnostics

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sonbn-225/goen-api/internal/config"
	"github.com/sonbn-225/goen-api/internal/response"
)

// PingResponse contains ping result.
type PingResponse struct {
	Service   string    `json:"service"`
	Env       string    `json:"env"`
	Time      time.Time `json:"time"`
	RequestID string    `json:"request_id"`
}

// HealthResponse contains health check result.
type HealthResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks,omitempty"`
}

// Handler handles HTTP requests for diagnostics.
type Handler struct {
	svc *Service
	cfg *config.Config
}

// NewHandler creates a new diagnostics handler.
func NewHandler(svc *Service, cfg *config.Config) *Handler {
	return &Handler{svc: svc, cfg: cfg}
}

// RegisterRoutes registers diagnostics routes (no auth required).
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/ping", h.Ping)
	r.Get("/healthz", h.Healthz)
	r.Get("/readyz", h.Readyz)
	r.Get("/connectivity", h.Connectivity)
}

// Ping handles GET /ping
// @Summary Ping
// @Description Lightweight endpoint for browser/app connectivity test.
// @Tags meta
// @Produce json
// @Success 200 {object} PingResponse
// @Router /ping [get]
func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, PingResponse{
		Service:   "goen-api",
		Env:       h.cfg.Env,
		Time:      time.Now().UTC(),
		RequestID: middleware.GetReqID(r.Context()),
	})
}

// Healthz handles GET /healthz
// @Summary Health check
// @Description Liveness check for goen-api.
// @Tags meta
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /healthz [get]
func (h *Handler) Healthz(w http.ResponseWriter, _ *http.Request) {
	response.WriteJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
}

// Readyz handles GET /readyz
// @Summary Readiness check
// @Description Readiness check; includes Postgres/Redis status when configured.
// @Tags meta
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 503 {object} HealthResponse
// @Router /readyz [get]
func (h *Handler) Readyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	checks, ready := h.svc.Readiness(ctx)
	statusCode := http.StatusOK
	if !ready {
		statusCode = http.StatusServiceUnavailable
	}

	resp := HealthResponse{Status: "ok", Checks: checks}
	if !ready {
		resp.Status = "unready"
	}

	response.WriteJSON(w, statusCode, resp)
}

// Connectivity handles GET /connectivity
// @Summary Connectivity probe
// @Description Probes Postgres and Redis and returns diagnostic info.
// @Tags meta
// @Produce json
// @Success 200 {object} ConnectivityResponse
// @Failure 503 {object} ConnectivityResponse
// @Router /connectivity [get]
func (h *Handler) Connectivity(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	svcResp := h.svc.Connectivity(ctx)

	status := http.StatusOK
	if !svcResp.Postgres.OK || !svcResp.Redis.OK {
		status = http.StatusServiceUnavailable
	}

	response.WriteJSON(w, status, svcResp)
}
