// Package apperr defines domain error types and their mapping to HTTP status
// codes, keeping handlers free of business-specific branching.
package apperr

import (
	"errors"
	"net/http"
)

// Kind classifies a domain error.
type Kind int

const (
	// KindInternal is the zero value and maps to 500.
	KindInternal Kind = iota
	// KindNotFound maps to 404.
	KindNotFound
	// KindValidation maps to 400.
	KindValidation
)

// Error is a domain error carrying a Kind and a client-facing message.
type Error struct {
	Kind    Kind
	Message string
	// Err is an optional wrapped cause, preserved for logging.
	Err error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *Error) Unwrap() error { return e.Err }

// NotFound builds a not-found domain error.
func NotFound(msg string, cause error) *Error {
	return &Error{Kind: KindNotFound, Message: msg, Err: cause}
}

// Validation builds a validation domain error.
func Validation(msg string) *Error {
	return &Error{Kind: KindValidation, Message: msg}
}

// HTTPStatus maps any error to an HTTP status code and a client-facing message.
// Non-domain errors are treated as internal and given a generic message so
// implementation details never leak to clients.
func HTTPStatus(err error) (int, string) {
	var appErr *Error
	if errors.As(err, &appErr) {
		switch appErr.Kind {
		case KindNotFound:
			return http.StatusNotFound, appErr.Message
		case KindValidation:
			return http.StatusBadRequest, appErr.Message
		}
	}
	return http.StatusInternalServerError, "internal server error"
}
