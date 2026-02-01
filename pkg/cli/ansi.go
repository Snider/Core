package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

// ANSI escape codes
const (
	ansiReset     = "\033[0m"
	ansiBold      = "\033[1m"
	ansiDim       = "\033[2m"
	ansiItalic    = "\033[3m"
	ansiUnderline = "\033[4m"
)

var (
	colorEnabled   = true
	colorEnabledMu sync.RWMutex
)

func init() {
	// NO_COLOR standard: https://no-color.org/
	// If NO_COLOR is set (to any value, including empty), disable colors.
	if _, exists := os.LookupEnv("NO_COLOR"); exists {
		colorEnabled = false
		return
	}

	// TERM=dumb indicates a terminal without color support.
	if os.Getenv("TERM") == "dumb" {
		colorEnabled = false
	}
}

// ColorEnabled returns true if ANSI color output is enabled.
func ColorEnabled() bool {
	colorEnabledMu.RLock()
	defer colorEnabledMu.RUnlock()
	return colorEnabled
}

// SetColorEnabled enables or disables ANSI color output.
// This overrides the NO_COLOR environment variable check.
func SetColorEnabled(enabled bool) {
	colorEnabledMu.Lock()
	colorEnabled = enabled
	colorEnabledMu.Unlock()
}

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
// Returns plain text if NO_COLOR is set or colors are disabled.
func (s *AnsiStyle) Render(text string) string {
	if s == nil || !ColorEnabled() {
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

// fgColorHex converts a hex string to an ANSI foreground color code.
func fgColorHex(hex string) string {
	r, g, b := hexToRGB(hex)
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
}

// bgColorHex converts a hex string to an ANSI background color code.
func bgColorHex(hex string) string {
	r, g, b := hexToRGB(hex)
	return fmt.Sprintf("\033[48;2;%d;%d;%dm", r, g, b)
}

// hexToRGB converts a hex string to RGB values.
func hexToRGB(hex string) (int, int, int) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 255, 255, 255
	}
	// Use 8-bit parsing since RGB values are 0-255, avoiding integer overflow on 32-bit systems.
	r, _ := strconv.ParseUint(hex[0:2], 16, 8)
	g, _ := strconv.ParseUint(hex[2:4], 16, 8)
	b, _ := strconv.ParseUint(hex[4:6], 16, 8)
	return int(r), int(g), int(b)
}
