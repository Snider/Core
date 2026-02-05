// Package log provides structured logging and error handling for Core applications.
//
// This file implements structured error types and combined log-and-return helpers
// that simplify common error handling patterns.

package log

import (
	"errors"
	"fmt"
)

// Err represents a structured error with operational context.
// It implements the error interface and supports unwrapping.
type Err struct {
	Op   string // Operation being performed (e.g., "user.Save")
	Msg  string // Human-readable message
	Err  error  // Underlying error (optional)
	Code string // Error code (optional, e.g., "VALIDATION_FAILED")
}

// Error implements the error interface.
func (e *Err) Error() string {
	var prefix string
	if e.Op != "" {
		prefix = e.Op + ": "
	}
	if e.Err != nil {
		if e.Code != "" {
			return fmt.Sprintf("%s%s [%s]: %v", prefix, e.Msg, e.Code, e.Err)
		}
		return fmt.Sprintf("%s%s: %v", prefix, e.Msg, e.Err)
	}
	if e.Code != "" {
		return fmt.Sprintf("%s%s [%s]", prefix, e.Msg, e.Code)
	}
	return fmt.Sprintf("%s%s", prefix, e.Msg)
}

// Unwrap returns the underlying error for use with errors.Is and errors.As.
func (e *Err) Unwrap() error {
	return e.Err
}

// --- Error Creation Functions ---

// E creates a new Err with operation context.
// The underlying error can be nil for creating errors without a cause.
//
// Example:
//
//	return log.E("user.Save", "failed to save user", err)
//	return log.E("api.Call", "rate limited", nil)  // No underlying cause
func E(op, msg string, err error) error {
	return &Err{Op: op, Msg: msg, Err: err}
}

// Wrap wraps an error with operation context.
// Returns nil if err is nil, to support conditional wrapping.
// Preserves error Code if the wrapped error is an *Err.
//
// Example:
//
//	return log.Wrap(err, "db.Query", "database query failed")
func Wrap(err error, op, msg string) error {
	if err == nil {
		return nil
	}
	// Preserve Code from wrapped *Err
	var logErr *Err
	if As(err, &logErr) && logErr.Code != "" {
		return &Err{Op: op, Msg: msg, Err: err, Code: logErr.Code}
	}
	return &Err{Op: op, Msg: msg, Err: err}
}

// WrapCode wraps an error with operation context and error code.
// Returns nil only if both err is nil AND code is empty.
// Useful for API errors that need machine-readable codes.
//
// Example:
//
//	return log.WrapCode(err, "VALIDATION_ERROR", "user.Validate", "invalid email")
func WrapCode(err error, code, op, msg string) error {
	if err == nil && code == "" {
		return nil
	}
	return &Err{Op: op, Msg: msg, Err: err, Code: code}
}

// NewCode creates an error with just code and message (no underlying error).
// Useful for creating sentinel errors with codes.
//
// Example:
//
//	var ErrNotFound = log.NewCode("NOT_FOUND", "resource not found")
func NewCode(code, msg string) error {
	return &Err{Msg: msg, Code: code}
}

// --- Standard Library Wrappers ---

// Is reports whether any error in err's tree matches target.
// Wrapper around errors.Is for convenience.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's tree that matches target.
// Wrapper around errors.As for convenience.
func As(err error, target any) bool {
	return errors.As(err, target)
}

// NewError creates a simple error with the given text.
// Wrapper around errors.New for convenience.
func NewError(text string) error {
	return errors.New(text)
}

// Join combines multiple errors into one.
// Wrapper around errors.Join for convenience.
func Join(errs ...error) error {
	return errors.Join(errs...)
}

// --- Error Introspection Helpers ---

// Op extracts the operation name from an error.
// Returns empty string if the error is not an *Err.
func Op(err error) string {
	var e *Err
	if As(err, &e) {
		return e.Op
	}
	return ""
}

// ErrCode extracts the error code from an error.
// Returns empty string if the error is not an *Err or has no code.
func ErrCode(err error) string {
	var e *Err
	if As(err, &e) {
		return e.Code
	}
	return ""
}

// Message extracts the message from an error.
// Returns the error's Error() string if not an *Err.
func Message(err error) string {
	if err == nil {
		return ""
	}
	var e *Err
	if As(err, &e) {
		return e.Msg
	}
	return err.Error()
}

// Root returns the root cause of an error chain.
// Unwraps until no more wrapped errors are found.
func Root(err error) error {
	if err == nil {
		return nil
	}
	for {
		unwrapped := errors.Unwrap(err)
		if unwrapped == nil {
			return err
		}
		err = unwrapped
	}
}

// StackTrace returns the logical stack trace (chain of operations) from an error.
// It returns an empty slice if no operational context is found.
func StackTrace(err error) []string {
	var stack []string
	for err != nil {
		if e, ok := err.(*Err); ok {
			if e.Op != "" {
				stack = append(stack, e.Op)
			}
		}
		err = errors.Unwrap(err)
	}
	return stack
}

// FormatStackTrace returns a pretty-printed logical stack trace.
func FormatStackTrace(err error) string {
	stack := StackTrace(err)
	if len(stack) == 0 {
		return ""
	}
	var res string
	for i, op := range stack {
		if i > 0 {
			res += " -> "
		}
		res += op
	}
	return res
}

// --- Combined Log-and-Return Helpers ---

// LogError logs an error at Error level and returns a wrapped error.
// Reduces boilerplate in error handling paths.
//
// Example:
//
//	// Before
//	if err != nil {
//	    log.Error("failed to save", "err", err)
//	    return errors.Wrap(err, "user.Save", "failed to save")
//	}
//
//	// After
//	if err != nil {
//	    return log.LogError(err, "user.Save", "failed to save")
//	}
func LogError(err error, op, msg string) error {
	if err == nil {
		return nil
	}
	wrapped := Wrap(err, op, msg)
	defaultLogger.Error(msg, "op", op, "err", err)
	return wrapped
}

// LogWarn logs at Warn level and returns a wrapped error.
// Use for recoverable errors that should be logged but not treated as critical.
//
// Example:
//
//	return log.LogWarn(err, "cache.Get", "cache miss, falling back to db")
func LogWarn(err error, op, msg string) error {
	if err == nil {
		return nil
	}
	wrapped := Wrap(err, op, msg)
	defaultLogger.Warn(msg, "op", op, "err", err)
	return wrapped
}

// Must panics if err is not nil, logging first.
// Use for errors that should never happen and indicate programmer error.
//
// Example:
//
//	log.Must(Initialize(), "app", "startup failed")
func Must(err error, op, msg string) {
	if err != nil {
		defaultLogger.Error(msg, "op", op, "err", err)
		panic(Wrap(err, op, msg))
	}
}
