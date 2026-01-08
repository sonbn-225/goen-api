package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/apierror"
	"github.com/sonbn-225/goen-api/internal/auth"
	"github.com/sonbn-225/goen-api/internal/domain"
)

type RefreshSecurityEventsResponse struct {
	Stream    string `json:"stream"`
	MessageID string `json:"message_id"`
}

// RefreshSecurityEvents godoc
// @Summary Enqueue refresh security events
// @Description Enqueue a vnstock job; worker will fetch corporate actions/events and upsert into security_events.
// @Tags investments
// @Produce json
// @Param securityId path string true "Security ID"
// @Success 202 {object} RefreshSecurityEventsResponse
// @Failure 400 {object} apierror.Envelope
// @Failure 401 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 503 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /securities/{securityId}/events/refresh [post]
func RefreshSecurityEvents(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}
		if d.Redis == nil {
			apierror.Write(w, http.StatusServiceUnavailable, "dependency_unavailable", "redis is not configured", nil)
			return
		}

		securityID := chi.URLParam(r, "securityId")
		if securityID == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "securityId is required", map[string]any{"field": "securityId"})
			return
		}

		if _, err := d.InvestmentService.GetSecurity(r.Context(), uid, securityID); err != nil {
			if err == domain.ErrSecurityNotFound {
				apierror.Write(w, http.StatusNotFound, "not_found", "security not found", nil)
				return
			}
			apierror.Write(w, http.StatusBadRequest, "validation_error", err.Error(), nil)
			return
		}

		force := r.URL.Query().Get("force")

		stream := "goen:market_data:jobs"
		values := map[string]any{
			"job_type":             "vnstock.security_events",
			"security_id":          securityID,
			"requested_by_user_id": uid,
		}
		if force != "" {
			values["force"] = force
		}

		id, err := d.Redis.XAdd(r.Context(), stream, values)
		if err != nil {
			apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}

		writeJSON(w, http.StatusAccepted, RefreshSecurityEventsResponse{Stream: stream, MessageID: id})
	}
}
