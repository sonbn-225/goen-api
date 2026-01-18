package account

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/apperrors"
	"github.com/sonbn-225/goen-api/internal/domain"
	"github.com/sonbn-225/goen-api/internal/httpapi"
	"github.com/sonbn-225/goen-api/internal/response"
)

// Handler handles HTTP requests for accounts.
type Handler struct {
	svc      *Service
	auditSvc AuditServiceInterface
}

// NewHandler creates a new account handler.
func NewHandler(svc *Service, auditSvc AuditServiceInterface) *Handler {
	return &Handler{svc: svc, auditSvc: auditSvc}
}

// RegisterRoutes registers all account routes.
func (h *Handler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.With(authMiddleware).Get("/accounts", h.List)
	r.With(authMiddleware).Get("/accounts/", h.List)
	r.With(authMiddleware).Post("/accounts", h.Create)
	r.With(authMiddleware).Post("/accounts/", h.Create)
	r.With(authMiddleware).Get("/accounts/balances", h.ListBalances)
	r.With(authMiddleware).Get("/accounts/{accountId}", h.Get)
	r.With(authMiddleware).Patch("/accounts/{accountId}", h.Patch)
	r.With(authMiddleware).Delete("/accounts/{accountId}", h.Delete)
	r.With(authMiddleware).Get("/accounts/{accountId}/shares", h.ListShares)
	r.With(authMiddleware).Put("/accounts/{accountId}/shares", h.UpsertShare)
	r.With(authMiddleware).Delete("/accounts/{accountId}/shares/{userId}", h.RevokeShare)
	r.With(authMiddleware).Get("/accounts/{accountId}/audit-events", h.ListAuditEvents)
}

// List handles GET /accounts
// @Summary List accounts
// @Description List accounts that the current user can access.
// @Tags accounts
// @Produce json
// @Success 200 {array} domain.Account
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /accounts [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	items, err := h.svc.List(r.Context(), userID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, items)
}

// Create handles POST /accounts
// @Summary Create account
// @Description Create a new financial account for the current user.
// @Tags accounts
// @Accept json
// @Produce json
// @Param body body CreateAccountRequest true "Create account request"
// @Success 200 {object} domain.Account
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /accounts [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	var req CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}

	account, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, account)
}

// Get handles GET /accounts/{accountId}
// @Summary Get account
// @Description Get a single account (must be accessible to current user).
// @Tags accounts
// @Produce json
// @Param accountId path string true "Account ID"
// @Success 200 {object} domain.Account
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /accounts/{accountId} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	accountID := chi.URLParam(r, "accountId")
	if accountID == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "accountId is required", map[string]any{"field": "accountId"})
		return
	}

	acc, err := h.svc.Get(r.Context(), userID, accountID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, acc)
}

// Patch handles PATCH /accounts/{accountId}
// @Summary Patch account
// @Description Update an account (owner-only).
// @Tags accounts
// @Accept json
// @Produce json
// @Param accountId path string true "Account ID"
// @Param body body domain.AccountPatch true "Account patch"
// @Success 200 {object} domain.Account
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 403 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /accounts/{accountId} [patch]
func (h *Handler) Patch(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	accountID := chi.URLParam(r, "accountId")
	if accountID == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "accountId is required", map[string]any{"field": "accountId"})
		return
	}

	var patch domain.AccountPatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}

	acc, err := h.svc.Patch(r.Context(), userID, accountID, patch)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, acc)
}

// Delete handles DELETE /accounts/{accountId}
// @Summary Delete account
// @Description Soft-delete an account (owner-only).
// @Tags accounts
// @Param accountId path string true "Account ID"
// @Success 204
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 403 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /accounts/{accountId} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	accountID := chi.URLParam(r, "accountId")
	if accountID == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "accountId is required", map[string]any{"field": "accountId"})
		return
	}

	if err := h.svc.Delete(r.Context(), userID, accountID); err != nil {
		h.writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListBalances handles GET /accounts/balances
// @Summary List account balances
// @Description List computed balances per account for the current user.
// @Tags accounts
// @Produce json
// @Success 200 {array} domain.AccountBalance
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /accounts/balances [get]
func (h *Handler) ListBalances(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	items, err := h.svc.ListBalances(r.Context(), userID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, items)
}

type upsertShareRequest struct {
	Login      string `json:"login"`
	Permission string `json:"permission"`
}

// ListShares handles GET /accounts/{accountId}/shares
func (h *Handler) ListShares(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	accountID := chi.URLParam(r, "accountId")
	items, err := h.svc.ListShares(r.Context(), userID, accountID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, items)
}

// UpsertShare handles PUT /accounts/{accountId}/shares
func (h *Handler) UpsertShare(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	accountID := chi.URLParam(r, "accountId")

	var req upsertShareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}

	item, err := h.svc.UpsertShare(r.Context(), userID, accountID, req.Login, req.Permission)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, item)
}

// RevokeShare handles DELETE /accounts/{accountId}/shares/{userId}
func (h *Handler) RevokeShare(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	accountID := chi.URLParam(r, "accountId")
	targetUserID := chi.URLParam(r, "userId")
	if targetUserID == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "userId is required", nil)
		return
	}

	if err := h.svc.RevokeShare(r.Context(), userID, accountID, targetUserID); err != nil {
		h.writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListAuditEvents handles GET /accounts/{accountId}/audit-events
func (h *Handler) ListAuditEvents(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	accountID := chi.URLParam(r, "accountId")
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}

	items, err := h.auditSvc.ListAuditEvents(context.Background(), userID, accountID, limit)
	if err != nil {
		var se *apperrors.Error
		if errors.As(err, &se) {
			response.WriteError(w, se.HTTPStatus(), string(se.Kind), se.Message, se.Details)
			return
		}
		response.WriteInternalError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	var se *apperrors.Error
	if errors.As(err, &se) {
		response.WriteError(w, se.HTTPStatus(), string(se.Kind), se.Message, se.Details)
		return
	}
	response.WriteInternalError(w, err)
}
