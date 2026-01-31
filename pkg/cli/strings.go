package cli

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// ─────────────────────────────────────────────────────────────────────────────
// String Formatting (replace fmt.Sprintf)
// ─────────────────────────────────────────────────────────────────────────────

// Sprintf formats a string.
// This is a direct replacement for fmt.Sprintf.
func Sprintf(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}

// Sprint formats using the default formats for its operands.
// This is a direct replacement for fmt.Sprint.
func Sprint(args ...any) string {
	return fmt.Sprint(args...)
}

// ─────────────────────────────────────────────────────────────────────────────
// Styled String Functions
// ─────────────────────────────────────────────────────────────────────────────

// Styled returns text formatted with a style.
// Example: cli.Styled(cli.Style.Success, "Done!")
func Styled(style lipgloss.Style, text string) string {
	return style.Render(text)
}

// Styledf returns formatted text with a style.
// Example: cli.Styledf(cli.Style.Success, "Processed %d items", count)
func Styledf(style lipgloss.Style, format string, args ...any) string {
	return style.Render(fmt.Sprintf(format, args...))
}

// ─────────────────────────────────────────────────────────────────────────────
// Pre-styled Formatting Functions
// ─────────────────────────────────────────────────────────────────────────────

// SuccessStr returns a success-styled string with checkmark.
func SuccessStr(msg string) string {
	return SuccessStyle.Render(SymbolCheck + " " + msg)
}

// ErrorStr returns an error-styled string with cross.
func ErrorStr(msg string) string {
	return ErrorStyle.Render(SymbolCross + " " + msg)
}

// WarningStr returns a warning-styled string with warning symbol.
func WarningStr(msg string) string {
	return WarningStyle.Render(SymbolWarning + " " + msg)
}

// InfoStr returns an info-styled string with info symbol.
func InfoStr(msg string) string {
	return InfoStyle.Render(SymbolInfo + " " + msg)
}

// DimStr returns a dim-styled string.
func DimStr(msg string) string {
	return DimStyle.Render(msg)
}

// BoldStr returns a bold-styled string.
func BoldStr(msg string) string {
	return BoldStyle.Render(msg)
}

// ─────────────────────────────────────────────────────────────────────────────
// Numeric Formatting
// ─────────────────────────────────────────────────────────────────────────────

// Itoa converts an integer to a string.
// This is a convenience function similar to strconv.Itoa.
func Itoa(n int) string {
	return fmt.Sprintf("%d", n)
}

// Itoa64 converts an int64 to a string.
func Itoa64(n int64) string {
	return fmt.Sprintf("%d", n)
}
