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

// GetDashboard godoc
// @Summary Get Dashboard Report
// @Description Retrieve aggregated analytics, total wealth, and other dashboard indicators for the user
// @Tags Reports
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /reports/dashboard [get]
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
