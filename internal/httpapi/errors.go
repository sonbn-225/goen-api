package httpapi

import (
	"net/http"

	"github.com/sonbn-225/goen-api/internal/platform/httpx"
	"github.com/sonbn-225/goen-api/internal/shared/contracts"
)

const (
	ErrorCodeUnauthorized    = contracts.ErrorCodeUnauthorized
	ErrorCodeValidation      = contracts.ErrorCodeValidation
	ErrorMessageUnauthorized = contracts.ErrorMessageUnauthorized
)

// WriteServiceError is a compatibility wrapper; prefer platform/httpx.WriteServiceError.
func WriteServiceError(w http.ResponseWriter, err error) {
	httpx.WriteServiceError(w, err)
}

// WriteUnauthorized is a compatibility wrapper; prefer platform/httpx.WriteUnauthorized.
func WriteUnauthorized(w http.ResponseWriter) {
	httpx.WriteUnauthorized(w)
}

// WriteValidationError is a compatibility wrapper; prefer platform/httpx.WriteValidationError.
func WriteValidationError(w http.ResponseWriter, message string, details map[string]any) {
	httpx.WriteValidationError(w, message, details)
}
