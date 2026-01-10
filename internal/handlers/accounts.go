package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/apierror"
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
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		items, err := d.AccountService.ListAccounts(r.Context(), uid)
		if err != nil {
			writeInternalError(w, err)
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
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		var req services.CreateAccountRequest
		if !decodeJSON(w, r, &req) {
			return
		}

		account, err := d.AccountService.CreateAccount(r.Context(), uid, req)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
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
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		accountID := chi.URLParam(r, "accountId")
		if accountID == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "accountId is required", map[string]any{"field": "accountId"})
			return
		}

		acc, err := d.AccountService.GetAccount(r.Context(), uid, accountID)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
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
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		accountID := chi.URLParam(r, "accountId")
		if accountID == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "accountId is required", map[string]any{"field": "accountId"})
			return
		}

		var patch domain.AccountPatch
		if !decodeJSON(w, r, &patch) {
			return
		}

		acc, err := d.AccountService.PatchAccount(r.Context(), uid, accountID, patch)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
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
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		accountID := chi.URLParam(r, "accountId")
		if accountID == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "accountId is required", map[string]any{"field": "accountId"})
			return
		}

		err := d.AccountService.DeleteAccount(r.Context(), uid, accountID)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
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
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		items, err := d.AccountService.ListAccountBalances(r.Context(), uid)
		if err != nil {
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, items)
	}
}
