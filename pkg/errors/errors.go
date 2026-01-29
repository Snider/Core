// Package errors provides standardized error handling for the Core CLI.
package errors

import "fmt"

// Error represents a standardized error with operational context.
type Error struct {
	Op  string // Operation being performed
	Msg string // Human-readable message
	Err error  // Underlying error
}

// E creates a new Error with operation context.
func E(op, msg string, err error) error {
	if err == nil {
		return &Error{Op: op, Msg: msg}
	}
	return &Error{Op: op, Msg: msg, Err: err}
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Op, e.Msg, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Op, e.Msg)
}

func (e *Error) Unwrap() error { return e.Err }
