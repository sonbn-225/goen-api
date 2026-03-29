package public

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/platform/httpx"
	"github.com/sonbn-225/goen-api/internal/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	if userID == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "userId is required", nil)
		return
	}

	prof, err := h.service.GetPublicProfile(r.Context(), userID)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, prof)
}

func (h *Handler) GetPaymentInfo(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	if userID == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "userId is required", nil)
		return
	}

	info, err := h.service.GetPaymentInfo(r.Context(), userID)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, info)
}

func (h *Handler) ListDebts(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	name := r.URL.Query().Get("name")

	if userID == "" || name == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "userId and name are required", nil)
		return
	}

	debts, err := h.service.ListDebtsByName(r.Context(), userID, name)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, debts)
}

func (h *Handler) ListParticipants(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	if userID == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "userId is required", nil)
		return
	}

	names, err := h.service.ListParticipants(r.Context(), userID)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, names)
}

