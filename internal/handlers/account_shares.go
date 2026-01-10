package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/apierror"
)

type upsertAccountShareRequest struct {
	Login      string `json:"login"`
	Permission string `json:"permission"`
}

func ListAccountShares(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		accountID := chi.URLParam(r, "accountId")
		items, err := d.AccountService.ListAccountShares(r.Context(), uid, accountID)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, items)
	}
}

func UpsertAccountShare(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		accountID := chi.URLParam(r, "accountId")

		var req upsertAccountShareRequest
		if ok := decodeJSON(w, r, &req); !ok {
			return
		}

		item, err := d.AccountService.UpsertAccountShare(r.Context(), uid, accountID, req.Login, req.Permission)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, item)
	}
}

func RevokeAccountShare(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		accountID := chi.URLParam(r, "accountId")
		targetUserID := chi.URLParam(r, "userId")
		if targetUserID == "" {
			apierror.Write(w, http.StatusBadRequest, "invalid_request", "userId is required", nil)
			return
		}

		if err := d.AccountService.RevokeAccountShare(r.Context(), uid, accountID, targetUserID); err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func ListAuditEvents(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
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
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, items)
	}
}
