package v1

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/handler/middleware"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/response"
)

type AccountHandler struct {
	svc interfaces.AccountService
}

func NewAccountHandler(svc interfaces.AccountService) *AccountHandler {
	return &AccountHandler{svc: svc}
}

func (h *AccountHandler) RegisterRoutes(r chi.Router, cfg *config.Config) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg))
		r.Get("/accounts", h.List)
		r.Post("/accounts", h.Create)
		r.Get("/accounts/balances", h.ListBalances)
		r.Get("/accounts/{accountId}", h.Get)
		r.Patch("/accounts/{accountId}", h.Patch)
		r.Delete("/accounts/{accountId}", h.Delete)
		r.Get("/accounts/{accountId}/shares", h.ListShares)
		r.Put("/accounts/{accountId}/shares", h.UpsertShare)
		r.Delete("/accounts/{accountId}/shares/{userId}", h.RevokeShare)
	})
}

func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	items, err := h.svc.List(r.Context(), userID)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, items)
}

func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	var req dto.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}

	account, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusCreated, account)
}

func (h *AccountHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	accountID := chi.URLParam(r, "accountId")
	acc, err := h.svc.Get(r.Context(), userID, accountID)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, acc)
}

func (h *AccountHandler) Patch(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	accountID := chi.URLParam(r, "accountId")
	var patch entity.AccountPatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}

	acc, err := h.svc.Patch(r.Context(), userID, accountID, patch)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, acc)
}

func (h *AccountHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	accountID := chi.URLParam(r, "accountId")
	if err := h.svc.Delete(r.Context(), userID, accountID); err != nil {
		response.WriteInternalError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AccountHandler) ListBalances(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	items, err := h.svc.ListBalances(r.Context(), userID)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, items)
}

func (h *AccountHandler) ListShares(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	accountID := chi.URLParam(r, "accountId")
	items, err := h.svc.ListShares(r.Context(), userID, accountID)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, items)
}

func (h *AccountHandler) UpsertShare(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	accountID := chi.URLParam(r, "accountId")
	var req dto.UpsertShareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}

	item, err := h.svc.UpsertShare(r.Context(), userID, accountID, req.Login, req.Permission)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, item)
}

func (h *AccountHandler) RevokeShare(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	accountID := chi.URLParam(r, "accountId")
	targetUserID := chi.URLParam(r, "userId")
	if err := h.svc.RevokeShare(r.Context(), userID, accountID, targetUserID); err != nil {
		response.WriteInternalError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
