package exec_test

import (
	"context"
	"strings"
	"testing"

	"github.com/host-uk/core/pkg/process/exec"
)

// mockLogger captures log calls for testing
type mockLogger struct {
	debugCalls []logCall
	errorCalls []logCall
}

type logCall struct {
	msg     string
	keyvals []any
}

func (m *mockLogger) Debug(msg string, keyvals ...any) {
	m.debugCalls = append(m.debugCalls, logCall{msg, keyvals})
}

func (m *mockLogger) Error(msg string, keyvals ...any) {
	m.errorCalls = append(m.errorCalls, logCall{msg, keyvals})
}

func TestCommand_Run_Good_LogsDebug(t *testing.T) {
	logger := &mockLogger{}
	ctx := context.Background()

	err := exec.Command(ctx, "echo", "hello").
		WithLogger(logger).
		Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(logger.debugCalls) != 1 {
		t.Fatalf("expected 1 debug call, got %d", len(logger.debugCalls))
	}
	if logger.debugCalls[0].msg != "executing command" {
		t.Errorf("expected msg 'executing command', got %q", logger.debugCalls[0].msg)
	}
	if len(logger.errorCalls) != 0 {
		t.Errorf("expected no error calls, got %d", len(logger.errorCalls))
	}
}

func TestCommand_Run_Bad_LogsError(t *testing.T) {
	logger := &mockLogger{}
	ctx := context.Background()

	err := exec.Command(ctx, "false").
		WithLogger(logger).
		Run()
	if err == nil {
		t.Fatal("expected error")
	}

	if len(logger.debugCalls) != 1 {
		t.Fatalf("expected 1 debug call, got %d", len(logger.debugCalls))
	}
	if len(logger.errorCalls) != 1 {
		t.Fatalf("expected 1 error call, got %d", len(logger.errorCalls))
	}
	if logger.errorCalls[0].msg != "command failed" {
		t.Errorf("expected msg 'command failed', got %q", logger.errorCalls[0].msg)
	}
}

func TestCommand_Output_Good(t *testing.T) {
	logger := &mockLogger{}
	ctx := context.Background()

	out, err := exec.Command(ctx, "echo", "test").
		WithLogger(logger).
		Output()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(string(out)) != "test" {
		t.Errorf("expected 'test', got %q", string(out))
	}
	if len(logger.debugCalls) != 1 {
		t.Errorf("expected 1 debug call, got %d", len(logger.debugCalls))
	}
}

func TestCommand_CombinedOutput_Good(t *testing.T) {
	logger := &mockLogger{}
	ctx := context.Background()

	out, err := exec.Command(ctx, "echo", "combined").
		WithLogger(logger).
		CombinedOutput()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(string(out)) != "combined" {
		t.Errorf("expected 'combined', got %q", string(out))
	}
	if len(logger.debugCalls) != 1 {
		t.Errorf("expected 1 debug call, got %d", len(logger.debugCalls))
	}
}

func TestNopLogger(t *testing.T) {
	// Verify NopLogger doesn't panic
	var nop exec.NopLogger
	nop.Debug("msg", "key", "val")
	nop.Error("msg", "key", "val")
}

func TestSetDefaultLogger(t *testing.T) {
	original := exec.DefaultLogger()
	defer exec.SetDefaultLogger(original)

	logger := &mockLogger{}
	exec.SetDefaultLogger(logger)

	if exec.DefaultLogger() != logger {
		t.Error("default logger not set correctly")
	}

	// Test nil resets to NopLogger
	exec.SetDefaultLogger(nil)
	if _, ok := exec.DefaultLogger().(exec.NopLogger); !ok {
		t.Error("expected NopLogger when setting nil")
	}
}

func TestCommand_UsesDefaultLogger(t *testing.T) {
	original := exec.DefaultLogger()
	defer exec.SetDefaultLogger(original)

	logger := &mockLogger{}
	exec.SetDefaultLogger(logger)

	ctx := context.Background()
	_ = exec.Command(ctx, "echo", "test").Run()

	if len(logger.debugCalls) != 1 {
		t.Errorf("expected default logger to receive 1 debug call, got %d", len(logger.debugCalls))
	}
}
