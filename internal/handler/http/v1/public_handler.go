package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/service"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/response"
)

type PublicHandler struct {
	svc *service.PublicService
}

func NewPublicHandler(svc *service.PublicService) *PublicHandler {
	return &PublicHandler{svc: svc}
}

func (h *PublicHandler) RegisterRoutes(r chi.Router, cfg *config.Config) {
	r.Route("/public/u/{username}", func(r chi.Router) {
		r.Get("/profile", h.GetPublicProfile)
		r.Get("/payment-info", h.GetPaymentInfo)
		r.Get("/participants", h.GetParticipants)
		r.Get("/debts", h.GetDebts)
	})
}

// GetPublicProfile godoc
// @Summary Get Public Profile
// @Description Retrieve a user's publicly visible profile data using their unique username
// @Tags Public
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} object
// @Failure 404 {object} response.ErrorEnvelope
// @Router /public/u/{username}/profile [get]
func (h *PublicHandler) GetPublicProfile(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	profile, err := h.svc.GetPublicProfile(r.Context(), username)
	if err != nil {
		response.WriteError(w, http.StatusNotFound, "user_not_found", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, profile)
}

// GetParticipants godoc
// @Summary Get Public Participants
// @Description Retrieve publicly visible group expense participants linked to the username
// @Tags Public
// @Produce json
// @Param username path string true "Username"
// @Success 200 {array} object
// @Failure 500 {object} response.ErrorEnvelope
// @Router /public/u/{username}/participants [get]
func (h *PublicHandler) GetParticipants(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	participants, err := h.svc.GetParticipants(r.Context(), username)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, participants)
}

// GetDebts godoc
// @Summary Get Public Debts
// @Description Retrieve publicly visible debts for a specific participant linked to the username
// @Tags Public
// @Produce json
// @Param username path string true "Username"
// @Param name query string true "Participant Name"
// @Success 200 {array} object
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /public/u/{username}/debts [get]
func (h *PublicHandler) GetDebts(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	name := r.URL.Query().Get("name")
	if name == "" {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "participant name is required", nil)
		return
	}
	debts, err := h.svc.GetDebts(r.Context(), username, name)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, debts)
}

// GetPaymentInfo godoc
// @Summary Get Payment Info
// @Description Retrieve public payment integration references like QR codes and banking details
// @Tags Public
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} object
// @Failure 404 {object} response.ErrorEnvelope
// @Router /public/u/{username}/payment-info [get]
func (h *PublicHandler) GetPaymentInfo(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	info, err := h.svc.GetPaymentInfo(r.Context(), username)
	if err != nil {
		response.WriteError(w, http.StatusNotFound, "not_found", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, info)
}
