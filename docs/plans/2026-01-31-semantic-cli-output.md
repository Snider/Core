# Semantic CLI Output Abstraction

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Zero external dependencies for CLI output. Consuming code only imports `cli` - no `fmt`, `i18n`, or `lipgloss`.

**Restore Point:** `96eaed5` - all deleted code recoverable from git history.

**Architecture:**
- Internal ANSI styling (~100 lines replaces lipgloss)
- Glyph system with themes (unicode/emoji/ascii)
- Semantic output functions (`cli.Success`, `cli.Error`, `cli.Progress`)
- HLCRF layout system for structured output (ported from RFC-001)
- Simple stdin prompts (replaces huh wizard)

**Tech Stack:** Go standard library only. Zero external dependencies for CLI output.

**Reference:** RFC-001-HLCRF-COMPOSITOR.md (lab/host.uk.com/doc/rfc/)

---

## Design Decisions

### 1. Explicit Styled Functions (NOT Prefix Detection)

The codebase uses keys like `cmd.dev.ci.short`, not `i18n.success.*`. Instead of prefix detection, use explicit functions:

```go
cli.Success("Build complete")           // ✓ Build complete (green)
cli.Error("Connection failed")          // ✗ Connection failed (red)
cli.Warn("Rate limited")                // ⚠ Rate limited (amber)
cli.Info("Connecting...")               // ℹ Connecting... (blue)

// With i18n
cli.Success(i18n.T("build.complete"))   // Caller handles translation
cli.Echo(key, args...)                  // Just translate + print, no styling
```

### 2. Delete-and-Replace Approach

No backward compatibility. Delete all lipgloss-based code, rewrite with internal ANSI:
- Delete `var Style = struct {...}` namespace (output.go)
- Delete all 50+ helper functions (styles.go)
- Delete `Symbol*` constants - replaced by glyph system
- Delete `Table` struct - rewrite with internal styling

### 3. Glyph System Replaces Symbol Constants

```go
// Before (styles.go)
const SymbolCheck = "✓"
fmt.Print(SuccessStyle.Render(SymbolCheck))

// After
cli.Success("Done")  // Internally uses Glyph(":check:")
cli.Print(":check: Done")  // Or explicit glyph
```

### 4. Simple Wizard Prompts

Replace huh forms with basic stdin:

```go
cli.Prompt("Project name", "my-project")  // text input
cli.Confirm("Continue?")                   // y/n
cli.Select("Choose", []string{"a", "b"})  // numbered list
```

---

## Phase -1: Zero-Dependency ANSI Styling

### Why

Current dependencies for ANSI escape codes:
- `lipgloss` → 15 transitive deps
- `huh` → 30 transitive deps
- Supply chain attack surface: ~45 packages

What we actually use: `style.Bold(true).Foreground(color).Render(text)`

This is ~100 lines of ANSI codes. We own it completely.

### Task -1.1: ANSI Style Package

**Files:**
- Create: `pkg/cli/ansi.go`

**Step 1: Create ansi.go with complete implementation**

```go
package cli

import (
	"fmt"
	"strconv"
	"strings"
)

// ANSI escape codes
const (
	ansiReset     = "\033[0m"
	ansiBold      = "\033[1m"
	ansiDim       = "\033[2m"
	ansiItalic    = "\033[3m"
	ansiUnderline = "\033[4m"
)

// AnsiStyle represents terminal text styling.
// Use NewStyle() to create, chain methods, call Render().
type AnsiStyle struct {
	bold      bool
	dim       bool
	italic    bool
	underline bool
	fg        string
	bg        string
}

// NewStyle creates a new empty style.
func NewStyle() *AnsiStyle {
	return &AnsiStyle{}
}

// Bold enables bold text.
func (s *AnsiStyle) Bold() *AnsiStyle {
	s.bold = true
	return s
}

// Dim enables dim text.
func (s *AnsiStyle) Dim() *AnsiStyle {
	s.dim = true
	return s
}

// Italic enables italic text.
func (s *AnsiStyle) Italic() *AnsiStyle {
	s.italic = true
	return s
}

// Underline enables underlined text.
func (s *AnsiStyle) Underline() *AnsiStyle {
	s.underline = true
	return s
}

// Foreground sets foreground color from hex string.
func (s *AnsiStyle) Foreground(hex string) *AnsiStyle {
	s.fg = fgColorHex(hex)
	return s
}

// Background sets background color from hex string.
func (s *AnsiStyle) Background(hex string) *AnsiStyle {
	s.bg = bgColorHex(hex)
	return s
}

// Render applies the style to text.
func (s *AnsiStyle) Render(text string) string {
	if s == nil {
		return text
	}

	var codes []string
	if s.bold {
		codes = append(codes, ansiBold)
	}
	if s.dim {
		codes = append(codes, ansiDim)
	}
	if s.italic {
		codes = append(codes, ansiItalic)
	}
	if s.underline {
		codes = append(codes, ansiUnderline)
	}
	if s.fg != "" {
		codes = append(codes, s.fg)
	}
	if s.bg != "" {
		codes = append(codes, s.bg)
	}

	if len(codes) == 0 {
		return text
	}

	return strings.Join(codes, "") + text + ansiReset
}

// Hex color support
func fgColorHex(hex string) string {
	r, g, b := hexToRGB(hex)
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
}

func bgColorHex(hex string) string {
	r, g, b := hexToRGB(hex)
	return fmt.Sprintf("\033[48;2;%d;%d;%dm", r, g, b)
}

func hexToRGB(hex string) (int, int, int) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 255, 255, 255
	}
	r, _ := strconv.ParseInt(hex[0:2], 16, 64)
	g, _ := strconv.ParseInt(hex[2:4], 16, 64)
	b, _ := strconv.ParseInt(hex[4:6], 16, 64)
	return int(r), int(g), int(b)
}
```

**Step 2: Verify build**

Run: `go build ./pkg/cli/...`
Expected: PASS

**Step 3: Commit**

```bash
git add pkg/cli/ansi.go
git commit -m "feat(cli): add zero-dependency ANSI styling

Replaces lipgloss with ~100 lines of owned code.
Supports bold, dim, italic, underline, RGB/hex colors.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task -1.2: Rewrite styles.go

**Files:**
- Rewrite: `pkg/cli/styles.go` (delete 672 lines, write ~150)

**Step 1: Delete entire file content and rewrite**

```go
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
```

**Step 2: Verify build**

Run: `go build ./pkg/cli/...`
Expected: PASS

**Step 3: Commit**

```bash
git add pkg/cli/styles.go
git commit -m "refactor(cli): rewrite styles with zero-dep ANSI

Deletes 672 lines of lipgloss code, replaces with ~150 lines.
Previous code available at 96eaed5 if needed.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task -1.3: Rewrite output.go

**Files:**
- Rewrite: `pkg/cli/output.go` (delete Style namespace, add semantic functions)

**Step 1: Delete entire file content and rewrite**

```go
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
```

**Step 2: Verify build**

Run: `go build ./pkg/cli/...`
Expected: PASS

**Step 3: Commit**

```bash
git add pkg/cli/output.go
git commit -m "refactor(cli): rewrite output with semantic functions

Replaces Style namespace with explicit Success/Error/Warn/Info.
Previous code available at 96eaed5 if needed.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task -1.4: Rewrite strings.go

**Files:**
- Rewrite: `pkg/cli/strings.go` (remove lipgloss import)

**Step 1: Delete and rewrite**

```go
package cli

import "fmt"

// Sprintf formats a string (fmt.Sprintf wrapper).
func Sprintf(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}

// Sprint formats using default formats (fmt.Sprint wrapper).
func Sprint(args ...any) string {
	return fmt.Sprint(args...)
}

// Styled returns text with a style applied.
func Styled(style *AnsiStyle, text string) string {
	return style.Render(text)
}

// Styledf returns formatted text with a style applied.
func Styledf(style *AnsiStyle, format string, args ...any) string {
	return style.Render(fmt.Sprintf(format, args...))
}

// SuccessStr returns success-styled string.
func SuccessStr(msg string) string {
	return SuccessStyle.Render(Glyph(":check:") + " " + msg)
}

// ErrorStr returns error-styled string.
func ErrorStr(msg string) string {
	return ErrorStyle.Render(Glyph(":cross:") + " " + msg)
}

// WarnStr returns warning-styled string.
func WarnStr(msg string) string {
	return WarningStyle.Render(Glyph(":warn:") + " " + msg)
}

// InfoStr returns info-styled string.
func InfoStr(msg string) string {
	return InfoStyle.Render(Glyph(":info:") + " " + msg)
}

// DimStr returns dim-styled string.
func DimStr(msg string) string {
	return DimStyle.Render(msg)
}
```

**Step 2: Verify build**

Run: `go build ./pkg/cli/...`
Expected: PASS

**Step 3: Commit**

```bash
git add pkg/cli/strings.go
git commit -m "refactor(cli): rewrite strings with zero-dep styling

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task -1.5: Update errors.go

**Files:**
- Modify: `pkg/cli/errors.go`

**Step 1: Replace SymbolCross with Glyph**

```go
// Before
fmt.Println(ErrorStyle.Render(SymbolCross + " " + msg))

// After
fmt.Println(ErrorStyle.Render(Glyph(":cross:") + " " + msg))
```

Apply to: `Fatalf`, `FatalWrap`, `FatalWrapVerb`

**Step 2: Verify build**

Run: `go build ./pkg/cli/...`
Expected: PASS

**Step 3: Commit**

```bash
git add pkg/cli/errors.go
git commit -m "refactor(cli): update errors to use glyph system

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task -1.6: Migrate pkg/php and pkg/vm

**Files:**
- Modify: `pkg/php/cmd_quality.go`
- Modify: `pkg/php/cmd_dev.go`
- Modify: `pkg/php/cmd.go`
- Modify: `pkg/vm/cmd_vm.go`

**Step 1: Replace lipgloss imports with cli**

In each file:
- Remove `"github.com/charmbracelet/lipgloss"` import
- Replace `lipgloss.NewStyle()...` with `cli.NewStyle()...`
- Replace colour references: `lipgloss.Color(...)` → hex string

**Step 2: Verify build**

Run: `go build ./pkg/php/... ./pkg/vm/...`
Expected: PASS

**Step 3: Commit**

```bash
git add pkg/php/*.go pkg/vm/*.go
git commit -m "refactor(php,vm): migrate to cli ANSI styling

Removes direct lipgloss imports.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task -1.7: Simple Wizard Prompts

**Files:**
- Create: `pkg/cli/prompt.go`
- Rewrite: `pkg/setup/cmd_wizard.go`

**Step 1: Create prompt.go**

```go
package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var stdin = bufio.NewReader(os.Stdin)

// Prompt asks for text input with a default value.
func Prompt(label, defaultVal string) (string, error) {
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", label, defaultVal)
	} else {
		fmt.Printf("%s: ", label)
	}

	input, err := stdin.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return defaultVal, nil
	}
	return input, nil
}

// Confirm asks a yes/no question.
func Confirm(label string) (bool, error) {
	fmt.Printf("%s [y/N]: ", label)

	input, err := stdin.ReadString('\n')
	if err != nil {
		return false, err
	}

	input = strings.ToLower(strings.TrimSpace(input))
	return input == "y" || input == "yes", nil
}

// Select presents numbered options and returns the selected value.
func Select(label string, options []string) (string, error) {
	fmt.Println(label)
	for i, opt := range options {
		fmt.Printf("  %d. %s\n", i+1, opt)
	}
	fmt.Printf("Choose [1-%d]: ", len(options))

	input, err := stdin.ReadString('\n')
	if err != nil {
		return "", err
	}

	n, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil || n < 1 || n > len(options) {
		return "", fmt.Errorf("invalid selection")
	}
	return options[n-1], nil
}

// MultiSelect presents checkboxes (space-separated numbers).
func MultiSelect(label string, options []string) ([]string, error) {
	fmt.Println(label)
	for i, opt := range options {
		fmt.Printf("  %d. %s\n", i+1, opt)
	}
	fmt.Printf("Choose (space-separated) [1-%d]: ", len(options))

	input, err := stdin.ReadString('\n')
	if err != nil {
		return nil, err
	}

	var selected []string
	for _, s := range strings.Fields(input) {
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 || n > len(options) {
			continue
		}
		selected = append(selected, options[n-1])
	}
	return selected, nil
}
```

**Step 2: Rewrite cmd_wizard.go to use simple prompts**

Remove huh import, replace form calls with cli.Prompt/Confirm/Select/MultiSelect.

**Step 3: Verify build**

Run: `go build ./pkg/cli/... ./pkg/setup/...`
Expected: PASS

**Step 4: Commit**

```bash
git add pkg/cli/prompt.go pkg/setup/cmd_wizard.go
git commit -m "refactor(setup): replace huh with simple stdin prompts

Removes ~30 transitive dependencies.
Previous wizard at 96eaed5 if needed.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task -1.8: Remove Charmbracelet from go.mod

**Step 1: Run go mod tidy**

```bash
go mod tidy
```

**Step 2: Verify no charmbracelet deps remain**

Run: `grep charmbracelet go.mod`
Expected: No output

**Step 3: Check binary size reduction**

```bash
go build -o /tmp/core-new ./cmd/core-cli
ls -lh /tmp/core-new
```

**Step 4: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: remove charmbracelet dependencies

Zero external dependencies for CLI output.
Binary size reduced.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

## Phase 0: HLCRF Layout System

### Task 0.1: Layout Parser

**Files:**
- Create: `pkg/cli/layout.go`

**Step 1: Create layout.go**

```go
package cli

import "fmt"

// Region represents one of the 5 HLCRF regions.
type Region rune

const (
	RegionHeader  Region = 'H'
	RegionLeft    Region = 'L'
	RegionContent Region = 'C'
	RegionRight   Region = 'R'
	RegionFooter  Region = 'F'
)

// Composite represents an HLCRF layout node.
type Composite struct {
	variant string
	path    string
	regions map[Region]*Slot
	parent  *Composite
}

// Slot holds content for a region.
type Slot struct {
	region Region
	path   string
	blocks []Renderable
	child  *Composite
}

// Renderable is anything that can be rendered to terminal.
type Renderable interface {
	Render() string
}

// StringBlock is a simple string that implements Renderable.
type StringBlock string

func (s StringBlock) Render() string { return string(s) }

// Layout creates a new layout from a variant string.
func Layout(variant string) *Composite {
	c, err := ParseVariant(variant)
	if err != nil {
		return &Composite{variant: variant, regions: make(map[Region]*Slot)}
	}
	return c
}

// ParseVariant parses a variant string like "H[LC]C[HCF]F".
func ParseVariant(variant string) (*Composite, error) {
	c := &Composite{
		variant: variant,
		path:    "",
		regions: make(map[Region]*Slot),
	}

	i := 0
	for i < len(variant) {
		r := Region(variant[i])
		if !isValidRegion(r) {
			return nil, fmt.Errorf("invalid region: %c", r)
		}

		slot := &Slot{region: r, path: string(r)}
		c.regions[r] = slot
		i++

		if i < len(variant) && variant[i] == '[' {
			end := findMatchingBracket(variant, i)
			if end == -1 {
				return nil, fmt.Errorf("unmatched bracket at %d", i)
			}
			nested, err := ParseVariant(variant[i+1 : end])
			if err != nil {
				return nil, err
			}
			nested.path = string(r) + "-"
			nested.parent = c
			slot.child = nested
			i = end + 1
		}
	}
	return c, nil
}

func isValidRegion(r Region) bool {
	return r == 'H' || r == 'L' || r == 'C' || r == 'R' || r == 'F'
}

func findMatchingBracket(s string, start int) int {
	depth := 0
	for i := start; i < len(s); i++ {
		if s[i] == '[' {
			depth++
		} else if s[i] == ']' {
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// H adds content to Header region.
func (c *Composite) H(items ...any) *Composite { c.addToRegion(RegionHeader, items...); return c }

// L adds content to Left region.
func (c *Composite) L(items ...any) *Composite { c.addToRegion(RegionLeft, items...); return c }

// C adds content to Content region.
func (c *Composite) C(items ...any) *Composite { c.addToRegion(RegionContent, items...); return c }

// R adds content to Right region.
func (c *Composite) R(items ...any) *Composite { c.addToRegion(RegionRight, items...); return c }

// F adds content to Footer region.
func (c *Composite) F(items ...any) *Composite { c.addToRegion(RegionFooter, items...); return c }

func (c *Composite) addToRegion(r Region, items ...any) {
	slot, ok := c.regions[r]
	if !ok {
		return
	}
	for _, item := range items {
		slot.blocks = append(slot.blocks, toRenderable(item))
	}
}

func toRenderable(item any) Renderable {
	switch v := item.(type) {
	case Renderable:
		return v
	case string:
		return StringBlock(v)
	default:
		return StringBlock(fmt.Sprint(v))
	}
}
```

**Step 2: Verify build**

Run: `go build ./pkg/cli/...`
Expected: PASS

**Step 3: Commit**

```bash
git add pkg/cli/layout.go
git commit -m "feat(cli): add HLCRF layout parser

Implements RFC-001 compositor pattern for terminal output.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 0.2: Terminal Renderer

**Files:**
- Create: `pkg/cli/render.go`

**Step 1: Create render.go**

```go
package cli

import (
	"fmt"
	"strings"
)

// RenderStyle controls how layouts are rendered.
type RenderStyle int

const (
	RenderFlat   RenderStyle = iota // No borders
	RenderSimple                     // --- separators
	RenderBoxed                      // Unicode box drawing
)

var currentRenderStyle = RenderFlat

func UseRenderFlat()   { currentRenderStyle = RenderFlat }
func UseRenderSimple() { currentRenderStyle = RenderSimple }
func UseRenderBoxed()  { currentRenderStyle = RenderBoxed }

// Render outputs the layout to terminal.
func (c *Composite) Render() {
	fmt.Print(c.String())
}

// String returns the rendered layout.
func (c *Composite) String() string {
	var sb strings.Builder
	c.renderTo(&sb, 0)
	return sb.String()
}

func (c *Composite) renderTo(sb *strings.Builder, depth int) {
	order := []Region{RegionHeader, RegionLeft, RegionContent, RegionRight, RegionFooter}

	var active []Region
	for _, r := range order {
		if slot, ok := c.regions[r]; ok {
			if len(slot.blocks) > 0 || slot.child != nil {
				active = append(active, r)
			}
		}
	}

	for i, r := range active {
		slot := c.regions[r]
		if i > 0 && currentRenderStyle != RenderFlat {
			c.renderSeparator(sb, depth)
		}
		c.renderSlot(sb, slot, depth)
	}
}

func (c *Composite) renderSeparator(sb *strings.Builder, depth int) {
	indent := strings.Repeat("  ", depth)
	switch currentRenderStyle {
	case RenderBoxed:
		sb.WriteString(indent + "├" + strings.Repeat("─", 40) + "┤\n")
	case RenderSimple:
		sb.WriteString(indent + strings.Repeat("─", 40) + "\n")
	}
}

func (c *Composite) renderSlot(sb *strings.Builder, slot *Slot, depth int) {
	indent := strings.Repeat("  ", depth)
	for _, block := range slot.blocks {
		for _, line := range strings.Split(block.Render(), "\n") {
			if line != "" {
				sb.WriteString(indent + line + "\n")
			}
		}
	}
	if slot.child != nil {
		slot.child.renderTo(sb, depth+1)
	}
}
```

**Step 2: Verify build**

Run: `go build ./pkg/cli/...`
Expected: PASS

**Step 3: Commit**

```bash
git add pkg/cli/render.go
git commit -m "feat(cli): add HLCRF terminal renderer

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

## Phase 1: Glyph System

### Task 1.1: Glyph Core

**Files:**
- Create: `pkg/cli/glyph.go`

**Step 1: Create glyph.go**

```go
package cli

import (
	"bytes"
	"unicode"
)

// GlyphTheme defines which symbols to use.
type GlyphTheme int

const (
	ThemeUnicode GlyphTheme = iota
	ThemeEmoji
	ThemeASCII
)

var currentTheme = ThemeUnicode

func UseUnicode() { currentTheme = ThemeUnicode }
func UseEmoji()   { currentTheme = ThemeEmoji }
func UseASCII()   { currentTheme = ThemeASCII }

func glyphMap() map[string]string {
	switch currentTheme {
	case ThemeEmoji:
		return glyphMapEmoji
	case ThemeASCII:
		return glyphMapASCII
	default:
		return glyphMapUnicode
	}
}

// Glyph converts a shortcode to its symbol.
func Glyph(code string) string {
	if sym, ok := glyphMap()[code]; ok {
		return sym
	}
	return code
}

func compileGlyphs(x string) string {
	if x == "" {
		return ""
	}
	input := bytes.NewBufferString(x)
	output := bytes.NewBufferString("")

	for {
		r, _, err := input.ReadRune()
		if err != nil {
			break
		}
		if r == ':' {
			output.WriteString(replaceGlyph(input))
		} else {
			output.WriteRune(r)
		}
	}
	return output.String()
}

func replaceGlyph(input *bytes.Buffer) string {
	code := bytes.NewBufferString(":")
	for {
		r, _, err := input.ReadRune()
		if err != nil {
			return code.String()
		}
		if r == ':' && code.Len() == 1 {
			return code.String() + replaceGlyph(input)
		}
		code.WriteRune(r)
		if unicode.IsSpace(r) {
			return code.String()
		}
		if r == ':' {
			return Glyph(code.String())
		}
	}
}
```

**Step 2: Verify build**

Run: `go build ./pkg/cli/...`
Expected: PASS

**Step 3: Commit**

```bash
git add pkg/cli/glyph.go
git commit -m "feat(cli): add glyph shortcode system

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 1.2: Glyph Maps

**Files:**
- Create: `pkg/cli/glyph_maps.go`

**Step 1: Create glyph_maps.go**

```go
package cli

var glyphMapUnicode = map[string]string{
	":check:": "✓", ":cross:": "✗", ":warn:": "⚠", ":info:": "ℹ",
	":question:": "?", ":skip:": "○", ":dot:": "●", ":circle:": "◯",
	":arrow_right:": "→", ":arrow_left:": "←", ":arrow_up:": "↑", ":arrow_down:": "↓",
	":pointer:": "▶", ":bullet:": "•", ":dash:": "─", ":pipe:": "│",
	":corner:": "└", ":tee:": "├", ":pending:": "…", ":spinner:": "⠋",
}

var glyphMapEmoji = map[string]string{
	":check:": "✅", ":cross:": "❌", ":warn:": "⚠️", ":info:": "ℹ️",
	":question:": "❓", ":skip:": "⏭️", ":dot:": "🔵", ":circle:": "⚪",
	":arrow_right:": "➡️", ":arrow_left:": "⬅️", ":arrow_up:": "⬆️", ":arrow_down:": "⬇️",
	":pointer:": "▶️", ":bullet:": "•", ":dash:": "─", ":pipe:": "│",
	":corner:": "└", ":tee:": "├", ":pending:": "⏳", ":spinner:": "🔄",
}

var glyphMapASCII = map[string]string{
	":check:": "[OK]", ":cross:": "[FAIL]", ":warn:": "[WARN]", ":info:": "[INFO]",
	":question:": "[?]", ":skip:": "[SKIP]", ":dot:": "[*]", ":circle:": "[ ]",
	":arrow_right:": "->", ":arrow_left:": "<-", ":arrow_up:": "^", ":arrow_down:": "v",
	":pointer:": ">", ":bullet:": "*", ":dash:": "-", ":pipe:": "|",
	":corner:": "`", ":tee:": "+", ":pending:": "...", ":spinner:": "-",
}
```

**Step 2: Verify build**

Run: `go build ./pkg/cli/...`
Expected: PASS

**Step 3: Commit**

```bash
git add pkg/cli/glyph_maps.go
git commit -m "feat(cli): add glyph maps for unicode/emoji/ascii

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

## Phase 2: DX-Focused Semantic Output

### Task 2.0: Semantic Patterns for Consuming Packages

**Files:**
- Create: `pkg/cli/check.go`
- Modify: `pkg/cli/output.go`

**Goal:** Eliminate display logic from consuming packages. Only `cli` knows about styling.

**Step 1: Create check.go with fluent Check builder**

```go
package cli

import "fmt"

// CheckBuilder provides fluent API for check results.
type CheckBuilder struct {
	name     string
	status   string
	style    *AnsiStyle
	icon     string
	duration string
}

// Check starts building a check result line.
//
//	cli.Check("audit").Pass()
//	cli.Check("fmt").Fail().Duration("2.3s")
//	cli.Check("test").Skip()
func Check(name string) *CheckBuilder {
	return &CheckBuilder{name: name}
}

// Pass marks the check as passed.
func (c *CheckBuilder) Pass() *CheckBuilder {
	c.status = "passed"
	c.style = SuccessStyle
	c.icon = Glyph(":check:")
	return c
}

// Fail marks the check as failed.
func (c *CheckBuilder) Fail() *CheckBuilder {
	c.status = "failed"
	c.style = ErrorStyle
	c.icon = Glyph(":cross:")
	return c
}

// Skip marks the check as skipped.
func (c *CheckBuilder) Skip() *CheckBuilder {
	c.status = "skipped"
	c.style = DimStyle
	c.icon = "-"
	return c
}

// Warn marks the check as warning.
func (c *CheckBuilder) Warn() *CheckBuilder {
	c.status = "warning"
	c.style = WarningStyle
	c.icon = Glyph(":warn:")
	return c
}

// Duration adds duration to the check result.
func (c *CheckBuilder) Duration(d string) *CheckBuilder {
	c.duration = d
	return c
}

// Message adds a custom message instead of status.
func (c *CheckBuilder) Message(msg string) *CheckBuilder {
	c.status = msg
	return c
}

// String returns the formatted check line.
func (c *CheckBuilder) String() string {
	icon := c.icon
	if c.style != nil {
		icon = c.style.Render(c.icon)
	}

	status := c.status
	if c.style != nil && c.status != "" {
		status = c.style.Render(c.status)
	}

	if c.duration != "" {
		return fmt.Sprintf("  %s %-20s %-10s %s", icon, c.name, status, DimStyle.Render(c.duration))
	}
	if status != "" {
		return fmt.Sprintf("  %s %s %s", icon, c.name, status)
	}
	return fmt.Sprintf("  %s %s", icon, c.name)
}

// Print outputs the check result.
func (c *CheckBuilder) Print() {
	fmt.Println(c.String())
}
```

**Step 2: Add semantic output functions to output.go**

```go
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
```

**Step 3: Add strings import to output.go**

```go
import (
	"fmt"
	"strings"

	"github.com/host-uk/core/pkg/i18n"
)
```

**Step 4: Verify build**

Run: `go build ./pkg/cli/...`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/cli/check.go pkg/cli/output.go
git commit -m "feat(cli): add DX-focused semantic output patterns

- Check() fluent builder for check results
- Task() for task headers
- Section() for section headers
- Hint() for labelled hints
- Severity() for severity-styled output
- Result() for pass/fail results

Consuming packages now have zero display logic.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

## Phase 3: Full Migration

### Task 3.1: Migrate All pkg/* Files

**Files:** All files in pkg/ that use:
- `i18n.T()` directly (should use `cli.Echo()`)
- `lipgloss.*` (should use `cli.*Style`)
- `fmt.Printf/Println` for output (should use `cli.Print/Println`)

**Step 1: Find all files needing migration**

```bash
grep -r "i18n\.T\|lipgloss\|fmt\.Print" pkg/ --include="*.go" | grep -v "pkg/cli/" | grep -v "_test.go"
```

**Step 2: Migrate each file**

Pattern replacements:
- `fmt.Printf(...)` → `cli.Print(...)`
- `fmt.Println(...)` → `cli.Println(...)`
- `i18n.T("key")` → `cli.Echo("key")` or keep for values
- `successStyle.Render(...)` → `cli.SuccessStyle.Render(...)`

**Step 3: Verify build**

Run: `go build ./...`
Expected: PASS

**Step 4: Commit**

```bash
git add pkg/
git commit -m "refactor: migrate all pkg/* to cli abstraction

No direct fmt/i18n/lipgloss imports outside pkg/cli.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 3.2: Tests

**Files:**
- Create: `pkg/cli/ansi_test.go`
- Create: `pkg/cli/glyph_test.go`
- Create: `pkg/cli/layout_test.go`

**Step 1: Write tests**

```go
// ansi_test.go
package cli

import "testing"

func TestAnsiStyle_Render(t *testing.T) {
	s := NewStyle().Bold().Foreground("#ff0000")
	got := s.Render("test")
	if got == "test" {
		t.Error("Expected styled output")
	}
	if !contains(got, "test") {
		t.Error("Output should contain text")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && s[len(s)-len(sub)-4:len(s)-4] == sub
}
```

**Step 2: Run tests**

Run: `go test ./pkg/cli/... -v`
Expected: PASS

**Step 3: Commit**

```bash
git add pkg/cli/*_test.go
git commit -m "test(cli): add unit tests for ANSI, glyph, layout

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 3.3: Final Verification

**Step 1: Full build**

Run: `go build ./...`
Expected: PASS

**Step 2: All tests**

Run: `go test ./...`
Expected: PASS

**Step 3: Verify zero charmbracelet**

Run: `grep charmbracelet go.mod`
Expected: No output

**Step 4: Binary test**

Run: `./bin/core dev health`
Expected: Output displays correctly

---

## Summary of New API

| Function | Purpose |
|----------|---------|
| `cli.Blank()` | Empty line |
| `cli.Echo(key, args...)` | Translate + print |
| `cli.Print(fmt, args...)` | Printf with glyphs |
| `cli.Println(fmt, args...)` | Println with glyphs |
| `cli.Success(msg)` | ✓ green |
| `cli.Error(msg)` | ✗ red |
| `cli.Warn(msg)` | ⚠ amber |
| `cli.Info(msg)` | ℹ blue |
| `cli.Dim(msg)` | Dimmed text |
| `cli.Progress(verb, n, total)` | Overwriting progress |
| `cli.ProgressDone()` | Clear progress |
| `cli.Label(word, value)` | "Label: value" |
| `cli.Prompt(label, default)` | Text input |
| `cli.Confirm(label)` | y/n |
| `cli.Select(label, opts)` | Numbered list |
| `cli.MultiSelect(label, opts)` | Multi-select |
| `cli.Glyph(code)` | Get symbol |
| `cli.UseUnicode/Emoji/ASCII()` | Set theme |
| `cli.Layout(variant)` | HLCRF layout |
| `cli.NewTable(headers...)` | Create table |
| `cli.FormatAge(time)` | "2h ago" |
| `cli.Truncate(s, max)` | Ellipsis truncation |
| `cli.Pad(s, width)` | Right-pad string |
| **DX Patterns** | |
| `cli.Task(label, msg)` | `[php] Running...` |
| `cli.Section(name)` | `── AUDIT ──` |
| `cli.Check(name).Pass/Fail/Skip()` | Fluent check result |
| `cli.Hint(label, msg)` | `install: composer...` |
| `cli.Severity(level, msg)` | Critical/high/med/low |
| `cli.Result(ok, msg)` | Pass/fail result |
