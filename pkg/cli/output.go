package cli

import (
	"fmt"

	"github.com/host-uk/core/pkg/i18n"
)

// Blank prints an empty line.
func Blank() {
	fmt.Println()
}

// Echo translates a key via i18n.T and prints with newline.
// No automatic styling - use Success/Error/Warn/Info for styled output.
func Echo(key string, args ...any) {
	fmt.Println(i18n.T(key, args...))
}

// Print outputs formatted text (no newline).
// Glyph shortcodes like :check: are converted.
func Print(format string, args ...any) {
	fmt.Print(compileGlyphs(fmt.Sprintf(format, args...)))
}

// Println outputs formatted text with newline.
// Glyph shortcodes like :check: are converted.
func Println(format string, args ...any) {
	fmt.Println(compileGlyphs(fmt.Sprintf(format, args...)))
}

// Success prints a success message with checkmark (green).
func Success(msg string) {
	fmt.Println(SuccessStyle.Render(Glyph(":check:") + " " + msg))
}

// Successf prints a formatted success message.
func Successf(format string, args ...any) {
	Success(fmt.Sprintf(format, args...))
}

// Error prints an error message with cross (red).
func Error(msg string) {
	fmt.Println(ErrorStyle.Render(Glyph(":cross:") + " " + msg))
}

// Errorf prints a formatted error message.
func Errorf(format string, args ...any) {
	Error(fmt.Sprintf(format, args...))
}

// Warn prints a warning message with warning symbol (amber).
func Warn(msg string) {
	fmt.Println(WarningStyle.Render(Glyph(":warn:") + " " + msg))
}

// Warnf prints a formatted warning message.
func Warnf(format string, args ...any) {
	Warn(fmt.Sprintf(format, args...))
}

// Info prints an info message with info symbol (blue).
func Info(msg string) {
	fmt.Println(InfoStyle.Render(Glyph(":info:") + " " + msg))
}

// Infof prints a formatted info message.
func Infof(format string, args ...any) {
	Info(fmt.Sprintf(format, args...))
}

// Dim prints dimmed text.
func Dim(msg string) {
	fmt.Println(DimStyle.Render(msg))
}

// Progress prints a progress indicator that overwrites the current line.
// Uses i18n.Progress for gerund form ("Checking...").
func Progress(verb string, current, total int, item ...string) {
	msg := i18n.Progress(verb)
	if len(item) > 0 && item[0] != "" {
		fmt.Printf("\033[2K\r%s %d/%d %s", DimStyle.Render(msg), current, total, item[0])
	} else {
		fmt.Printf("\033[2K\r%s %d/%d", DimStyle.Render(msg), current, total)
	}
}

// ProgressDone clears the progress line.
func ProgressDone() {
	fmt.Print("\033[2K\r")
}

// Label prints a "Label: value" line.
func Label(word, value string) {
	fmt.Printf("%s %s\n", KeyStyle.Render(i18n.Label(word)), value)
}

// Scanln reads from stdin.
func Scanln(a ...any) (int, error) {
	return fmt.Scanln(a...)
}