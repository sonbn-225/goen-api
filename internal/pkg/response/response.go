package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
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
