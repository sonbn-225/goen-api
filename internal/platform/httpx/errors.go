package httpx

import (
	"errors"
	"net/http"

	"github.com/sonbn-225/goen-api/internal/apperrors"
	"github.com/sonbn-225/goen-api/internal/response"
	"github.com/sonbn-225/goen-api/internal/shared/contracts"
)

// WriteServiceError normalizes app/service errors to the standard API error envelope.
func WriteServiceError(w http.ResponseWriter, err error) {
	var se *apperrors.Error
	if errors.As(err, &se) {
		response.WriteError(w, se.HTTPStatus(), string(se.Kind), se.Message, se.Details)
		return
	}
	response.WriteInternalError(w, err)
}

// WriteUnauthorized writes the standard unauthorized error envelope.
func WriteUnauthorized(w http.ResponseWriter) {
	response.WriteError(w, http.StatusUnauthorized, contracts.ErrorCodeUnauthorized, contracts.ErrorMessageUnauthorized, nil)
}

// WriteValidationError writes the standard validation error envelope.
func WriteValidationError(w http.ResponseWriter, message string, details map[string]any) {
	response.WriteError(w, http.StatusBadRequest, contracts.ErrorCodeValidation, message, details)
}
