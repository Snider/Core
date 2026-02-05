package cli

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func TestSemanticOutput(t *testing.T) {
	UseASCII()

	// Test Success
	out := captureOutput(func() {
		Success("done")
	})
	if out == "" {
		t.Error("Success output empty")
	}

	// Test Error
	out = captureOutput(func() {
		Error("fail")
	})
	if out == "" {
		t.Error("Error output empty")
	}

	// Test Warn
	out = captureOutput(func() {
		Warn("warn")
	})
	if out == "" {
		t.Error("Warn output empty")
	}

	// Test Info
	out = captureOutput(func() {
		Info("info")
	})
	if out == "" {
		t.Error("Info output empty")
	}

	// Test Task
	out = captureOutput(func() {
		Task("task", "msg")
	})
	if out == "" {
		t.Error("Task output empty")
	}

	// Test Section
	out = captureOutput(func() {
		Section("section")
	})
	if out == "" {
		t.Error("Section output empty")
	}

	// Test Hint
	out = captureOutput(func() {
		Hint("hint", "msg")
	})
	if out == "" {
		t.Error("Hint output empty")
	}

	// Test Result
	out = captureOutput(func() {
		Result(true, "pass")
	})
	if out == "" {
		t.Error("Result(true) output empty")
	}

	out = captureOutput(func() {
		Result(false, "fail")
	})
	if out == "" {
		t.Error("Result(false) output empty")
	}
}
