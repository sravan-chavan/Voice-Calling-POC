package voice

import (
	"errors"
	"fmt"
)

// Sentinel error classes for callers to inspect with errors.Is.
var (
	ErrInvalidInput   = errors.New("invalid input")
	ErrConfiguration  = errors.New("configuration error")
	ErrAuthentication = errors.New("authentication error")
	ErrProviderAPI    = errors.New("provider api error")
	ErrNetwork        = errors.New("network error")
	ErrUnexpected     = errors.New("unexpected error")
)

// Error is a structured, wrap-friendly error returned by the voice module.
type Error struct {
	Kind    error
	Message string
	Cause   error
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Kind.Error(), e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Kind.Error(), e.Message)
}

func (e *Error) Unwrap() []error {
	if e.Cause == nil {
		return []error{e.Kind}
	}
	return []error{e.Kind, e.Cause}
}

func NewError(kind error, message string, cause error) *Error {
	return &Error{Kind: kind, Message: message, Cause: cause}
}
