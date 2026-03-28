// Package apperrors provides unified error handling for the application.
// It combines domain sentinel errors with rich service-layer errors.
package apperrors

import (
	"errors"
	"net/http"
)

// =====================
// Error Kinds (for HTTP mapping)
// =====================

// Kind represents the category of an error.
type Kind string

const (
	KindValidation            Kind = "validation_error"
	KindInvalidRequest        Kind = "invalid_request"
	KindUnauthorized          Kind = "unauthorized"
	KindForbidden             Kind = "forbidden"
	KindNotFound              Kind = "not_found"
	KindConflict              Kind = "conflict"
	KindDependencyUnavailable Kind = "dependency_unavailable"
	KindInternal              Kind = "internal_error"
)

// =====================
// Rich Error Type
// =====================

// Error represents a structured application error.
type Error struct {
	Kind    Kind
	Message string
	Details map[string]any
	Cause   error
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

// Unwrap returns the underlying cause for errors.Is/As support.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// HTTPStatus returns the appropriate HTTP status code.
func (e *Error) HTTPStatus() int {
	if e == nil {
		return http.StatusInternalServerError
	}
	switch e.Kind {
	case KindValidation, KindInvalidRequest:
		return http.StatusBadRequest
	case KindUnauthorized:
		return http.StatusUnauthorized
	case KindForbidden:
		return http.StatusForbidden
	case KindNotFound:
		return http.StatusNotFound
	case KindConflict:
		return http.StatusConflict
	case KindDependencyUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// =====================
// Constructors
// =====================

func Validation(message string, details map[string]any) *Error {
	return &Error{Kind: KindValidation, Message: message, Details: details}
}

func NotFound(message string, details map[string]any) *Error {
	return &Error{Kind: KindNotFound, Message: message, Details: details}
}

func Forbidden(message string) *Error {
	return &Error{Kind: KindForbidden, Message: message}
}

func Unauthorized(message string) *Error {
	return &Error{Kind: KindUnauthorized, Message: message}
}

func Conflict(message string) *Error {
	return &Error{Kind: KindConflict, Message: message}
}

func Internal(message string) *Error {
	return &Error{Kind: KindInternal, Message: message}
}

func DependencyUnavailable(message string) *Error {
	return &Error{Kind: KindDependencyUnavailable, Message: message}
}

func Wrap(kind Kind, message string, cause error) *Error {
	return &Error{Kind: kind, Message: message, Cause: cause}
}

// =====================
// Sentinel Errors (Domain Layer)
// =====================

// User
var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserIDRequired     = errors.New("userID is required")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// Account
var (
	ErrAccountNotFound     = errors.New("account not found")
	ErrAccountForbidden    = errors.New("account forbidden")
	ErrAccountInvalidInput = errors.New("invalid account input")
	ErrAccountClosed       = errors.New("account is closed")
)

// Account Share
var (
	ErrAccountShareForbidden    = errors.New("account share forbidden")
	ErrAccountShareInvalidInput = errors.New("invalid account share input")
)

// Transaction
var (
	ErrTransactionNotFound    = errors.New("transaction not found")
	ErrTransactionForbidden   = errors.New("transaction forbidden")
	ErrTransactionInvalidType = errors.New("type is invalid")
	ErrTransactionPatchFailed = errors.New("patch failed")
	ErrAccountIDRequired      = errors.New("account_id is required")
	ErrFromAccountIDRequired  = errors.New("from_account_id is required")
	ErrToAccountIDRequired    = errors.New("to_account_id is required")
	ErrFXAmountsRequired      = errors.New("from_amount and to_amount are required for FX transfer")
	ErrCategoryIDInvalid      = errors.New("category_id is invalid")
	ErrTagIDsInvalid          = errors.New("tag_ids contains invalid tag")
)

// Category
var ErrCategoryNotFound = errors.New("category not found")

// Tag
var ErrTagNotFound = errors.New("tag not found")

// Budget
var ErrBudgetNotFound = errors.New("budget not found")

// Audit
var ErrAuditForbidden = errors.New("audit forbidden")

// Savings
var ErrSavingsInstrumentNotFound = errors.New("savings instrument not found")

// Rotating Savings
var (
	ErrRotatingSavingsGroupNotFound        = errors.New("rotating savings group not found")
	ErrRotatingSavingsContributionNotFound = errors.New("rotating savings contribution not found")
	ErrRotatingSavingsNameRequired         = errors.New("rotating savings name is required")
)

// Debt
var ErrDebtNotFound = errors.New("debt not found")

// Group Expense
var (
	ErrGroupExpenseParticipantNotFound       = errors.New("group expense participant not found")
	ErrGroupExpenseParticipantAlreadySettled = errors.New("group expense participant already settled")
	ErrNotImplemented                        = errors.New("not implemented")
)

// Investment
var (
	ErrInvestmentAccountNotFound     = errors.New("investment account not found")
	ErrSecurityNotFound              = errors.New("security not found")
	ErrTradeNotFound                 = errors.New("trade not found")
	ErrHoldingNotFound               = errors.New("holding not found")
	ErrSecurityEventNotFound         = errors.New("security event not found")
	ErrSecurityEventElectionNotFound = errors.New("security event election not found")
	ErrInvestmentForbidden           = errors.New("investment forbidden")
)

// Infrastructure
var (
	ErrDatabaseNotReady     = errors.New("database not ready")
	ErrRedisNotReady        = errors.New("redis not ready")
	ErrInvalidDecimalAmount = errors.New("invalid decimal amount")
	ErrInvalidCursor        = errors.New("invalid cursor")
	ErrStreamRequired       = errors.New("stream is required")
)

