package cli

import (
	"strings"
	"testing"
)

func TestAnsiStyle_Render(t *testing.T) {
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
