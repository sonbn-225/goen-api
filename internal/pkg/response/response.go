package response

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/sonbn-225/goen-api/internal/pkg/apperr"
)

type SuccessEnvelope struct {
	Data any `json:"data"`
	Meta any `json:"meta,omitempty"`
}

type ErrorEnvelope struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if body != nil {
		_ = json.NewEncoder(w).Encode(body)
	}
}

func WriteSuccess(w http.ResponseWriter, status int, data any) {
	WriteJSON(w, status, SuccessEnvelope{
		Data: data,
	})
}

func WriteSuccessWithMeta(w http.ResponseWriter, status int, data any, meta any) {
	WriteJSON(w, status, SuccessEnvelope{
		Data: data,
		Meta: meta,
	})
}

func WriteAppError(w http.ResponseWriter, err *apperr.AppError) {
	WriteJSON(w, err.StatusCode, ErrorEnvelope{
		Error: ErrorBody{
			Code:    err.Code,
			Message: err.Message,
			Details: err.Details,
		},
	})
}

func WriteError(w http.ResponseWriter, status int, code, message string, details map[string]any) {
	WriteJSON(w, status, ErrorEnvelope{
		Error: ErrorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

func WriteInternalError(w http.ResponseWriter, err error) {
	if err != nil {
		slog.Error("internal server error", "error", err)
	}
	WriteError(w, http.StatusInternalServerError, "internal_error", "An internal server error occurred", nil)
}

func HandleError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	if appErr, ok := err.(*apperr.AppError); ok {
		WriteAppError(w, appErr)
		return
	}

	WriteInternalError(w, err)
}
