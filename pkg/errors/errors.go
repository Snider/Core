// Package errors provides structured error handling for Core applications.
//
// Errors include operational context (what was being done) and support
// error wrapping for debugging while keeping user-facing messages clean:
//
//	err := errors.E("user.Create", "email already exists", nil)
//	err := errors.Wrap(dbErr, "user.Create", "failed to save user")
//
//	// Check error types
//	if errors.Is(err, sql.ErrNoRows) { ... }
//
//	// Extract operation
//	var e *errors.Error
//	if errors.As(err, &e) {
//	    fmt.Println("Operation:", e.Op)
//	}
package errors

import (
	stderrors "errors"
	"fmt"
)

// Error represents a structured error with operational context.
type Error struct {
	Op   string // Operation being performed (e.g., "user.Create")
	Msg  string // Human-readable message
	Err  error  // Underlying error (optional)
	Code string // Error code for i18n/categorisation (optional)
}

// E creates a new Error with operation context.
//
//	err := errors.E("config.Load", "file not found", os.ErrNotExist)
//	err := errors.E("api.Call", "rate limited", nil)
func E(op, msg string, err error) error {
	return &Error{Op: op, Msg: msg, Err: err}
}

// Wrap wraps an error with operation context.
// Returns nil if err is nil.
//
//	return errors.Wrap(err, "db.Query", "failed to fetch user")
func Wrap(err error, op, msg string) error {
	if err == nil {
		return nil
	}
	return &Error{Op: op, Msg: msg, Err: err}
}

// WrapCode wraps an error with operation context and an error code.
//
//	return errors.WrapCode(err, "ERR_NOT_FOUND", "user.Get", "user not found")
func WrapCode(err error, code, op, msg string) error {
	if err == nil && code == "" {
		return nil
	}
	return &Error{Op: op, Msg: msg, Err: err, Code: code}
}

// Code creates an error with just a code and message.
//
//	return errors.Code("ERR_VALIDATION", "invalid email format")
func Code(code, msg string) error {
	return &Error{Code: code, Msg: msg}
}

// Error returns the error message.
func (e *Error) Error() string {
	if e.Op != "" && e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Op, e.Msg, e.Err)
	}
	if e.Op != "" {
		return fmt.Sprintf("%s: %s", e.Op, e.Msg)
	}
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Msg, e.Err)
	}
	return e.Msg
}

// Unwrap returns the underlying error.
func (e *Error) Unwrap() error {
	return e.Err
}

// --- Standard library wrappers ---

// Is reports whether any error in err's tree matches target.
func Is(err, target error) bool {
	return stderrors.Is(err, target)
}

// As finds the first error in err's tree that matches target.
func As(err error, target any) bool {
	return stderrors.As(err, target)
}

// New returns an error with the given text.
func New(text string) error {
	return stderrors.New(text)
}

// Join returns an error that wraps the given errors.
func Join(errs ...error) error {
	return stderrors.Join(errs...)
}

// --- Helper functions ---

// Op extracts the operation from an error, or empty string if not an Error.
func Op(err error) string {
	var e *Error
	if As(err, &e) {
		return e.Op
	}
	return ""
}

// ErrCode extracts the error code, or empty string if not set.
func ErrCode(err error) string {
	var e *Error
	if As(err, &e) {
		return e.Code
	}
	return ""
}

// Message extracts the message from an error.
// For Error types, returns Msg; otherwise returns err.Error().
func Message(err error) string {
	if err == nil {
		return ""
	}
	var e *Error
	if As(err, &e) {
		return e.Msg
	}
	return err.Error()
}

// Root returns the deepest error in the chain.
func Root(err error) error {
	for {
		unwrapped := stderrors.Unwrap(err)
		if unwrapped == nil {
			return err
		}
		err = unwrapped
	}
}
