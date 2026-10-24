package apperr

import (
	"errors"
	"net/http"
)

type Kind int

const (
	KindInternal Kind = iota
	KindInvalid
	KindValidation
	KindUnauthorized
	KindForbidden
	KindNotFound
	KindConflict
)

const genericMessage = "something went wrong"

type Error struct {
	Kind    Kind
	Message string
	cause   error
}

func (e *Error) Error() string {
	if e.cause != nil {
		return e.Message + ": " + e.cause.Error()
	}
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.cause
}

func New(kind Kind, message string) *Error {
	return &Error{Kind: kind, Message: message}
}

func Wrap(kind Kind, message string, cause error) *Error {
	return &Error{Kind: kind, Message: message, cause: cause}
}

func statusFor(kind Kind) int {
	switch kind {
	case KindInvalid, KindValidation:
		return http.StatusBadRequest
	case KindUnauthorized:
		return http.StatusUnauthorized
	case KindForbidden:
		return http.StatusForbidden
	case KindNotFound:
		return http.StatusNotFound
	case KindConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func Resolve(err error) (int, string) {
	if err == nil {
		return http.StatusOK, ""
	}

	var appErr *Error
	if errors.As(err, &appErr) {
		status := statusFor(appErr.Kind)
		if status >= http.StatusInternalServerError {
			return status, genericMessage
		}
		return status, appErr.Message
	}

	return http.StatusInternalServerError, genericMessage
}
