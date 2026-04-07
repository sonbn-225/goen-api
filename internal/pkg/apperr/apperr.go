package apperr

import (
	"fmt"
	"net/http"
)

type AppError struct {
	StatusCode int            `json:"-"`
	Code       string         `json:"code"`
	Message    string         `json:"message"`
	Details    map[string]any `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func New(status int, code, message string) *AppError {
	return &AppError{
		StatusCode: status,
		Code:       code,
		Message:    message,
	}
}

func NewWithDetails(status int, code, message string, details map[string]any) *AppError {
	return &AppError{
		StatusCode: status,
		Code:       code,
		Message:    message,
		Details:    details,
	}
}

// Common Error Factories
func BadRequest(code, message string) *AppError {
	if code == "" {
		code = "bad_request"
	}
	return New(http.StatusBadRequest, code, message)
}

func Unauthorized(message string) *AppError {
	return New(http.StatusUnauthorized, "unauthorized", message)
}

func Forbidden(message string) *AppError {
	return New(http.StatusForbidden, "forbidden", message)
}

func NotFound(message string) *AppError {
	return New(http.StatusNotFound, "not_found", message)
}

func Conflict(message string) *AppError {
	return New(http.StatusConflict, "conflict", message)
}

func Internal(message string) *AppError {
	if message == "" {
		message = "An internal server error occurred"
	}
	return New(http.StatusInternalServerError, "internal_error", message)
}

// Specific Error Presets
var (
	ErrInvalidID          = BadRequest("invalid_id", "The provided ID is invalid")
	ErrInvalidRequest     = BadRequest("invalid_request", "The request is invalid or malformed")
	ErrInvalidCredentials = Unauthorized("Invalid login credentials")
	ErrTokenExpired       = Unauthorized("Session expired or invalid")
	ErrNotFound           = NotFound("The requested resource was not found")
	ErrConflict           = Conflict("A conflict occurred with the current state of the resource")
	ErrInternal           = Internal("An internal server error occurred")
)
