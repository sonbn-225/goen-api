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
	r.Get("/public/profile/{userRef}", h.GetPublicProfile)
	r.Get("/public/payment-info/{userRef}", h.GetPaymentInfo)
}

func (h *PublicHandler) GetPublicProfile(w http.ResponseWriter, r *http.Request) {
	userRef := chi.URLParam(r, "userRef")
	profile, err := h.svc.GetPublicProfile(r.Context(), userRef)
	if err != nil {
		response.WriteError(w, http.StatusNotFound, "user_not_found", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, profile)
}

func (h *PublicHandler) GetPaymentInfo(w http.ResponseWriter, r *http.Request) {
	userRef := chi.URLParam(r, "userRef")
	info, err := h.svc.GetPaymentInfo(r.Context(), userRef)
	if err != nil {
		response.WriteError(w, http.StatusNotFound, "not_found", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, info)
}
