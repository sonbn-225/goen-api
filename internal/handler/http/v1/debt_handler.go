package v1

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/handler/middleware"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/response"
)

type DebtHandler struct {
	svc interfaces.DebtService
}

func NewDebtHandler(svc interfaces.DebtService) *DebtHandler {
	return &DebtHandler{svc: svc}
}

func (h *DebtHandler) RegisterRoutes(r chi.Router, cfg *config.Config) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg))
		r.Get("/debts", h.List)
		r.Post("/debts", h.Create)
		r.Get("/debts/{id}", h.Get)
		r.Patch("/debts/{id}", h.Update)
		r.Delete("/debts/{id}", h.Delete)
		r.Post("/debts/{id}/payments", h.AddPayment)
		r.Get("/debts/{id}/payments", h.ListPayments)
	})
}

func (h *DebtHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	items, err := h.svc.List(r.Context(), userID)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, items)
}

func (h *DebtHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	var req dto.CreateDebtRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
		return
	}
	res, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusCreated, res)
}

func (h *DebtHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")
	res, err := h.svc.Get(r.Context(), userID, id)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, res)
}

func (h *DebtHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")
	var req dto.UpdateDebtRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
		return
	}
	res, err := h.svc.Update(r.Context(), userID, id, req)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, res)
}

func (h *DebtHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), userID, id); err != nil {
		response.WriteInternalError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *DebtHandler) AddPayment(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")
	var req dto.DebtPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
		return
	}
	res, err := h.svc.AddPayment(r.Context(), userID, id, req)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, res)
}

func (h *DebtHandler) ListPayments(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")
	items, err := h.svc.ListPayments(r.Context(), userID, id)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, items)
}
