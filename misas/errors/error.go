package errors

import (
	"errors"
)

// NotFoundCode expresses that a resource was not found.
const NotFoundCode = "not_found"

// AlreadyExists expresses that a resource already existed when it was expected not to.
const AlreadyExists = "already_exists"

// SystemError Error represents an error
type SystemError interface {
	// Code returns an error code that can be used by other parts of the system so they can craft user-friendly error messages.
	Code() string

	// Error Implements the standard error
	Error() string

	// Unwrap allows to make errors compatible with go 1.13+ errors
	Unwrap() error
}

// Error is an implementation of a SystemError to easily define custom errors, without requiring a new type.
type Error struct {
	Message string
	code    string

	// returns the cause of the error.
	Cause error
}

func (e Error) Code() string {
	return e.code
}

func (e Error) Error() string {
	return e.Message
}

func (e Error) Unwrap() error {
	return e.Cause
}

// New returns a new error with code.
func New(code string) Error {
	return Error{
		code: code,
	}
}

// NewWithMessage returns a new error with code and Message.
func NewWithMessage(code string, message string) Error {
	return Error{
		code:    code,
		Message: message,
	}
}

// Wrap returns a new error with a given code that wraps another error .
func Wrap(err error, code string) Error {
	return Error{
		Message: "",
		code:    code,
		Cause:   err,
	}
}

// WrapWithMessage returns a new error with  given code and Message that wraps another error.
func WrapWithMessage(err error, code string, message string) Error {
	return Error{
		Message: message,
		code:    code,
		Cause:   err,
	}
}

// HasCode indicates if a given error or any error it wraps has a certain code.
func HasCode(err error, code string) bool {
	return hasCode(err, code, true)
}

// HasCodeStrict indicates if a given error has a certain code, without unwrapping.
func HasCodeStrict(err error, code string) bool {
	return hasCode(err, code, false)
}

func hasCode(err error, code string, unwrap bool) bool {
	e, ok := err.(SystemError)
	if !ok {
		return false
	}

	if e.Code() == code {
		return true
	}

	if unwrap {
		return hasCode(e.Unwrap(), code, unwrap)
	}

	return false
}

// RootCause returns the root cause of this error unwrapping until the first one.
// If there is no root cause in any wrapped errors, the error is returned.
func RootCause(err error) error {
	if err, ok := err.(interface{ Unwrap() error }); ok {
		previous := err.Unwrap()
		if previous != nil {
			return RootCause(previous)
		}
	}

	// no previous this is the root cause
	return err
}

// As see errors.As for more details
func As(err error, target any) bool {
	//goland:noinspection GoErrorsAs
	return errors.As(err, target)
}

// Is see errors.Is for more details
func Is(err, target error) bool {
	return errors.Is(err, target)
}
