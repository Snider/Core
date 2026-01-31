package cli

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/host-uk/core/pkg/i18n"
)

// ─────────────────────────────────────────────────────────────────────────────
// Style Namespace
// ─────────────────────────────────────────────────────────────────────────────

// Styles provides namespaced access to CLI styles.
// Usage: cli.Style.Dim.Render("text"), cli.Style.Success.Render("done")
var Style = struct {
	// Text styles
	Dim     lipgloss.Style
	Muted   lipgloss.Style
	Bold    lipgloss.Style
	Value   lipgloss.Style
	Accent  lipgloss.Style
	Code    lipgloss.Style
	Key     lipgloss.Style
	Number  lipgloss.Style
	Link    lipgloss.Style
	Header  lipgloss.Style
	Title   lipgloss.Style
	Stage   lipgloss.Style
	PrNum   lipgloss.Style
	AccentL lipgloss.Style

	// Status styles
	Success lipgloss.Style
	Error   lipgloss.Style
	Warning lipgloss.Style
	Info    lipgloss.Style

	// Git styles
	Dirty    lipgloss.Style
	Ahead    lipgloss.Style
	Behind   lipgloss.Style
	Clean    lipgloss.Style
	Conflict lipgloss.Style

	// Repo name style
	Repo lipgloss.Style

	// Coverage styles
	CoverageHigh lipgloss.Style
	CoverageMed  lipgloss.Style
	CoverageLow  lipgloss.Style

	// Priority styles
	PriorityHigh   lipgloss.Style
	PriorityMedium lipgloss.Style
	PriorityLow    lipgloss.Style

	// Severity styles
	SeverityCritical lipgloss.Style
	SeverityHigh     lipgloss.Style
	SeverityMedium   lipgloss.Style
	SeverityLow      lipgloss.Style

	// Status indicator styles
	StatusPending lipgloss.Style
	StatusRunning lipgloss.Style
	StatusSuccess lipgloss.Style
	StatusError   lipgloss.Style
	StatusWarning lipgloss.Style

	// Deploy styles
	DeploySuccess lipgloss.Style
	DeployPending lipgloss.Style
	DeployFailed  lipgloss.Style

	// Box styles
	Box        lipgloss.Style
	BoxHeader  lipgloss.Style
	ErrorBox   lipgloss.Style
	SuccessBox lipgloss.Style
}{
	// Text styles
	Dim:     DimStyle,
	Muted:   MutedStyle,
	Bold:    BoldStyle,
	Value:   ValueStyle,
	Accent:  AccentStyle,
	Code:    CodeStyle,
	Key:     KeyStyle,
	Number:  NumberStyle,
	Link:    LinkStyle,
	Header:  HeaderStyle,
	Title:   TitleStyle,
	Stage:   StageStyle,
	PrNum:   PrNumberStyle,
	AccentL: AccentLabelStyle,

	// Status styles
	Success: SuccessStyle,
	Error:   ErrorStyle,
	Warning: WarningStyle,
	Info:    InfoStyle,

	// Git styles
	Dirty:    GitDirtyStyle,
	Ahead:    GitAheadStyle,
	Behind:   GitBehindStyle,
	Clean:    GitCleanStyle,
	Conflict: GitConflictStyle,

	// Repo name style
	Repo: RepoNameStyle,

	// Coverage styles
	CoverageHigh: CoverageHighStyle,
	CoverageMed:  CoverageMedStyle,
	CoverageLow:  CoverageLowStyle,

	// Priority styles
	PriorityHigh:   PriorityHighStyle,
	PriorityMedium: PriorityMediumStyle,
	PriorityLow:    PriorityLowStyle,

	// Severity styles
	SeverityCritical: SeverityCriticalStyle,
	SeverityHigh:     SeverityHighStyle,
	SeverityMedium:   SeverityMediumStyle,
	SeverityLow:      SeverityLowStyle,

	// Status indicator styles
	StatusPending: StatusPendingStyle,
	StatusRunning: StatusRunningStyle,
	StatusSuccess: StatusSuccessStyle,
	StatusError:   StatusErrorStyle,
	StatusWarning: StatusWarningStyle,

	// Deploy styles
	DeploySuccess: DeploySuccessStyle,
	DeployPending: DeployPendingStyle,
	DeployFailed:  DeployFailedStyle,

	// Box styles
	Box:        BoxStyle,
	BoxHeader:  BoxHeaderStyle,
	ErrorBox:   ErrorBoxStyle,
	SuccessBox: SuccessBoxStyle,
}

// ─────────────────────────────────────────────────────────────────────────────
// Core Output Functions
// ─────────────────────────────────────────────────────────────────────────────

// Line translates a key via i18n.T and prints with newline.
// If no key is provided, prints an empty line.
//
//	cli.Line("i18n.progress.check")           // prints "Checking...\n"
//	cli.Line("cmd.dev.ci.short")              // prints translated text + \n
//	cli.Line("greeting", map[string]any{"Name": "World"})  // with args
//	cli.Line("")                              // prints empty line
func Line(key string, args ...any) {
	if key == "" {
		fmt.Println()
		return
	}
	fmt.Println(i18n.T(key, args...))
}

// ─────────────────────────────────────────────────────────────────────────────
// Input Functions
// ─────────────────────────────────────────────────────────────────────────────

// Scanln reads from stdin, similar to fmt.Scanln.
func Scanln(a ...any) (int, error) {
	return fmt.Scanln(a...)
}
