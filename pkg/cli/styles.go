// Package cli provides semantic CLI output with zero external dependencies.
package cli

import (
	"fmt"
	"strings"
	"time"
)

// Tailwind colour palette (hex strings)
const (
	ColourBlue50    = "#eff6ff"
	ColourBlue100   = "#dbeafe"
	ColourBlue200   = "#bfdbfe"
	ColourBlue300   = "#93c5fd"
	ColourBlue400   = "#60a5fa"
	ColourBlue500   = "#3b82f6"
	ColourBlue600   = "#2563eb"
	ColourBlue700   = "#1d4ed8"
	ColourGreen400  = "#4ade80"
	ColourGreen500  = "#22c55e"
	ColourGreen600  = "#16a34a"
	ColourRed400    = "#f87171"
	ColourRed500    = "#ef4444"
	ColourRed600    = "#dc2626"
	ColourAmber400  = "#fbbf24"
	ColourAmber500  = "#f59e0b"
	ColourAmber600  = "#d97706"
	ColourOrange500 = "#f97316"
	ColourYellow500 = "#eab308"
	ColourEmerald500= "#10b981"
	ColourPurple500 = "#a855f7"
	ColourViolet400 = "#a78bfa"
	ColourViolet500 = "#8b5cf6"
	ColourIndigo500 = "#6366f1"
	ColourCyan500   = "#06b6d4"
	ColourGray50    = "#f9fafb"
	ColourGray100   = "#f3f4f6"
	ColourGray200   = "#e5e7eb"
	ColourGray300   = "#d1d5db"
	ColourGray400   = "#9ca3af"
	ColourGray500   = "#6b7280"
	ColourGray600   = "#4b5563"
	ColourGray700   = "#374151"
	ColourGray800   = "#1f2937"
	ColourGray900   = "#111827"
)

// Core styles
var (
	SuccessStyle = NewStyle().Bold().Foreground(ColourGreen500)
	ErrorStyle   = NewStyle().Bold().Foreground(ColourRed500)
	WarningStyle = NewStyle().Bold().Foreground(ColourAmber500)
	InfoStyle    = NewStyle().Foreground(ColourBlue400)
	DimStyle     = NewStyle().Dim().Foreground(ColourGray500)
	MutedStyle   = NewStyle().Foreground(ColourGray600)
	BoldStyle    = NewStyle().Bold()
	KeyStyle     = NewStyle().Foreground(ColourGray400)
	ValueStyle   = NewStyle().Foreground(ColourGray200)
	AccentStyle  = NewStyle().Foreground(ColourCyan500)
	LinkStyle    = NewStyle().Foreground(ColourBlue500).Underline()
	HeaderStyle  = NewStyle().Bold().Foreground(ColourGray200)
	TitleStyle   = NewStyle().Bold().Foreground(ColourBlue500)
	CodeStyle    = NewStyle().Foreground(ColourGray300)
	NumberStyle  = NewStyle().Foreground(ColourBlue300)
	RepoStyle    = NewStyle().Bold().Foreground(ColourBlue500)
)

// Truncate shortens a string to max length with ellipsis.
func Truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

// Pad right-pads a string to width.
func Pad(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// FormatAge formats a time as human-readable age (e.g., "2h ago", "3d ago").
func FormatAge(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dw ago", int(d.Hours()/(24*7)))
	default:
		return fmt.Sprintf("%dmo ago", int(d.Hours()/(24*30)))
	}
}

// Table renders tabular data with aligned columns.
// HLCRF is for layout; Table is for tabular data - they serve different purposes.
type Table struct {
	Headers []string
	Rows    [][]string
	Style   TableStyle
}

type TableStyle struct {
	HeaderStyle *AnsiStyle
	CellStyle   *AnsiStyle
	Separator   string
}

// DefaultTableStyle returns sensible defaults.
func DefaultTableStyle() TableStyle {
	return TableStyle{
		HeaderStyle: HeaderStyle,
		CellStyle:   nil,
		Separator:   "  ",
	}
}

// NewTable creates a table with headers.
func NewTable(headers ...string) *Table {
	return &Table{
		Headers: headers,
		Style:   DefaultTableStyle(),
	}
}

// AddRow adds a row to the table.
func (t *Table) AddRow(cells ...string) *Table {
	t.Rows = append(t.Rows, cells)
	return t
}

// String renders the table.
func (t *Table) String() string {
	if len(t.Headers) == 0 && len(t.Rows) == 0 {
		return ""
	}

	// Calculate column widths
	cols := len(t.Headers)
	if cols == 0 && len(t.Rows) > 0 {
		cols = len(t.Rows[0])
	}
	widths := make([]int, cols)

	for i, h := range t.Headers {
		if len(h) > widths[i] {
			widths[i] = len(h)
		}
	}
	for _, row := range t.Rows {
		for i, cell := range row {
			if i < cols && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	var sb strings.Builder
	sep := t.Style.Separator

	// Headers
	if len(t.Headers) > 0 {
		for i, h := range t.Headers {
			if i > 0 {
				sb.WriteString(sep)
			}
			styled := Pad(h, widths[i])
			if t.Style.HeaderStyle != nil {
				styled = t.Style.HeaderStyle.Render(styled)
			}
			sb.WriteString(styled)
		}
		sb.WriteString("\n")
	}

	// Rows
	for _, row := range t.Rows {
		for i, cell := range row {
			if i > 0 {
				sb.WriteString(sep)
			}
			styled := Pad(cell, widths[i])
			if t.Style.CellStyle != nil {
				styled = t.Style.CellStyle.Render(styled)
			}
			sb.WriteString(styled)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// Render prints the table to stdout.
func (t *Table) Render() {
	fmt.Print(t.String())
}