package errors

import (
	"fmt"

	"go.joshhogle.dev/errorx"
)

// None indicates there is no error at all.
type None struct{}

// NewNone creates a new None error.
func NewNone() *None {
	return &None{}
}

// Attrs returns a map of attributes associated with the error.
//
// This function always returns an empty map.
func (e *None) Attrs() map[string]any {
	return map[string]any{}
}

// Code returns the corresponding error code.
func (e *None) Code() int {
	return NoneCode
}

// Error returns the string version of the error.
func (e *None) Error() string {
	return "the command completed successfully"
}

// File always returns an empty string since there was no error.
func (e *None) File() string {
	return ""
}

// InternalError returns the internal error object.
func (e *None) InternalError() error {
	return nil
}

// Line always returns 0 since there was no error.
func (e *None) Line() int {
	return 0
}

// Method always returns an empty string since there was no error.
func (e *None) Method() string {
	return ""
}

// NestedErrors returns the list of nested errors associated with the error.
//
// This function always returns an empty list.
func (e *None) NestedErrors() []errorx.Error {
	return []errorx.Error{}
}

// UsageError indicates there was a usage error.
type UsageError struct {
	*errorx.BaseError
}

// NewUsageError creates a new UsageError error.
func NewUsageError(err error) *UsageError {
	return &UsageError{
		BaseError: errorx.NewBaseError(UsageErrorCode, err),
	}
}

// NewUsageErrorWithCaller creates a new UsageError error with caller information.
func NewUsageErrorWithCaller(err error) *UsageError {
	return &UsageError{
		BaseError: errorx.NewBaseErrorWithCaller(UsageErrorCode, err, 0),
	}
}

// Error returns the string version of the error.
func (e *UsageError) Error() string {
	return e.InternalError().Error()
}

// GeneralFailure indicates there was a general system error.
type GeneralFailure struct {
	*errorx.BaseError

	// unexported variables
	msg string
}

// NewGeneralFailure creates a new GeneralFailure error.
func NewGeneralFailure(msg string, err error) *GeneralFailure {
	return &GeneralFailure{
		BaseError: errorx.NewBaseError(UsageErrorCode, err),
		msg:       msg,
	}
}

// NewGeneralFailureWithCaller creates a new GeneralFailure error with caller information.
func NewGeneralFailureWithCaller(msg string, err error) *GeneralFailure {
	return &GeneralFailure{
		BaseError: errorx.NewBaseErrorWithCaller(UsageErrorCode, err, 0),
		msg:       msg,
	}
}

// Error returns the string version of the error.
func (e *GeneralFailure) Error() string {
	return fmt.Sprintf("%s: %s", e.msg, e.InternalError().Error())
}

// Msg returns just the message associated with the error.
func (e *GeneralFailure) Msg() string {
	return e.msg
}
