package log

import (
	"bytes"
	"context"
	"encoding/json"
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

func TestLogger_JSON(t *testing.T) {
	var buf bytes.Buffer
	l := New(Options{
		Level:  LevelDebug,
		Format: FormatJSON,
		Output: &buf,
	})

	l.Info("test message", "key1", "value1", "key2", 42)

	var data map[string]any
	if err := json.Unmarshal(buf.Bytes(), &data); err != nil {
		t.Fatalf("failed to unmarshal JSON output: %v", err)
	}

	if data["msg"] != "test message" {
		t.Errorf("expected msg to be %q, got %q", "test message", data["msg"])
	}
	if data["key1"] != "value1" {
		t.Errorf("expected key1 to be %q, got %q", "value1", data["key1"])
	}
	if data["key2"] != float64(42) {
		t.Errorf("expected key2 to be %v, got %v", 42, data["key2"])
	}
	if data["time"] == nil {
		t.Error("expected time field to exist")
	}
	if data["level"] != "INFO" {
		t.Errorf("expected level to be %q, got %q", "INFO", data["level"])
	}
}

func TestLogger_Context(t *testing.T) {
	var buf bytes.Buffer
	l := New(Options{Level: LevelInfo, Output: &buf})
	ctx := context.Background()

	l.InfoContext(ctx, "context info")
	if !strings.Contains(buf.String(), "context info") {
		t.Error("expected context info in output")
	}

	buf.Reset()
	l.ErrorContext(ctx, "context error")
	if !strings.Contains(buf.String(), "context error") {
		t.Error("expected context error in output")
	}
}

func TestLogger_StackTrace(t *testing.T) {
	t.Run("JSON mode has stack", func(t *testing.T) {
		var buf bytes.Buffer
		l := New(Options{
			Level:  LevelInfo,
			Format: FormatJSON,
			Output: &buf,
		})

		l.Error("test error")

		var data map[string]any
		if err := json.Unmarshal(buf.Bytes(), &data); err != nil {
			t.Fatalf("failed to unmarshal JSON output: %v", err)
		}

		if data["stack"] == nil {
			t.Error("expected stack trace in JSON error output")
		}
	})

	t.Run("Text mode has no stack", func(t *testing.T) {
		var buf bytes.Buffer
		l := New(Options{
			Level:  LevelInfo,
			Format: FormatText,
			Output: &buf,
		})

		l.Error("test error")

		if strings.Contains(buf.String(), "stack=") {
			t.Error("did not expect stack trace in text error output")
		}
	})
}

func TestLogger_Grouping(t *testing.T) {
	var buf bytes.Buffer
	l := New(Options{
		Level:  LevelDebug,
		Format: FormatText,
		Output: &buf,
	})

	l.slog.WithGroup("group1").Info("msg", "key", "val")

	output := buf.String()
	if !strings.Contains(output, "group1.key=val") {
		t.Errorf("expected group1.key=val in output, got %q", output)
	}
}
