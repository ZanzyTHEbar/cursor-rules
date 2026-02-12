// Package errors provides a unified error API built on github.com/ZanzyTHEbar/errbuilder-go.
// All error creation and wrapping in the codebase should go through this package so that
// errors have consistent codes, messages, and support errors.Is/errors.As.
package errors

import (
	"fmt"

	"github.com/ZanzyTHEbar/errbuilder-go"
)

// ErrCode is the error code type (re-exported from errbuilder-go).
type ErrCode = errbuilder.ErrCode

// Error codes (re-exported from errbuilder-go for single-import usage).
const (
	CodeCanceled           = errbuilder.CodeCanceled
	CodeUnknown            = errbuilder.CodeUnknown
	CodeInvalidArgument    = errbuilder.CodeInvalidArgument
	CodeDeadlineExceeded   = errbuilder.CodeDeadlineExceeded
	CodeNotFound           = errbuilder.CodeNotFound
	CodeAlreadyExists      = errbuilder.CodeAlreadyExists
	CodePermissionDenied   = errbuilder.CodePermissionDenied
	CodeResourceExhausted  = errbuilder.CodeResourceExhausted
	CodeFailedPrecondition = errbuilder.CodeFailedPrecondition
	CodeAborted            = errbuilder.CodeAborted
	CodeOutOfRange         = errbuilder.CodeOutOfRange
	CodeUnimplemented      = errbuilder.CodeUnimplemented
	CodeInternal           = errbuilder.CodeInternal
	CodeUnavailable        = errbuilder.CodeUnavailable
	CodeDataLoss           = errbuilder.CodeDataLoss
	CodeUnauthenticated    = errbuilder.CodeUnauthenticated
)

// Wrap wraps cause with the given code and message, producing an error that
// implements the error interface and Unwrap() for errors.Is/errors.As.
func Wrap(cause error, code ErrCode, msg string) error {
	if cause == nil {
		return nil
	}
	return errbuilder.New().
		WithCause(cause).
		WithCode(code).
		WithMsg(msg)
}

// Wrapf wraps cause with the given code and a formatted message.
func Wrapf(cause error, code ErrCode, format string, args ...interface{}) error {
	if cause == nil {
		return nil
	}
	return errbuilder.New().
		WithCause(cause).
		WithCode(code).
		WithMsg(fmt.Sprintf(format, args...))
}

// New builds an error with the given code and message (no cause).
func New(code ErrCode, msg string) error {
	return errbuilder.New().
		WithCode(code).
		WithMsg(msg)
}

// Newf builds an error with the given code and formatted message.
func Newf(code ErrCode, format string, args ...interface{}) error {
	return errbuilder.New().
		WithCode(code).
		WithMsg(fmt.Sprintf(format, args...))
}

// CodeOf returns the error's code if it is or wraps an errbuilder error, otherwise CodeUnknown.
func CodeOf(err error) ErrCode {
	return errbuilder.CodeOf(err)
}

// GenericErr builds an internal error with message and cause (convenience for errbuilder.GenericErr).
func GenericErr(msg string, cause error) error {
	return errbuilder.GenericErr(msg, cause)
}

// NotFoundErr builds a not-found error with cause (convenience for errbuilder.NotFoundErr).
func NotFoundErr(cause error) error {
	return errbuilder.NotFoundErr(cause)
}
