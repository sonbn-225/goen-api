package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/apierror"
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
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		securityID := chi.URLParam(r, "securityId")
		if securityID == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "securityId is required", map[string]any{"field": "securityId"})
			return
		}

		force := r.URL.Query().Get("force")
		resp, err := d.MarketDataService.EnqueueSecurityEvents(r.Context(), uid, securityID, force)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusAccepted, RefreshSecurityEventsResponse{Stream: resp.Stream, MessageID: resp.MessageID})
	}
}
