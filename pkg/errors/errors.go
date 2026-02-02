// Package errors provides structured error handling for Core applications.
//
// Deprecated: Use pkg/log instead. This package is maintained for backward
// compatibility and will be removed in a future version. All error handling
// functions are now available in pkg/log:
//
//	// Instead of:
//	import "github.com/host-uk/core/pkg/errors"
//	err := errors.E("op", "msg", cause)
//
//	// Use:
//	import "github.com/host-uk/core/pkg/log"
//	err := log.E("op", "msg", cause)
//
// Migration guide:
//   - errors.Error -> log.Err
//   - errors.E -> log.E
//   - errors.Wrap -> log.Wrap
//   - errors.WrapCode -> log.WrapCode
//   - errors.Code -> log.NewCode
//   - errors.New -> log.NewError
//   - errors.Is -> log.Is
//   - errors.As -> log.As
//   - errors.Join -> log.Join
//   - errors.Op -> log.Op
//   - errors.ErrCode -> log.ErrCode
//   - errors.Message -> log.Message
//   - errors.Root -> log.Root
package errors

import (
	"github.com/host-uk/core/pkg/log"
)

// Error represents a structured error with operational context.
//
// Deprecated: Use log.Err instead.
type Error = log.Err

// E creates a new Error with operation context.
//
// Deprecated: Use log.E instead.
func E(op, msg string, err error) error {
	return log.E(op, msg, err)
}

// Wrap wraps an error with operation context.
// Returns nil if err is nil.
//
// Deprecated: Use log.Wrap instead.
func Wrap(err error, op, msg string) error {
	return log.Wrap(err, op, msg)
}

// WrapCode wraps an error with operation context and an error code.
//
// Deprecated: Use log.WrapCode instead.
func WrapCode(err error, code, op, msg string) error {
	return log.WrapCode(err, code, op, msg)
}

// Code creates an error with just a code and message.
//
// Deprecated: Use log.NewCode instead.
func Code(code, msg string) error {
	return log.NewCode(code, msg)
}

// --- Standard library wrappers ---

// Is reports whether any error in err's tree matches target.
//
// Deprecated: Use log.Is instead.
func Is(err, target error) bool {
	return log.Is(err, target)
}

// As finds the first error in err's tree that matches target.
//
// Deprecated: Use log.As instead.
func As(err error, target any) bool {
	return log.As(err, target)
}

// New returns an error with the given text.
//
// Deprecated: Use log.NewError instead.
func New(text string) error {
	return log.NewError(text)
}

// Join returns an error that wraps the given errors.
//
// Deprecated: Use log.Join instead.
func Join(errs ...error) error {
	return log.Join(errs...)
}

// --- Helper functions ---

// Op extracts the operation from an error, or empty string if not an Error.
//
// Deprecated: Use log.Op instead.
func Op(err error) string {
	return log.Op(err)
}

// ErrCode extracts the error code, or empty string if not set.
//
// Deprecated: Use log.ErrCode instead.
func ErrCode(err error) string {
	return log.ErrCode(err)
}

// Message extracts the message from an error.
// For Error types, returns Msg; otherwise returns err.Error().
//
// Deprecated: Use log.Message instead.
func Message(err error) string {
	return log.Message(err)
}

// Root returns the deepest error in the chain.
//
// Deprecated: Use log.Root instead.
func Root(err error) error {
	return log.Root(err)
}
