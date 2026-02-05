package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/host-uk/core/pkg/i18n"
)

// ─────────────────────────────────────────────────────────────────────────────
// Error Creation (replace fmt.Errorf)
// ─────────────────────────────────────────────────────────────────────────────

// Err creates a new error from a format string.
// This is a direct replacement for fmt.Errorf.
func Err(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

// Wrap wraps an error with a message.
// Returns nil if err is nil.
//
//	return cli.Wrap(err, "load config")  // "load config: <original error>"
func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}

// WrapVerb wraps an error using i18n grammar for "Failed to verb subject".
// Uses the i18n.ActionFailed function for proper grammar composition.
// Returns nil if err is nil.
//
//	return cli.WrapVerb(err, "load", "config")  // "Failed to load config: <original error>"
func WrapVerb(err error, verb, subject string) error {
	if err == nil {
		return nil
	}
	msg := i18n.ActionFailed(verb, subject)
	return fmt.Errorf("%s: %w", msg, err)
}

// WrapAction wraps an error using i18n grammar for "Failed to verb".
// Uses the i18n.ActionFailed function for proper grammar composition.
// Returns nil if err is nil.
//
//	return cli.WrapAction(err, "connect")  // "Failed to connect: <original error>"
func WrapAction(err error, verb string) error {
	if err == nil {
		return nil
	}
	msg := i18n.ActionFailed(verb, "")
	return fmt.Errorf("%s: %w", msg, err)
}

// ─────────────────────────────────────────────────────────────────────────────
// Error Helpers
// ─────────────────────────────────────────────────────────────────────────────

// Is reports whether any error in err's tree matches target.
// This is a re-export of errors.Is for convenience.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's tree that matches target.
// This is a re-export of errors.As for convenience.
func As(err error, target any) bool {
	return errors.As(err, target)
}

// Join returns an error that wraps the given errors.
// This is a re-export of errors.Join for convenience.
func Join(errs ...error) error {
	return errors.Join(errs...)
}

// ExitError represents an error that should cause the CLI to exit with a specific code.
type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	if e.Err == nil {
		return ""
	}
	return e.Err.Error()
}

func (e *ExitError) Unwrap() error {
	return e.Err
}

// Exit creates a new ExitError with the given code and error.
// Use this to return an error from a command with a specific exit code.
func Exit(code int, err error) error {
	if err == nil {
		return nil
	}
	return &ExitError{Code: code, Err: err}
}

// ─────────────────────────────────────────────────────────────────────────────
// Fatal Functions (Deprecated - return error from command instead)
// ─────────────────────────────────────────────────────────────────────────────

// Fatal prints an error message to stderr, logs it, and exits with code 1.
//
// Deprecated: return an error from the command instead.
func Fatal(err error) {
	if err != nil {
		LogError("Fatal error", "err", err)
		fmt.Fprintln(os.Stderr, ErrorStyle.Render(Glyph(":cross:")+" "+err.Error()))
		os.Exit(1)
	}
}

// Fatalf prints a formatted error message to stderr, logs it, and exits with code 1.
//
// Deprecated: return an error from the command instead.
func Fatalf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	LogError("Fatal error", "msg", msg)
	fmt.Fprintln(os.Stderr, ErrorStyle.Render(Glyph(":cross:")+" "+msg))
	os.Exit(1)
}

// FatalWrap prints a wrapped error message to stderr, logs it, and exits with code 1.
// Does nothing if err is nil.
//
// Deprecated: return an error from the command instead.
//
//	cli.FatalWrap(err, "load config")  // Prints "✗ load config: <error>" and exits
func FatalWrap(err error, msg string) {
	if err == nil {
		return
	}
	LogError("Fatal error", "msg", msg, "err", err)
	fullMsg := fmt.Sprintf("%s: %v", msg, err)
	fmt.Fprintln(os.Stderr, ErrorStyle.Render(Glyph(":cross:")+" "+fullMsg))
	os.Exit(1)
}

// FatalWrapVerb prints a wrapped error using i18n grammar to stderr, logs it, and exits with code 1.
// Does nothing if err is nil.
//
// Deprecated: return an error from the command instead.
//
//	cli.FatalWrapVerb(err, "load", "config")  // Prints "✗ Failed to load config: <error>" and exits
func FatalWrapVerb(err error, verb, subject string) {
	if err == nil {
		return
	}
	msg := i18n.ActionFailed(verb, subject)
	LogError("Fatal error", "msg", msg, "err", err, "verb", verb, "subject", subject)
	fullMsg := fmt.Sprintf("%s: %v", msg, err)
	fmt.Fprintln(os.Stderr, ErrorStyle.Render(Glyph(":cross:")+" "+fullMsg))
	os.Exit(1)
}
