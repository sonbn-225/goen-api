package response

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/core/logx"
)

type Meta struct {
	Total      int `json:"total,omitempty"`
	Page       int `json:"page,omitempty"`
	Limit      int `json:"limit,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

type Envelope struct {
	Data any   `json:"data"`
	Meta *Meta `json:"meta,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorEnvelope struct {
	Error APIError `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	WriteData(w, status, payload)
}

func WriteData(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{Data: payload})
}

func WriteList(w http.ResponseWriter, status int, payload any, meta Meta) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{Data: payload, Meta: &meta})
}

func WriteError(w http.ResponseWriter, err error) {
	kind := apperrors.KindOf(err)
	status := statusFromKind(kind)
	code := codeFromKind(kind)
	maskedAttrs := logx.MaskAttrs(
		"status", status,
		"code", code,
		"kind", kind,
		"message", err.Error(),
	)
	logger := logx.LoggerFromContext(context.TODO())

	if status >= http.StatusInternalServerError {
		logger.Error("api_error", maskedAttrs...)
	} else {
		logger.Warn("api_error", maskedAttrs...)
	}

	writeRaw(w, status, ErrorEnvelope{
		Error: APIError{
			Code:    code,
			Message: err.Error(),
		},
	})
}

func writeRaw(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func statusFromKind(kind apperrors.Kind) int {
	switch kind {
	case apperrors.KindValidation:
		return http.StatusBadRequest
	case apperrors.KindUnauth:
		return http.StatusUnauthorized
	case apperrors.KindForbidden:
		return http.StatusForbidden
	case apperrors.KindNotFound:
		return http.StatusNotFound
	case apperrors.KindConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func codeFromKind(kind apperrors.Kind) string {
	switch kind {
	case apperrors.KindValidation:
		return "validation_error"
	case apperrors.KindUnauth:
		return "unauthorized"
	case apperrors.KindForbidden:
		return "forbidden"
	case apperrors.KindNotFound:
		return "not_found"
	case apperrors.KindConflict:
		return "conflict"
	default:
		return "internal_error"
	}
}
