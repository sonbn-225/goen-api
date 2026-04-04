package report

import (
	"net/http"

	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/core/httpx"
	"github.com/sonbn-225/goen-api-v2/internal/core/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// getDashboardReport godoc
// @Summary Get Dashboard Report
// @Description Return dashboard metrics including account balances, 6-month cashflow and top expenses in current month.
// @Tags reports
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Envelope{data=DashboardReport}
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /reports/dashboard [get]
func (h *Handler) getDashboardReport(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	report, err := h.service.GetDashboardReport(r.Context(), userID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteData(w, http.StatusOK, report)
}
