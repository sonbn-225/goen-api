package v1

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/handler/middleware"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/response"
)

type GroupExpenseHandler struct {
	svc interfaces.GroupExpenseService
}

func NewGroupExpenseHandler(svc interfaces.GroupExpenseService) *GroupExpenseHandler {
	return &GroupExpenseHandler{svc: svc}
}

func (h *GroupExpenseHandler) RegisterRoutes(r chi.Router, cfg *config.Config) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg))
		r.Post("/group-expenses", h.Create)
		r.Get("/group-expenses/participants/{transactionId}", h.ListByTransaction)
		r.Post("/group-expenses/settle/{participantId}", h.Settle)
		r.Get("/group-expenses/names", h.ListNames)
	})
}

func (h *GroupExpenseHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	var req dto.CreateGroupExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "failed to decode request", nil)
		return
	}

	res, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusCreated, res)
}

func (h *GroupExpenseHandler) ListByTransaction(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	txID := chi.URLParam(r, "transactionId")
	if txID == "" {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "transaction ID is required", nil)
		return
	}

	items, err := h.svc.ListByTransaction(r.Context(), userID, txID)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, items)
}

func (h *GroupExpenseHandler) Settle(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	pID := chi.URLParam(r, "participantId")
	if pID == "" {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "participant ID is required", nil)
		return
	}

	var req dto.GroupExpenseSettleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "failed to decode request", nil)
		return
	}

	tx, err := h.svc.Settle(r.Context(), userID, pID, req)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, tx)
}

func (h *GroupExpenseHandler) ListNames(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil {
			limit = val
		}
	}

	names, err := h.svc.ListUniqueParticipantNames(r.Context(), userID, limit)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, names)
}
