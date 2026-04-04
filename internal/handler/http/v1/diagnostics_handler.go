package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/service"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/response"
)

type DiagnosticsHandler struct {
	svc *service.DiagnosticsService
}

func NewDiagnosticsHandler(svc *service.DiagnosticsService) *DiagnosticsHandler {
	return &DiagnosticsHandler{svc: svc}
}

func (h *DiagnosticsHandler) RegisterRoutes(r chi.Router, cfg *config.Config) {
	r.Get("/diagnostics", h.GetDiagnostics)
}

// GetDiagnostics godoc
// @Summary System Diagnostics
// @Description Retrieve metrics, configurations, and connectivity health (internal use docs)
// @Tags System
// @Produce json
// @Success 200 {object} object
// @Failure 500 {object} response.ErrorEnvelope
// @Router /diagnostics [get]
func (h *DiagnosticsHandler) GetDiagnostics(w http.ResponseWriter, r *http.Request) {
	diag, err := h.svc.GetDiagnostics(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, diag)
}
