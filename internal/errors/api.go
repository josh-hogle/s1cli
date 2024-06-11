package errors

import (
	"fmt"

	"go.joshhogle.dev/errorx"
)

type S1ClientError struct {
	*errorx.BaseError

	// unexported variables
	msg string
}

// NewS1ClientError creates a new S1ClientError error.
func NewS1ClientError(msg string, err error) *S1ClientRequestError {
	return &S1ClientRequestError{
		BaseError: errorx.NewBaseError(S1ClientErrorCode, err),
		msg:       msg,
	}
}

// Error returns the string version of the error.
func (e *S1ClientError) Error() string {
	return fmt.Sprintf("%s : %s", e.msg, e.InternalError().Error())
}

// Msg returns just the message associated with the error.
func (e *S1ClientError) Msg() string {
	return e.msg
}

type S1ClientRequestError struct {
	*errorx.BaseError

	// unexported variables
	method string
	msg    string
	url    string
}

// NewS1ClientRequestError creates a new S1ClientRequestError error.
func NewS1ClientRequestError(method, url, msg string, err error) *S1ClientRequestError {
	return &S1ClientRequestError{
		BaseError: errorx.NewBaseError(S1ClientRequestErrorCode, err),
		method:    method,
		msg:       msg,
		url:       url,
	}
}

// Error returns the string version of the error.
func (e *S1ClientRequestError) Error() string {
	return fmt.Sprintf("%s %s | %s : %s", e.method, e.url, e.msg, e.InternalError().Error())
}

// Method returns just the HTTP method associated with the error.
func (e *S1ClientRequestError) Method() string {
	return e.method
}

// Msg returns just the message associated with the error.
func (e *S1ClientRequestError) Msg() string {
	return e.msg
}

// URL returns just the URL of the API called that is associated with the error.
func (e *S1ClientRequestError) URL() string {
	return e.url
}
