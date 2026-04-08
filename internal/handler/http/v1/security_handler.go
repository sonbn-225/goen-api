package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/handler/middleware"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/response"
)

type SecurityHandler struct {
	svc interfaces.SecurityService
}

func NewSecurityHandler(svc interfaces.SecurityService) *SecurityHandler {
	return &SecurityHandler{svc: svc}
}

func (h *SecurityHandler) RegisterRoutes(r chi.Router, cfg *config.Config) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg))

		r.Route("/securities", func(r chi.Router) {
			r.Get("/", h.ListSecurities)
			r.Route("/{securityId}", func(r chi.Router) {
				r.Get("/", h.GetSecurity)
				r.Get("/prices-daily", h.ListSecurityPrices)
				r.Get("/events", h.ListSecurityEvents)
			})
		})
	})
}

// ListSecurities godoc
// @Summary List Securities
// @Description Retrieve list of all available securities in the system
// @Tags Securities
// @Produce json
// @Success 200 {object} response.SuccessEnvelope{data=[]dto.SecurityResponse}
// @Failure 500 {object} response.ErrorEnvelope
// @Router /securities [get]
func (h *SecurityHandler) ListSecurities(w http.ResponseWriter, r *http.Request) {
	securities, err := h.svc.ListSecurities(r.Context())
	if err != nil {
		response.HandleError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, securities)
}

// GetSecurity godoc
// @Summary Get Security
// @Description Retrieve details of a specific security by ID
// @Tags Securities
// @Produce json
// @Param securityId path string true "Security ID"
// @Success 200 {object} response.SuccessEnvelope{data=dto.SecurityResponse}
// @Failure 500 {object} response.ErrorEnvelope
// @Router /securities/{securityId} [get]
func (h *SecurityHandler) GetSecurity(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "securityId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid security id format", nil)
		return
	}
	security, err := h.svc.GetSecurity(r.Context(), id)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	if security == nil {
		response.WriteError(w, http.StatusNotFound, "not_found", "security not found", nil)
		return
	}
	response.WriteSuccess(w, http.StatusOK, security)
}

// ListSecurityPrices godoc
// @Summary List Security Prices
// @Description Retrieve historical/daily prices for a security between dates
// @Tags Securities
// @Produce json
// @Param securityId path string true "Security ID"
// @Param from query string false "From Date (YYYY-MM-DD)"
// @Param to query string false "To Date (YYYY-MM-DD)"
// @Success 200 {object} response.SuccessEnvelope{data=[]dto.SecurityPriceDailyResponse}
// @Failure 500 {object} response.ErrorEnvelope
// @Router /securities/{securityId}/prices-daily [get]
func (h *SecurityHandler) ListSecurityPrices(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "securityId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid security id format", nil)
		return
	}
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	var fromPtr, toPtr *string
	if from != "" {
		fromPtr = &from
	}
	if to != "" {
		toPtr = &to
	}

	prices, err := h.svc.ListSecurityPrices(r.Context(), id, fromPtr, toPtr)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, prices)
}

// ListSecurityEvents godoc
// @Summary List Security Events
// @Description Retrieve events (dividends, splits, etc) mapping to a security
// @Tags Securities
// @Produce json
// @Param securityId path string true "Security ID"
// @Param from query string false "From Date (YYYY-MM-DD)"
// @Param to query string false "To Date (YYYY-MM-DD)"
// @Success 200 {object} response.SuccessEnvelope{data=[]dto.SecurityEventResponse}
// @Failure 500 {object} response.ErrorEnvelope
// @Router /securities/{securityId}/events [get]
func (h *SecurityHandler) ListSecurityEvents(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "securityId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid security id format", nil)
		return
	}
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	var fromPtr, toPtr *string
	if from != "" {
		fromPtr = &from
	}
	if to != "" {
		toPtr = &to
	}

	events, err := h.svc.ListSecurityEvents(r.Context(), id, fromPtr, toPtr)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, events)
}
