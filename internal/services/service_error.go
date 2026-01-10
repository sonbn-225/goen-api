package services

import "net/http"

type ErrorKind string

const (
	ErrorKindValidation            ErrorKind = "validation_error"
	ErrorKindInvalidRequest        ErrorKind = "invalid_request"
	ErrorKindUnauthorized          ErrorKind = "unauthorized"
	ErrorKindForbidden             ErrorKind = "forbidden"
	ErrorKindNotFound              ErrorKind = "not_found"
	ErrorKindConflict              ErrorKind = "conflict"
	ErrorKindDependencyUnavailable ErrorKind = "dependency_unavailable"
	ErrorKindInternal              ErrorKind = "internal_error"
)

type ServiceError struct {
	Kind    ErrorKind
	Message string
	Details map[string]any
	Cause   error
}

func (e *ServiceError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func (e *ServiceError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func (e *ServiceError) HTTPStatus() int {
	if e == nil {
		return http.StatusInternalServerError
	}
	switch e.Kind {
	case ErrorKindValidation:
		return http.StatusBadRequest
	case ErrorKindInvalidRequest:
		return http.StatusBadRequest
	case ErrorKindUnauthorized:
		return http.StatusUnauthorized
	case ErrorKindForbidden:
		return http.StatusForbidden
	case ErrorKindNotFound:
		return http.StatusNotFound
	case ErrorKindConflict:
		return http.StatusConflict
	case ErrorKindDependencyUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

func ValidationError(message string, details map[string]any) error {
	return &ServiceError{Kind: ErrorKindValidation, Message: message, Details: details}
}

func ValidationErrorWithCause(message string, details map[string]any, cause error) error {
	return &ServiceError{Kind: ErrorKindValidation, Message: message, Details: details, Cause: cause}
}

func InvalidRequestError(message string, details map[string]any) error {
	return &ServiceError{Kind: ErrorKindInvalidRequest, Message: message, Details: details}
}

func InvalidRequestErrorWithCause(message string, details map[string]any, cause error) error {
	return &ServiceError{Kind: ErrorKindInvalidRequest, Message: message, Details: details, Cause: cause}
}

func DependencyUnavailableError(message string) error {
	return &ServiceError{Kind: ErrorKindDependencyUnavailable, Message: message}
}

func DependencyUnavailableErrorWithCause(message string, cause error) error {
	return &ServiceError{Kind: ErrorKindDependencyUnavailable, Message: message, Cause: cause}
}

func NotFoundError(message string, details map[string]any) error {
	return &ServiceError{Kind: ErrorKindNotFound, Message: message, Details: details}
}

func NotFoundErrorWithCause(message string, details map[string]any, cause error) error {
	return &ServiceError{Kind: ErrorKindNotFound, Message: message, Details: details, Cause: cause}
}

func ForbiddenError(message string) error {
	return &ServiceError{Kind: ErrorKindForbidden, Message: message}
}

func ForbiddenErrorWithCause(message string, details map[string]any, cause error) error {
	return &ServiceError{Kind: ErrorKindForbidden, Message: message, Details: details, Cause: cause}
}

func UnauthorizedError(message string) error {
	return &ServiceError{Kind: ErrorKindUnauthorized, Message: message}
}

func UnauthorizedErrorWithCause(message string, details map[string]any, cause error) error {
	return &ServiceError{Kind: ErrorKindUnauthorized, Message: message, Details: details, Cause: cause}
}

func ConflictError(message string) error {
	return &ServiceError{Kind: ErrorKindConflict, Message: message}
}

func ConflictErrorWithCause(message string, details map[string]any, cause error) error {
	return &ServiceError{Kind: ErrorKindConflict, Message: message, Details: details, Cause: cause}
}

func InternalError(message string) error {
	return &ServiceError{Kind: ErrorKindInternal, Message: message}
}

func InternalErrorWithCause(message string, details map[string]any, cause error) error {
	return &ServiceError{Kind: ErrorKindInternal, Message: message, Details: details, Cause: cause}
}
