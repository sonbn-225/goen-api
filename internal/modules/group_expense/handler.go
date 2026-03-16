package group_expense

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/platform/httpx"
	"github.com/sonbn-225/goen-api/internal/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.With(authMiddleware).Post("/transactions/group-expense", h.Create)
	r.With(authMiddleware).Post("/transactions/group-expense/", h.Create)

	r.With(authMiddleware).Get("/transactions/{transactionId}/group-expense-participants", h.ListByTransaction)
	r.With(authMiddleware).Get("/transactions/{transactionId}/group-expense-participants/", h.ListByTransaction)

	r.With(authMiddleware).Get("/group-expense/participants", h.ListParticipantNames)
	r.With(authMiddleware).Get("/group-expense/participants/", h.ListParticipantNames)

	r.With(authMiddleware).Post("/group-expense-participants/{participantId}/settle", h.Settle)
	r.With(authMiddleware).Post("/group-expense-participants/{participantId}/settle/", h.Settle)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid request body", nil)
		return
	}

	var body CreateRequest
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}

	res, err := h.svc.Create(r.Context(), userID, body)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, res)
}

func (h *Handler) ListByTransaction(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	txID := chi.URLParam(r, "transactionId")
	if txID == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "transactionId is required", map[string]any{"field": "transactionId"})
		return
	}

	items, err := h.svc.ListByTransaction(r.Context(), userID, txID)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) ListParticipantNames(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}

	names, err := h.svc.ListUniqueParticipantNames(r.Context(), userID, limit)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, names)
}

func (h *Handler) Settle(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	pid := chi.URLParam(r, "participantId")
	if pid == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "participantId is required", map[string]any{"field": "participantId"})
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid request body", nil)
		return
	}

	var body SettleRequest
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}

	tx, err := h.svc.Settle(r.Context(), userID, pid, body)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, tx)
}

