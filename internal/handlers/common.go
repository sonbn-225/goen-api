package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/sonbn-225/goen-api/internal/apierror"
	"github.com/sonbn-225/goen-api/internal/auth"
	"github.com/sonbn-225/goen-api/internal/services"
)

const maxJSONBodyBytes int64 = 1 << 20 // 1 MiB

func requireUserID(w http.ResponseWriter, r *http.Request) (string, bool) {
	uid, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return "", false
	}
	return uid, true
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		// Provide actionable validation messages (still safe to expose).
		var syntaxErr *json.SyntaxError
		var typeErr *json.UnmarshalTypeError
		switch {
		case errors.Is(err, io.EOF):
			apierror.Write(w, http.StatusBadRequest, "validation_error", "empty json body", nil)
		case errors.As(err, &syntaxErr):
			apierror.Write(w, http.StatusBadRequest, "validation_error", "malformed json", nil)
		case errors.As(err, &typeErr):
			details := map[string]any{}
			if typeErr.Field != "" {
				details["field"] = typeErr.Field
			}
			apierror.Write(w, http.StatusBadRequest, "validation_error", "invalid json type for field", details)
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			// Example: json: unknown field "currency"
			field := strings.TrimPrefix(err.Error(), "json: unknown field ")
			field = strings.Trim(field, "\"")
			apierror.Write(w, http.StatusBadRequest, "validation_error", "unknown field", map[string]any{"field": field})
		default:
			apierror.Write(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		}
		return false
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		apierror.Write(w, http.StatusBadRequest, "validation_error", "invalid json body", map[string]any{"reason": "multiple json values"})
		return false
	}
	return true
}

func writeInternalError(w http.ResponseWriter, err error) {
	if err != nil {
		slog.Error("internal error", "err", err)
	}
	apierror.Write(w, http.StatusInternalServerError, "internal_error", "internal error", nil)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeServiceError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	var se *services.ServiceError
	if errors.As(err, &se) {
		apierror.Write(w, se.HTTPStatus(), string(se.Kind), se.Message, se.Details)
		return true
	}
	return false
}

func isClientError(err error) bool {
	// Helpers for gradual migration: treat common errors as client-facing.
	return err == nil || errors.Is(err, io.EOF)
}
