// Package response provides standardized HTTP response utilities.
package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// ErrorEnvelope wraps error responses.
type ErrorEnvelope struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody contains error details.
type ErrorBody struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

// WriteError writes a standardized error response.
func WriteError(w http.ResponseWriter, status int, code, message string, details map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorEnvelope{
		Error: ErrorBody{Code: code, Message: message, Details: details},
	})
}

// WriteJSON writes a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// WriteInternalError logs the error and writes a generic 500 response.
func WriteInternalError(w http.ResponseWriter, err error) {
	if err != nil {
		slog.Error("internal error", "err", err)
	}
	WriteError(w, http.StatusInternalServerError, "internal_error", "internal error", nil)
}
