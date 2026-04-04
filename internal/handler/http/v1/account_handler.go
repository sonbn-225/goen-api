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

// List godoc
// @Summary List Accounts
// @Description Retrieve a list of accounts for the current user
// @Tags Accounts
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.SuccessEnvelope{data=[]dto.AccountResponse}
// @Failure 401 {object} response.ErrorEnvelope
// @Router /accounts [get]
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

	response.WriteSuccess(w, http.StatusOK, items)
}

// Create godoc
// @Summary Create Account
// @Description Create a new banking or manual account for the user
// @Tags Accounts
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateAccountRequest true "Account Creation Payload"
// @Success 201 {object} response.SuccessEnvelope{data=dto.AccountResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /accounts [post]
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

	response.WriteSuccess(w, http.StatusCreated, account)
}

// Get godoc
// @Summary Get Account
// @Description Retrieve specific account properties by ID
// @Tags Accounts
// @Produce json
// @Security BearerAuth
// @Param accountId path string true "Account ID"
// @Success 200 {object} response.SuccessEnvelope{data=dto.AccountResponse}
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Router /accounts/{accountId} [get]
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

	response.WriteSuccess(w, http.StatusOK, acc)
}

// Patch godoc
// @Summary Update Account
// @Description Partially update account information
// @Tags Accounts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param accountId path string true "Account ID"
// @Param request body entity.AccountPatch true "Account Patch Payload"
// @Success 200 {object} response.SuccessEnvelope{data=dto.AccountResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /accounts/{accountId} [patch]
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

	response.WriteSuccess(w, http.StatusOK, acc)
}

// Delete godoc
// @Summary Delete Account
// @Description Delete an account (and its associated dependencies according to business rules)
// @Tags Accounts
// @Security BearerAuth
// @Param accountId path string true "Account ID"
// @Success 204 "No Content"
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Router /accounts/{accountId} [delete]
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

// ListBalances godoc
// @Summary List Account Balances
// @Description Get current financial balance calculations for each account
// @Tags Accounts
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.SuccessEnvelope{data=[]dto.AccountBalanceResponse}
// @Failure 401 {object} response.ErrorEnvelope
// @Router /accounts/balances [get]
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

	response.WriteSuccess(w, http.StatusOK, items)
}

// ListShares godoc
// @Summary List Account Shares
// @Description List active sharing links provided to other users for this account
// @Tags Accounts
// @Produce json
// @Security BearerAuth
// @Param accountId path string true "Account ID"
// @Success 200 {object} response.SuccessEnvelope{data=[]dto.AccountShareResponse}
// @Failure 401 {object} response.ErrorEnvelope
// @Router /accounts/{accountId}/shares [get]
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

	response.WriteSuccess(w, http.StatusOK, items)
}

// UpsertShare godoc
// @Summary Upsert Account Share
// @Description Add or update view/admin permissions for another user on an account
// @Tags Accounts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param accountId path string true "Account ID"
// @Param request body dto.UpsertShareRequest true "Upsert Share Payload"
// @Success 200 {object} response.SuccessEnvelope{data=object}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /accounts/{accountId}/shares [put]
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

	response.WriteSuccess(w, http.StatusOK, item)
}

// RevokeShare godoc
// @Summary Revoke Account Share
// @Description Remove access permissions given to another user for this account
// @Tags Accounts
// @Security BearerAuth
// @Param accountId path string true "Account ID"
// @Param userId path string true "Target User ID or Username"
// @Success 204 "No Content"
// @Failure 401 {object} response.ErrorEnvelope
// @Router /accounts/{accountId}/shares/{userId} [delete]
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
