// Package shared provides common utilities and styles for CLI commands.
//
// This package contains:
//   - Terminal styling using lipgloss with Tailwind colours
//   - Unicode symbols for consistent visual indicators
//   - Helper functions for common output patterns
//   - Git and GitHub CLI utilities
package shared

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ─────────────────────────────────────────────────────────────────────────────
// Tailwind Colour Palette
// ─────────────────────────────────────────────────────────────────────────────

// Tailwind colours for consistent theming across the CLI.
const (
	ColourBlue50    = lipgloss.Color("#eff6ff")
	ColourBlue100   = lipgloss.Color("#dbeafe")
	ColourBlue200   = lipgloss.Color("#bfdbfe")
	ColourBlue300   = lipgloss.Color("#93c5fd")
	ColourBlue400   = lipgloss.Color("#60a5fa")
	ColourBlue500   = lipgloss.Color("#3b82f6")
	ColourBlue600   = lipgloss.Color("#2563eb")
	ColourBlue700   = lipgloss.Color("#1d4ed8")
	ColourGreen400  = lipgloss.Color("#4ade80")
	ColourGreen500  = lipgloss.Color("#22c55e")
	ColourGreen600  = lipgloss.Color("#16a34a")
	ColourRed400    = lipgloss.Color("#f87171")
	ColourRed500    = lipgloss.Color("#ef4444")
	ColourRed600    = lipgloss.Color("#dc2626")
	ColourAmber400  = lipgloss.Color("#fbbf24")
	ColourAmber500  = lipgloss.Color("#f59e0b")
	ColourAmber600  = lipgloss.Color("#d97706")
	ColourOrange500 = lipgloss.Color("#f97316")
	ColourViolet400 = lipgloss.Color("#a78bfa")
	ColourViolet500 = lipgloss.Color("#8b5cf6")
	ColourIndigo500 = lipgloss.Color("#6366f1")
	ColourCyan500   = lipgloss.Color("#06b6d4")
	ColourGray50    = lipgloss.Color("#f9fafb")
	ColourGray100   = lipgloss.Color("#f3f4f6")
	ColourGray200   = lipgloss.Color("#e5e7eb")
	ColourGray300   = lipgloss.Color("#d1d5db")
	ColourGray400   = lipgloss.Color("#9ca3af")
	ColourGray500   = lipgloss.Color("#6b7280")
	ColourGray600   = lipgloss.Color("#4b5563")
	ColourGray700   = lipgloss.Color("#374151")
	ColourGray800   = lipgloss.Color("#1f2937")
	ColourGray900   = lipgloss.Color("#111827")
)

// ─────────────────────────────────────────────────────────────────────────────
// Unicode Symbols
// ─────────────────────────────────────────────────────────────────────────────

// Symbols for consistent visual indicators across commands.
const (
	// Status indicators
	SymbolCheck    = "✓" // Success, pass, complete
	SymbolCross    = "✗" // Error, fail, incomplete
	SymbolWarning  = "⚠" // Warning, caution
	SymbolInfo     = "ℹ" // Information
	SymbolQuestion = "?" // Unknown, needs attention
	SymbolSkip     = "○" // Skipped, neutral
	SymbolDot      = "●" // Active, selected
	SymbolCircle   = "◯" // Inactive, unselected

	// Arrows and pointers
	SymbolArrowRight = "→"
	SymbolArrowLeft  = "←"
	SymbolArrowUp    = "↑"
	SymbolArrowDown  = "↓"
	SymbolPointer    = "▶"

	// Decorative
	SymbolBullet    = "•"
	SymbolDash      = "─"
	SymbolPipe      = "│"
	SymbolCorner    = "└"
	SymbolTee       = "├"
	SymbolSeparator = " │ "

	// Progress
	SymbolSpinner = "⠋" // First frame of braille spinner
	SymbolPending = "…"
)

// ─────────────────────────────────────────────────────────────────────────────
// Base Styles
// ─────────────────────────────────────────────────────────────────────────────

// Terminal styles using Tailwind colour palette.
// These are shared across command packages for consistent output.
var (
	// RepoNameStyle highlights repository names (blue, bold).
	RepoNameStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColourBlue500)

	// SuccessStyle indicates successful operations (green, bold).
	SuccessStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColourGreen500)

	// ErrorStyle indicates errors and failures (red, bold).
	ErrorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColourRed500)

	// WarningStyle indicates warnings and cautions (amber, bold).
	WarningStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColourAmber500)

	// InfoStyle for informational messages (blue).
	InfoStyle = lipgloss.NewStyle().
			Foreground(ColourBlue400)

	// DimStyle for secondary/muted text (gray).
	DimStyle = lipgloss.NewStyle().
			Foreground(ColourGray500)

	// MutedStyle for very subtle text (darker gray).
	MutedStyle = lipgloss.NewStyle().
			Foreground(ColourGray600)

	// ValueStyle for data values and output (light gray).
	ValueStyle = lipgloss.NewStyle().
			Foreground(ColourGray200)

	// AccentStyle for highlighted values (cyan).
	AccentStyle = lipgloss.NewStyle().
			Foreground(ColourCyan500)

	// LinkStyle for URLs and clickable references (blue, underlined).
	LinkStyle = lipgloss.NewStyle().
			Foreground(ColourBlue500).
			Underline(true)

	// HeaderStyle for section headers (light gray, bold).
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColourGray200)

	// TitleStyle for command titles (blue, bold).
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColourBlue500)

	// BoldStyle for emphasis without colour.
	BoldStyle = lipgloss.NewStyle().
			Bold(true)

	// CodeStyle for inline code or paths.
	CodeStyle = lipgloss.NewStyle().
			Foreground(ColourGray300).
			Background(ColourGray800).
			Padding(0, 1)

	// KeyStyle for labels/keys in key-value pairs.
	KeyStyle = lipgloss.NewStyle().
			Foreground(ColourGray400)

	// NumberStyle for numeric values.
	NumberStyle = lipgloss.NewStyle().
			Foreground(ColourBlue300)
)

// ─────────────────────────────────────────────────────────────────────────────
// Status Styles (for consistent status indicators)
// ─────────────────────────────────────────────────────────────────────────────

var (
	// StatusPendingStyle for pending/waiting states.
	StatusPendingStyle = lipgloss.NewStyle().Foreground(ColourGray500)

	// StatusRunningStyle for in-progress states.
	StatusRunningStyle = lipgloss.NewStyle().Foreground(ColourBlue500)

	// StatusSuccessStyle for completed/success states.
	StatusSuccessStyle = lipgloss.NewStyle().Foreground(ColourGreen500)

	// StatusErrorStyle for failed/error states.
	StatusErrorStyle = lipgloss.NewStyle().Foreground(ColourRed500)

	// StatusWarningStyle for warning states.
	StatusWarningStyle = lipgloss.NewStyle().Foreground(ColourAmber500)
)

// ─────────────────────────────────────────────────────────────────────────────
// Box Styles (for bordered content)
// ─────────────────────────────────────────────────────────────────────────────

var (
	// BoxStyle creates a rounded border box.
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColourGray600).
			Padding(0, 1)

	// BoxHeaderStyle for box titles.
	BoxHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColourBlue400).
			MarginBottom(1)

	// ErrorBoxStyle for error message boxes.
	ErrorBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColourRed500).
			Padding(0, 1)

	// SuccessBoxStyle for success message boxes.
	SuccessBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColourGreen500).
			Padding(0, 1)
)

// ─────────────────────────────────────────────────────────────────────────────
// Helper Functions
// ─────────────────────────────────────────────────────────────────────────────

// Success returns a styled success message with checkmark.
func Success(msg string) string {
	return fmt.Sprintf("%s %s", SuccessStyle.Render(SymbolCheck), msg)
}

// Error returns a styled error message with cross.
func Error(msg string) string {
	return fmt.Sprintf("%s %s", ErrorStyle.Render(SymbolCross), msg)
}

// Warning returns a styled warning message with warning symbol.
func Warning(msg string) string {
	return fmt.Sprintf("%s %s", WarningStyle.Render(SymbolWarning), msg)
}

// Info returns a styled info message with info symbol.
func Info(msg string) string {
	return fmt.Sprintf("%s %s", InfoStyle.Render(SymbolInfo), msg)
}

// Pending returns a styled pending message with ellipsis.
func Pending(msg string) string {
	return fmt.Sprintf("%s %s", DimStyle.Render(SymbolPending), DimStyle.Render(msg))
}

// Skip returns a styled skipped message with circle.
func Skip(msg string) string {
	return fmt.Sprintf("%s %s", DimStyle.Render(SymbolSkip), DimStyle.Render(msg))
}

// KeyValue returns a styled "key: value" string.
func KeyValue(key, value string) string {
	return fmt.Sprintf("%s %s", KeyStyle.Render(key+":"), value)
}

// KeyValueBold returns a styled "key: value" with bold value.
func KeyValueBold(key, value string) string {
	return fmt.Sprintf("%s %s", KeyStyle.Render(key+":"), BoldStyle.Render(value))
}

// StatusLine creates a pipe-separated status line from parts.
// Each part should be pre-styled.
func StatusLine(parts ...string) string {
	return strings.Join(parts, DimStyle.Render(SymbolSeparator))
}

// StatusPart creates a styled count + label for status lines.
// Example: StatusPart(5, "repos", SuccessStyle) -> "5 repos" in green
func StatusPart(count int, label string, style lipgloss.Style) string {
	return style.Render(fmt.Sprintf("%d", count)) + DimStyle.Render(" "+label)
}

// StatusText creates a styled label for status lines.
// Example: StatusText("clean", SuccessStyle) -> "clean" in green
func StatusText(text string, style lipgloss.Style) string {
	return style.Render(text)
}

// Bullet returns a bulleted item.
func Bullet(text string) string {
	return fmt.Sprintf("  %s %s", DimStyle.Render(SymbolBullet), text)
}

// TreeItem returns a tree item with branch indicator.
func TreeItem(text string, isLast bool) string {
	prefix := SymbolTee
	if isLast {
		prefix = SymbolCorner
	}
	return fmt.Sprintf("  %s %s", DimStyle.Render(prefix), text)
}

// Indent returns text indented by the specified level (2 spaces per level).
func Indent(level int, text string) string {
	return strings.Repeat("  ", level) + text
}

// Separator returns a horizontal separator line.
func Separator(width int) string {
	return DimStyle.Render(strings.Repeat(SymbolDash, width))
}

// Header returns a styled section header with optional separator.
func Header(title string, withSeparator bool) string {
	if withSeparator {
		return fmt.Sprintf("\n%s\n%s", HeaderStyle.Render(title), Separator(len(title)))
	}
	return fmt.Sprintf("\n%s", HeaderStyle.Render(title))
}

// Title returns a styled command/section title.
func Title(text string) string {
	return TitleStyle.Render(text)
}

// Dim returns dimmed text.
func Dim(text string) string {
	return DimStyle.Render(text)
}

// Bold returns bold text.
func Bold(text string) string {
	return BoldStyle.Render(text)
}

// Highlight returns text with accent colour.
func Highlight(text string) string {
	return AccentStyle.Render(text)
}

// Path renders a file path in code style.
func Path(p string) string {
	return CodeStyle.Render(p)
}

// Command renders a command in code style.
func Command(cmd string) string {
	return CodeStyle.Render(cmd)
}

// Number renders a number with styling.
func Number(n int) string {
	return NumberStyle.Render(fmt.Sprintf("%d", n))
}

// ─────────────────────────────────────────────────────────────────────────────
// Table Helpers
// ─────────────────────────────────────────────────────────────────────────────

// TableRow represents a row in a simple table.
type TableRow []string

// Table provides simple table rendering.
type Table struct {
	Headers []string
	Rows    []TableRow
	Widths  []int // Optional: fixed column widths
}

// Render returns the table as a formatted string.
func (t *Table) Render() string {
	if len(t.Rows) == 0 && len(t.Headers) == 0 {
		return ""
	}

	// Calculate column widths if not provided
	widths := t.Widths
	if len(widths) == 0 {
		cols := len(t.Headers)
		if cols == 0 && len(t.Rows) > 0 {
			cols = len(t.Rows[0])
		}
		widths = make([]int, cols)

		for i, h := range t.Headers {
			if len(h) > widths[i] {
				widths[i] = len(h)
			}
		}
		for _, row := range t.Rows {
			for i, cell := range row {
				if i < len(widths) && len(cell) > widths[i] {
					widths[i] = len(cell)
				}
			}
		}
	}

	var sb strings.Builder

	// Headers
	if len(t.Headers) > 0 {
		for i, h := range t.Headers {
			if i > 0 {
				sb.WriteString("  ")
			}
			sb.WriteString(HeaderStyle.Render(padRight(h, widths[i])))
		}
		sb.WriteString("\n")

		// Separator
		for i, w := range widths {
			if i > 0 {
				sb.WriteString("  ")
			}
			sb.WriteString(DimStyle.Render(strings.Repeat("─", w)))
		}
		sb.WriteString("\n")
	}

	// Rows
	for _, row := range t.Rows {
		for i, cell := range row {
			if i > 0 {
				sb.WriteString("  ")
			}
			if i < len(widths) {
				sb.WriteString(padRight(cell, widths[i]))
			} else {
				sb.WriteString(cell)
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
