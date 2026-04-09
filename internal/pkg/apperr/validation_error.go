package apperr

import "net/http"

// NewValidationError creates a standardized validation error response.
// The fields map should contain JSON field names and their error codes/messages.
func NewValidationError(fields map[string]any) *AppError {
	return NewWithDetails(
		http.StatusBadRequest,
		"validation_failed",
		"One or more fields are invalid",
		map[string]any{
			"fields": fields,
		},
	)
}
