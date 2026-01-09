package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/apierror"
	"github.com/sonbn-225/goen-api/internal/auth"
	"github.com/sonbn-225/goen-api/internal/domain"
	"github.com/sonbn-225/goen-api/internal/services"
)

// ListAccounts godoc
// @Summary List accounts
// @Description List accounts that the current user can access.
// @Tags accounts
// @Produce json
// @Success 200 {array} domain.Account
// @Failure 401 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /accounts [get]
func ListAccounts(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}

		items, err := d.AccountService.ListAccounts(r.Context(), uid)
		if err != nil {
			apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}

		writeJSON(w, http.StatusOK, items)
	}
}

// CreateAccount godoc
// @Summary Create account
// @Description Create a new financial account for the current user.
// @Tags accounts
// @Accept json
// @Produce json
// @Param X-Client-Id header string false "Client instance ID (recommended)"
// @Param body body services.CreateAccountRequest true "Create account request"
// @Success 200 {object} domain.Account
// @Failure 400 {object} apierror.Envelope
// @Failure 401 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /accounts [post]
func CreateAccount(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}

		var req services.CreateAccountRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "invalid json", nil)
			return
		}

		account, err := d.AccountService.CreateAccount(r.Context(), uid, req)
		if err != nil {
			// Map some validation errors to field hints for UI.
			msg := err.Error()
			details := map[string]any{}
			switch {
			case msg == "name is required":
				details["field"] = "name"
			case msg == "color is invalid":
				details["field"] = "color"
			case msg == "account_type is invalid":
				details["field"] = "account_type"
			case msg == "currency must be ISO4217":
				details["field"] = "currency"
			case msg == "parent_account_id is required" || msg == "parent account must be bank" || msg == "parent account must be bank or wallet" || msg == "parent_account_id must be empty":
				details["field"] = "parent_account_id"
			}

			if errors.Is(err, domain.ErrAccountNotFound) {
				apierror.Write(w, http.StatusNotFound, "not_found", "account not found", nil)
				return
			}

			// Treat most request errors as validation_error for MVP.
			if len(details) > 0 {
				apierror.Write(w, http.StatusBadRequest, "validation_error", msg, details)
				return
			}
			apierror.Write(w, http.StatusBadRequest, "validation_error", msg, nil)
			return
		}

		writeJSON(w, http.StatusOK, account)
	}
}

// GetAccount godoc
// @Summary Get account
// @Description Get a single account (must be accessible to current user).
// @Tags accounts
// @Produce json
// @Param accountId path string true "Account ID"
// @Success 200 {object} domain.Account
// @Failure 401 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /accounts/{accountId} [get]
func GetAccount(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}

		accountID := chi.URLParam(r, "accountId")
		if accountID == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "accountId is required", map[string]any{"field": "accountId"})
			return
		}

		acc, err := d.AccountService.GetAccount(r.Context(), uid, accountID)
		if err != nil {
			if errors.Is(err, domain.ErrAccountNotFound) {
				apierror.Write(w, http.StatusNotFound, "not_found", "account not found", nil)
				return
			}
			apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}

		writeJSON(w, http.StatusOK, acc)
	}
}

// PatchAccount godoc
// @Summary Patch account
// @Description Update an account (owner-only). Supports name and status (active|closed).
// @Tags accounts
// @Accept json
// @Produce json
// @Param accountId path string true "Account ID"
// @Param body body domain.AccountPatch true "Account patch"
// @Success 200 {object} domain.Account
// @Failure 400 {object} apierror.Envelope
// @Failure 401 {object} apierror.Envelope
// @Failure 403 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /accounts/{accountId} [patch]
func PatchAccount(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}

		accountID := chi.URLParam(r, "accountId")
		if accountID == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "accountId is required", map[string]any{"field": "accountId"})
			return
		}

		var patch domain.AccountPatch
		if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "invalid json", nil)
			return
		}

		acc, err := d.AccountService.PatchAccount(r.Context(), uid, accountID, patch)
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrAccountInvalidInput):
				apierror.Write(w, http.StatusBadRequest, "validation_error", "invalid account input", nil)
				return
			case errors.Is(err, domain.ErrAccountForbidden):
				apierror.Write(w, http.StatusForbidden, "forbidden", "forbidden", nil)
				return
			case errors.Is(err, domain.ErrAccountNotFound):
				apierror.Write(w, http.StatusNotFound, "not_found", "account not found", nil)
				return
			default:
				apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
				return
			}
		}

		writeJSON(w, http.StatusOK, acc)
	}
}

// DeleteAccount godoc
// @Summary Delete account
// @Description Soft-delete an account (owner-only).
// @Tags accounts
// @Param accountId path string true "Account ID"
// @Success 204
// @Failure 401 {object} apierror.Envelope
// @Failure 403 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /accounts/{accountId} [delete]
func DeleteAccount(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}

		accountID := chi.URLParam(r, "accountId")
		if accountID == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "accountId is required", map[string]any{"field": "accountId"})
			return
		}

		err := d.AccountService.DeleteAccount(r.Context(), uid, accountID)
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrAccountForbidden):
				apierror.Write(w, http.StatusForbidden, "forbidden", "forbidden", nil)
				return
			case errors.Is(err, domain.ErrAccountNotFound):
				apierror.Write(w, http.StatusNotFound, "not_found", "account not found", nil)
				return
			case errors.Is(err, domain.ErrAccountInvalidInput):
				apierror.Write(w, http.StatusBadRequest, "validation_error", "invalid account input", nil)
				return
			default:
				apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
				return
			}
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// ListAccountBalances godoc
// @Summary List account balances
// @Description List computed balances per account for the current user.
// @Tags accounts
// @Produce json
// @Success 200 {array} domain.AccountBalance
// @Failure 401 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /accounts/balances [get]
func ListAccountBalances(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}

		items, err := d.AccountService.ListAccountBalances(r.Context(), uid)
		if err != nil {
			apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}

		writeJSON(w, http.StatusOK, items)
	}
}
