package cli

import (
	"strings"
	"testing"
)

func TestAnsiStyle_Render(t *testing.T) {
	// Ensure colors are enabled for this test
	SetColorEnabled(true)
	defer SetColorEnabled(true) // Reset after test

	s := NewStyle().Bold().Foreground("#ff0000")
	got := s.Render("test")
	if got == "test" {
		t.Error("Expected styled output")
	}
	if !strings.Contains(got, "test") {
		t.Error("Output should contain text")
	}
	if !strings.Contains(got, "[1m") {
		t.Error("Output should contain bold code")
	}
}

func TestColorEnabled_Good(t *testing.T) {
	// Save original state
	original := ColorEnabled()
	defer SetColorEnabled(original)

	// Test enabling
	SetColorEnabled(true)
	if !ColorEnabled() {
		t.Error("ColorEnabled should return true")
	}

	// Test disabling
	SetColorEnabled(false)
	if ColorEnabled() {
		t.Error("ColorEnabled should return false")
	}
}

func TestRender_ColorDisabled_Good(t *testing.T) {
	// Save original state
	original := ColorEnabled()
	defer SetColorEnabled(original)

	// Disable colors
	SetColorEnabled(false)

	s := NewStyle().Bold().Foreground("#ff0000")
	got := s.Render("test")

	// Should return plain text without ANSI codes
	if got != "test" {
		t.Errorf("Expected plain 'test', got %q", got)
	}
}

func TestRender_ColorEnabled_Good(t *testing.T) {
	// Save original state
	original := ColorEnabled()
	defer SetColorEnabled(original)

	// Enable colors
	SetColorEnabled(true)

	s := NewStyle().Bold()
	got := s.Render("test")

	// Should contain ANSI codes
	if !strings.Contains(got, "\033[") {
		t.Error("Expected ANSI codes when colors enabled")
	}
}

func TestUseASCII_Good(t *testing.T) {
	// Save original state
	original := ColorEnabled()
	defer SetColorEnabled(original)

	// Enable first, then UseASCII should disable colors
	SetColorEnabled(true)
	UseASCII()
	if ColorEnabled() {
		t.Error("UseASCII should disable colors")
	}
}

func TestRender_NilStyle_Good(t *testing.T) {
	var s *AnsiStyle
	got := s.Render("test")
	if got != "test" {
		t.Errorf("Nil style should return plain text, got %q", got)
	}
}
