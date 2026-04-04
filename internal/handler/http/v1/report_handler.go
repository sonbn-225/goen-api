package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/handler/middleware"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/response"
)

type ReportHandler struct {
	reportSvc interfaces.ReportService
}

func NewReportHandler(reportSvc interfaces.ReportService) *ReportHandler {
	return &ReportHandler{reportSvc: reportSvc}
}

func (h *ReportHandler) RegisterRoutes(r chi.Router, cfg *config.Config) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg))
		r.Get("/reports/dashboard", h.GetDashboard)
	})
}

// GetDashboard handles GET /reports/dashboard
func (h *ReportHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	reportData, err := h.reportSvc.GetDashboardReport(r.Context(), userID)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, reportData)
}
