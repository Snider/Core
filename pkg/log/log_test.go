package log

import (
	"bytes"
	"strings"
	"testing"
)

func TestLogger_Levels(t *testing.T) {
	tests := []struct {
		name     string
		level    Level
		logFunc  func(*Logger, string, ...any)
		expected bool
	}{
		{"debug at debug", LevelDebug, (*Logger).Debug, true},
		{"info at debug", LevelDebug, (*Logger).Info, true},
		{"warn at debug", LevelDebug, (*Logger).Warn, true},
		{"error at debug", LevelDebug, (*Logger).Error, true},

		{"debug at info", LevelInfo, (*Logger).Debug, false},
		{"info at info", LevelInfo, (*Logger).Info, true},
		{"warn at info", LevelInfo, (*Logger).Warn, true},
		{"error at info", LevelInfo, (*Logger).Error, true},

		{"debug at warn", LevelWarn, (*Logger).Debug, false},
		{"info at warn", LevelWarn, (*Logger).Info, false},
		{"warn at warn", LevelWarn, (*Logger).Warn, true},
		{"error at warn", LevelWarn, (*Logger).Error, true},

		{"debug at error", LevelError, (*Logger).Debug, false},
		{"info at error", LevelError, (*Logger).Info, false},
		{"warn at error", LevelError, (*Logger).Warn, false},
		{"error at error", LevelError, (*Logger).Error, true},

		{"debug at quiet", LevelQuiet, (*Logger).Debug, false},
		{"info at quiet", LevelQuiet, (*Logger).Info, false},
		{"warn at quiet", LevelQuiet, (*Logger).Warn, false},
		{"error at quiet", LevelQuiet, (*Logger).Error, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			l := New(Options{Level: tt.level, Output: &buf})
			tt.logFunc(l, "test message")

			hasOutput := buf.Len() > 0
			if hasOutput != tt.expected {
				t.Errorf("expected output=%v, got output=%v", tt.expected, hasOutput)
			}
		})
	}
}

func TestLogger_KeyValues(t *testing.T) {
	var buf bytes.Buffer
	l := New(Options{Level: LevelDebug, Output: &buf})

	l.Info("test message", "key1", "value1", "key2", 42)

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Error("expected message in output")
	}
	if !strings.Contains(output, "key1=value1") {
		t.Error("expected key1=value1 in output")
	}
	if !strings.Contains(output, "key2=42") {
		t.Error("expected key2=42 in output")
	}
}

func TestLogger_SetLevel(t *testing.T) {
	l := New(Options{Level: LevelInfo})

	if l.Level() != LevelInfo {
		t.Error("expected initial level to be Info")
	}

	l.SetLevel(LevelDebug)
	if l.Level() != LevelDebug {
		t.Error("expected level to be Debug after SetLevel")
	}
}

func TestLevel_String(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{LevelQuiet, "quiet"},
		{LevelError, "error"},
		{LevelWarn, "warn"},
		{LevelInfo, "info"},
		{LevelDebug, "debug"},
		{Level(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestDefault(t *testing.T) {
	// Default logger should exist
	if Default() == nil {
		t.Error("expected default logger to exist")
	}

	// Package-level functions should work
	var buf bytes.Buffer
	l := New(Options{Level: LevelDebug, Output: &buf})
	SetDefault(l)

	Info("test")
	if buf.Len() == 0 {
		t.Error("expected package-level Info to produce output")
	}
}
