// Package cli provides common utilities and styles for CLI commands.
//
// This package contains:
//   - Terminal styling using lipgloss with Tailwind colours
//   - Unicode symbols for consistent visual indicators
//   - Helper functions for common output patterns
//   - Git and GitHub CLI utilities
package cli

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
	ColourBlue50     = lipgloss.Color("#eff6ff")
	ColourBlue100    = lipgloss.Color("#dbeafe")
	ColourBlue200    = lipgloss.Color("#bfdbfe")
	ColourBlue300    = lipgloss.Color("#93c5fd")
	ColourBlue400    = lipgloss.Color("#60a5fa")
	ColourBlue500    = lipgloss.Color("#3b82f6")
	ColourBlue600    = lipgloss.Color("#2563eb")
	ColourBlue700    = lipgloss.Color("#1d4ed8")
	ColourGreen400   = lipgloss.Color("#4ade80")
	ColourGreen500   = lipgloss.Color("#22c55e")
	ColourGreen600   = lipgloss.Color("#16a34a")
	ColourRed400     = lipgloss.Color("#f87171")
	ColourRed500     = lipgloss.Color("#ef4444")
	ColourRed600     = lipgloss.Color("#dc2626")
	ColourAmber400   = lipgloss.Color("#fbbf24")
	ColourAmber500   = lipgloss.Color("#f59e0b")
	ColourAmber600   = lipgloss.Color("#d97706")
	ColourOrange500  = lipgloss.Color("#f97316")
	ColourYellow500  = lipgloss.Color("#eab308")
	ColourEmerald500 = lipgloss.Color("#10b981")
	ColourPurple500  = lipgloss.Color("#a855f7")
	ColourViolet400  = lipgloss.Color("#a78bfa")
	ColourViolet500  = lipgloss.Color("#8b5cf6")
	ColourIndigo500  = lipgloss.Color("#6366f1")
	ColourCyan500    = lipgloss.Color("#06b6d4")
	ColourGray50     = lipgloss.Color("#f9fafb")
	ColourGray100    = lipgloss.Color("#f3f4f6")
	ColourGray200    = lipgloss.Color("#e5e7eb")
	ColourGray300    = lipgloss.Color("#d1d5db")
	ColourGray400    = lipgloss.Color("#9ca3af")
	ColourGray500    = lipgloss.Color("#6b7280")
	ColourGray600    = lipgloss.Color("#4b5563")
	ColourGray700    = lipgloss.Color("#374151")
	ColourGray800    = lipgloss.Color("#1f2937")
	ColourGray900    = lipgloss.Color("#111827")
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

	// PrNumberStyle for pull request numbers (purple, bold).
	PrNumberStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColourPurple500)

	// AccentLabelStyle for highlighted labels (violet).
	AccentLabelStyle = lipgloss.NewStyle().
				Foreground(ColourViolet400)

	// StageStyle for pipeline/QA stage headers (indigo, bold).
	StageStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColourIndigo500)
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
// Coverage Styles (for test/code coverage display)
// ─────────────────────────────────────────────────────────────────────────────

var (
	// CoverageHighStyle for good coverage (80%+).
	CoverageHighStyle = lipgloss.NewStyle().Foreground(ColourGreen500)

	// CoverageMedStyle for moderate coverage (50-79%).
	CoverageMedStyle = lipgloss.NewStyle().Foreground(ColourAmber500)

	// CoverageLowStyle for low coverage (<50%).
	CoverageLowStyle = lipgloss.NewStyle().Foreground(ColourRed500)
)

// ─────────────────────────────────────────────────────────────────────────────
// Priority Styles (for task/issue priority levels)
// ─────────────────────────────────────────────────────────────────────────────

var (
	// PriorityHighStyle for high/critical priority (red, bold).
	PriorityHighStyle = lipgloss.NewStyle().Bold(true).Foreground(ColourRed500)

	// PriorityMediumStyle for medium priority (amber).
	PriorityMediumStyle = lipgloss.NewStyle().Foreground(ColourAmber500)

	// PriorityLowStyle for low priority (green).
	PriorityLowStyle = lipgloss.NewStyle().Foreground(ColourGreen500)
)

// ─────────────────────────────────────────────────────────────────────────────
// Severity Styles (for security/QA severity levels)
// ─────────────────────────────────────────────────────────────────────────────

var (
	// SeverityCriticalStyle for critical issues (red, bold).
	SeverityCriticalStyle = lipgloss.NewStyle().Bold(true).Foreground(ColourRed500)

	// SeverityHighStyle for high severity issues (orange, bold).
	SeverityHighStyle = lipgloss.NewStyle().Bold(true).Foreground(ColourOrange500)

	// SeverityMediumStyle for medium severity issues (amber).
	SeverityMediumStyle = lipgloss.NewStyle().Foreground(ColourAmber500)

	// SeverityLowStyle for low severity issues (gray).
	SeverityLowStyle = lipgloss.NewStyle().Foreground(ColourGray500)
)

// ─────────────────────────────────────────────────────────────────────────────
// Git Status Styles (for repo state indicators)
// ─────────────────────────────────────────────────────────────────────────────

var (
	// GitDirtyStyle for uncommitted changes (red).
	GitDirtyStyle = lipgloss.NewStyle().Foreground(ColourRed500)

	// GitAheadStyle for unpushed commits (green).
	GitAheadStyle = lipgloss.NewStyle().Foreground(ColourGreen500)

	// GitBehindStyle for unpulled commits (amber).
	GitBehindStyle = lipgloss.NewStyle().Foreground(ColourAmber500)

	// GitCleanStyle for clean state (gray).
	GitCleanStyle = lipgloss.NewStyle().Foreground(ColourGray500)

	// GitConflictStyle for merge conflicts (red, bold).
	GitConflictStyle = lipgloss.NewStyle().Bold(true).Foreground(ColourRed500)
)

// ─────────────────────────────────────────────────────────────────────────────
// Deploy Styles (for deployment status display)
// ─────────────────────────────────────────────────────────────────────────────

var (
	// DeploySuccessStyle for successful deployments (emerald).
	DeploySuccessStyle = lipgloss.NewStyle().Foreground(ColourEmerald500)

	// DeployPendingStyle for pending deployments (amber).
	DeployPendingStyle = lipgloss.NewStyle().Foreground(ColourAmber500)

	// DeployFailedStyle for failed deployments (red).
	DeployFailedStyle = lipgloss.NewStyle().Foreground(ColourRed500)
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

// FmtSuccess returns a styled success message with checkmark.
func FmtSuccess(msg string) string {
	return fmt.Sprintf("%s %s", SuccessStyle.Render(SymbolCheck), msg)
}

// FmtError returns a styled error message with cross.
func FmtError(msg string) string {
	return fmt.Sprintf("%s %s", ErrorStyle.Render(SymbolCross), msg)
}

// FmtWarning returns a styled warning message with warning symbol.
func FmtWarning(msg string) string {
	return fmt.Sprintf("%s %s", WarningStyle.Render(SymbolWarning), msg)
}

// FmtInfo returns a styled info message with info symbol.
func FmtInfo(msg string) string {
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

// CheckMark returns a styled checkmark (✓) or dash (—) based on presence.
// Useful for showing presence/absence in tables and lists.
func CheckMark(present bool) string {
	if present {
		return SuccessStyle.Render(SymbolCheck)
	}
	return DimStyle.Render("—")
}

// CheckMarkCustom returns a styled indicator with custom symbols and styles.
func CheckMarkCustom(present bool, presentStyle, absentStyle lipgloss.Style, presentSymbol, absentSymbol string) string {
	if present {
		return presentStyle.Render(presentSymbol)
	}
	return absentStyle.Render(absentSymbol)
}

// Label returns a styled label for key-value display.
// Example: Label("Status") -> "Status:" in dim gray
func Label(text string) string {
	return KeyStyle.Render(text + ":")
}

// LabelValue returns a styled "label: value" pair.
// Example: LabelValue("Branch", "main") -> "Branch: main"
func LabelValue(label, value string) string {
	return fmt.Sprintf("%s %s", Label(label), value)
}

// LabelValueStyled returns a styled "label: value" pair with custom value style.
func LabelValueStyled(label, value string, valueStyle lipgloss.Style) string {
	return fmt.Sprintf("%s %s", Label(label), valueStyle.Render(value))
}

// CheckResult formats a check result with name and optional version.
// Used for environment checks like `✓ go 1.22.0` or `✗ docker`.
func CheckResult(ok bool, name string, version string) string {
	symbol := ErrorStyle.Render(SymbolCross)
	if ok {
		symbol = SuccessStyle.Render(SymbolCheck)
	}
	if version != "" {
		return fmt.Sprintf("  %s %s %s", symbol, name, DimStyle.Render(version))
	}
	return fmt.Sprintf("  %s %s", symbol, name)
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

// FmtTitle returns a styled command/section title.
func FmtTitle(text string) string {
	return TitleStyle.Render(text)
}

// FmtDim returns dimmed text.
func FmtDim(text string) string {
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

// CommandStr renders a command string in code style.
func CommandStr(cmd string) string {
	return CodeStyle.Render(cmd)
}

// Number renders a number with styling.
func Number(n int) string {
	return NumberStyle.Render(fmt.Sprintf("%d", n))
}

// FormatCoverage formats a coverage percentage with colour based on thresholds.
// High (green) >= 80%, Medium (amber) >= 50%, Low (red) < 50%.
func FormatCoverage(percent float64) string {
	var style lipgloss.Style
	switch {
	case percent >= 80:
		style = CoverageHighStyle
	case percent >= 50:
		style = CoverageMedStyle
	default:
		style = CoverageLowStyle
	}
	return style.Render(fmt.Sprintf("%.1f%%", percent))
}

// FormatCoverageCustom formats coverage with custom thresholds.
func FormatCoverageCustom(percent, highThreshold, medThreshold float64) string {
	var style lipgloss.Style
	switch {
	case percent >= highThreshold:
		style = CoverageHighStyle
	case percent >= medThreshold:
		style = CoverageMedStyle
	default:
		style = CoverageLowStyle
	}
	return style.Render(fmt.Sprintf("%.1f%%", percent))
}

// FormatSeverity returns styled text for a severity level.
func FormatSeverity(level string) string {
	switch strings.ToLower(level) {
	case "critical":
		return SeverityCriticalStyle.Render(level)
	case "high":
		return SeverityHighStyle.Render(level)
	case "medium", "med":
		return SeverityMediumStyle.Render(level)
	default:
		return SeverityLowStyle.Render(level)
	}
}

// FormatPriority returns styled text for a priority level.
func FormatPriority(level string) string {
	switch strings.ToLower(level) {
	case "high", "critical", "urgent":
		return PriorityHighStyle.Render(level)
	case "medium", "med", "normal":
		return PriorityMediumStyle.Render(level)
	default:
		return PriorityLowStyle.Render(level)
	}
}

// FormatTaskStatus returns styled text for a task status.
// Supports: pending, in_progress, completed, blocked, failed.
func FormatTaskStatus(status string) string {
	switch strings.ToLower(status) {
	case "in_progress", "in-progress", "running", "active":
		return StatusRunningStyle.Render(status)
	case "completed", "done", "finished", "success":
		return StatusSuccessStyle.Render(status)
	case "blocked", "failed", "error":
		return StatusErrorStyle.Render(status)
	default: // pending, waiting, queued
		return StatusPendingStyle.Render(status)
	}
}

// StatusPrefix returns a styled ">>" prefix for status messages.
func StatusPrefix(style lipgloss.Style) string {
	return style.Render(">>")
}

// ProgressLabel returns a dimmed label with colon for progress output.
// Example: ProgressLabel("Installing") -> "Installing:" in dim gray
func ProgressLabel(label string) string {
	return DimStyle.Render(label + ":")
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
