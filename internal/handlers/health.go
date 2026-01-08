package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/sonbn-225/goen-api/internal/config"
	"github.com/sonbn-225/goen-api/internal/services"
	"github.com/sonbn-225/goen-api/internal/storage"
)

type Deps struct {
	Cfg         *config.Config
	DB          *storage.Postgres
	Redis       *storage.Redis
	AuthService services.AuthService
	AccountService services.AccountService
	AuditService services.AuditService
	TransactionService services.TransactionService
	CategoryService services.CategoryService
	TagService services.TagService
	BudgetService services.BudgetService
	SavingsService services.SavingsService
	RotatingSavingsService services.RotatingSavingsService
	DebtService services.DebtService
	InvestmentService services.InvestmentService
}

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

		checks := map[string]string{}
		statusCode := http.StatusOK

		if d.DB != nil {
			if err := d.DB.Ping(ctx); err != nil {
				checks["postgres"] = "error"
				statusCode = http.StatusServiceUnavailable
			} else {
				checks["postgres"] = "ok"
			}
		}
		if d.Redis != nil {
			if err := d.Redis.Ping(ctx); err != nil {
				checks["redis"] = "error"
				statusCode = http.StatusServiceUnavailable
			} else {
				checks["redis"] = "ok"
			}
		}

		resp := HealthResponse{Status: "ok", Checks: checks}
		if statusCode != http.StatusOK {
			resp.Status = "unready"
		}

		writeJSON(w, statusCode, resp)
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
