package errors

import (
	"errors"
	"fmt"
)

// PErrors errors represents errors of this package
type PErrors struct {
	Code    int
	Message string
	Details string
	Err     error
}

const (
	ErrCodeValueNotSupported        = 1
	ErrCodeUnexpectedLength         = 2
	ErrCodeNotAList                 = 3
	ErrCodeNotAString               = 4
	ErrCodeCannotValueBeASingleByte = 5
	ErrCodeTxTypeNotSupported       = 6
	ErrCodeInvalidSig               = 7
	ErrCodeInvalidPkb               = 8
)

var (
	ErrValueNotSupport          = NewPError(ErrCodeValueNotSupported, "value not supported")
	ErrUnexpectedLength         = NewPError(ErrCodeUnexpectedLength, "unexpected length")
	ErrNotAList                 = NewPError(ErrCodeNotAList, "not a list")
	ErrNotAString               = NewPError(ErrCodeNotAString, "not a string")
	ErrCannotValueBeASingleByte = NewPError(ErrCodeCannotValueBeASingleByte, "cannot be a single byte")
	ErrTxTypeNotSupported       = NewPError(ErrCodeTxTypeNotSupported, "tx type not supported")
	ErrInvalidSig               = NewPError(ErrCodeInvalidSig, "invalid signature")
	ErrInvalidPkb               = NewPError(ErrCodeInvalidPkb, "invalid public key")
)

// NewPError creates a new PErrors
func NewPError(code int, message string) *PErrors {
	return &PErrors{
		Code:    code,
		Message: message,
	}
}

// Error implements the error interface
func (e *PErrors) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *PErrors) Unwrap() error {
	return e.Err
}

// WithMessage wraps an existing error with a new message
func (e *PErrors) WithMessage(message string) *PErrors {
	return &PErrors{
		Code:    e.Code,
		Details: message,
		Message: e.Message,
		Err:     e,
	}
}

// WithMessagef wraps an existing error with a formatted message
func (e *PErrors) WithMessagef(format string, args ...interface{}) *PErrors {
	return &PErrors{
		Code:    e.Code,
		Details: fmt.Sprintf(format, args...),
		Message: e.Message,
		Err:     e,
	}
}

// Is checks if the error is of a specific type
func (e *PErrors) Is(target error) bool {
	var t *PErrors
	ok := errors.As(target, &t)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// WithError wraps an existing error
func (e *PErrors) WithError(err error) *PErrors {
	return &PErrors{
		Code:    e.Code,
		Message: e.Message,
		Details: e.Details,
		Err:     err,
	}
}

// Is checks if the error is of a specific type
func Is(err, target error) bool {
	return errors.Is(err, target)
}
