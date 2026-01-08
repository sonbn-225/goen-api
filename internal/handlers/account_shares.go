package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/apierror"
	"github.com/sonbn-225/goen-api/internal/auth"
	"github.com/sonbn-225/goen-api/internal/domain"
)

type upsertAccountShareRequest struct {
	Login      string `json:"login"`
	Permission string `json:"permission"`
}

func ListAccountShares(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}

		accountID := chi.URLParam(r, "accountId")
		items, err := d.AccountService.ListAccountShares(r.Context(), uid, accountID)
		if err != nil {
			if errors.Is(err, domain.ErrAccountNotFound) {
				apierror.Write(w, http.StatusNotFound, "not_found", "account not found", nil)
				return
			}
			if errors.Is(err, domain.ErrAccountForbidden) || errors.Is(err, domain.ErrAccountShareForbidden) {
				apierror.Write(w, http.StatusForbidden, "forbidden", "forbidden", nil)
				return
			}
			apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}

		writeJSON(w, http.StatusOK, items)
	}
}

func UpsertAccountShare(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}

		accountID := chi.URLParam(r, "accountId")

		var req upsertAccountShareRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierror.Write(w, http.StatusBadRequest, "invalid_request", "invalid json", nil)
			return
		}

		item, err := d.AccountService.UpsertAccountShare(r.Context(), uid, accountID, req.Login, req.Permission)
		if err != nil {
			if errors.Is(err, domain.ErrAccountShareInvalidInput) {
				apierror.Write(w, http.StatusBadRequest, "invalid_request", "invalid request", nil)
				return
			}
			if errors.Is(err, domain.ErrUserNotFound) {
				apierror.Write(w, http.StatusBadRequest, "invalid_request", "user not found", nil)
				return
			}
			if errors.Is(err, domain.ErrAccountForbidden) || errors.Is(err, domain.ErrAccountShareForbidden) {
				apierror.Write(w, http.StatusForbidden, "forbidden", "forbidden", nil)
				return
			}
			if errors.Is(err, domain.ErrAccountNotFound) {
				apierror.Write(w, http.StatusNotFound, "not_found", "account not found", nil)
				return
			}
			apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}

		writeJSON(w, http.StatusOK, item)
	}
}

func RevokeAccountShare(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}

		accountID := chi.URLParam(r, "accountId")
		targetUserID := chi.URLParam(r, "userId")
		if targetUserID == "" {
			apierror.Write(w, http.StatusBadRequest, "invalid_request", "userId is required", nil)
			return
		}

		if err := d.AccountService.RevokeAccountShare(r.Context(), uid, accountID, targetUserID); err != nil {
			if errors.Is(err, domain.ErrAccountForbidden) || errors.Is(err, domain.ErrAccountShareForbidden) {
				apierror.Write(w, http.StatusForbidden, "forbidden", "forbidden", nil)
				return
			}
			if errors.Is(err, domain.ErrAccountNotFound) {
				apierror.Write(w, http.StatusNotFound, "not_found", "account not found", nil)
				return
			}
			apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func ListAuditEvents(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}

		accountID := chi.URLParam(r, "accountId")
		limit := 50
		if v := r.URL.Query().Get("limit"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				limit = n
			}
		}

		items, err := d.AuditService.ListAuditEvents(r.Context(), uid, accountID, limit)
		if err != nil {
			if errors.Is(err, domain.ErrAuditForbidden) || errors.Is(err, domain.ErrAccountForbidden) {
				apierror.Write(w, http.StatusForbidden, "forbidden", "forbidden", nil)
				return
			}
			apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}

		writeJSON(w, http.StatusOK, items)
	}
}
