package cli

import (
	"fmt"
	"strings"

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

// Text prints arguments like fmt.Println, but handling glyphs.
func Text(args ...any) {
	fmt.Println(compileGlyphs(fmt.Sprint(args...)))
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

// Task prints a task header: "[label] message"
//
//	cli.Task("php", "Running tests...")  // [php] Running tests...
//	cli.Task("go", i18n.Progress("build"))  // [go] Building...
func Task(label, message string) {
	fmt.Printf("%s %s\n\n", DimStyle.Render("["+label+"]"), message)
}

// Section prints a section header: "── SECTION ──"
//
//	cli.Section("audit")  // ── AUDIT ──
func Section(name string) {
	header := "── " + strings.ToUpper(name) + " ──"
	fmt.Println(AccentStyle.Render(header))
}

// Hint prints a labelled hint: "label: message"
//
//	cli.Hint("install", "composer require vimeo/psalm")
//	cli.Hint("fix", "core php fmt --fix")
func Hint(label, message string) {
	fmt.Printf("  %s %s\n", DimStyle.Render(label+":"), message)
}

// Severity prints a severity-styled message.
//
//	cli.Severity("critical", "SQL injection")  // red, bold
//	cli.Severity("high", "XSS vulnerability")  // orange
//	cli.Severity("medium", "Missing CSRF")     // amber
//	cli.Severity("low", "Debug enabled")       // gray
func Severity(level, message string) {
	var style *AnsiStyle
	switch strings.ToLower(level) {
	case "critical":
		style = NewStyle().Bold().Foreground(ColourRed500)
	case "high":
		style = NewStyle().Bold().Foreground(ColourOrange500)
	case "medium":
		style = NewStyle().Foreground(ColourAmber500)
	case "low":
		style = NewStyle().Foreground(ColourGray500)
	default:
		style = DimStyle
	}
	fmt.Printf("  %s %s\n", style.Render("["+level+"]"), message)
}

// Result prints a result line: "✓ message" or "✗ message"
//
//	cli.Result(passed, "All tests passed")
//	cli.Result(false, "3 tests failed")
func Result(passed bool, message string) {
	if passed {
		Success(message)
	} else {
		Error(message)
	}
}