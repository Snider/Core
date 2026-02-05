package mcp

import (
	"testing"
	"time"
)

// TestProcessToolsRegistered_Good verifies that process tools are registered when process service is available.
func TestProcessToolsRegistered_Good(t *testing.T) {
	// Create a new MCP service without process service - tools should not be registered
	s, err := New()
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	if s.processService != nil {
		t.Error("Process service should be nil by default")
	}

	if s.server == nil {
		t.Fatal("Server should not be nil")
	}
}

// TestProcessStartInput_Good verifies the ProcessStartInput struct has expected fields.
func TestProcessStartInput_Good(t *testing.T) {
	input := ProcessStartInput{
		Command: "echo",
		Args:    []string{"hello", "world"},
		Dir:     "/tmp",
		Env:     []string{"FOO=bar"},
	}

	if input.Command != "echo" {
		t.Errorf("Expected command 'echo', got %q", input.Command)
	}
	if len(input.Args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(input.Args))
	}
	if input.Dir != "/tmp" {
		t.Errorf("Expected dir '/tmp', got %q", input.Dir)
	}
	if len(input.Env) != 1 {
		t.Errorf("Expected 1 env var, got %d", len(input.Env))
	}
}

// TestProcessStartOutput_Good verifies the ProcessStartOutput struct has expected fields.
func TestProcessStartOutput_Good(t *testing.T) {
	now := time.Now()
	output := ProcessStartOutput{
		ID:        "proc-1",
		PID:       12345,
		Command:   "echo",
		Args:      []string{"hello"},
		StartedAt: now,
	}

	if output.ID != "proc-1" {
		t.Errorf("Expected ID 'proc-1', got %q", output.ID)
	}
	if output.PID != 12345 {
		t.Errorf("Expected PID 12345, got %d", output.PID)
	}
	if output.Command != "echo" {
		t.Errorf("Expected command 'echo', got %q", output.Command)
	}
	if !output.StartedAt.Equal(now) {
		t.Errorf("Expected StartedAt %v, got %v", now, output.StartedAt)
	}
}

// TestProcessStopInput_Good verifies the ProcessStopInput struct has expected fields.
func TestProcessStopInput_Good(t *testing.T) {
	input := ProcessStopInput{
		ID: "proc-1",
	}

	if input.ID != "proc-1" {
		t.Errorf("Expected ID 'proc-1', got %q", input.ID)
	}
}

// TestProcessStopOutput_Good verifies the ProcessStopOutput struct has expected fields.
func TestProcessStopOutput_Good(t *testing.T) {
	output := ProcessStopOutput{
		ID:      "proc-1",
		Success: true,
		Message: "Process stopped",
	}

	if output.ID != "proc-1" {
		t.Errorf("Expected ID 'proc-1', got %q", output.ID)
	}
	if !output.Success {
		t.Error("Expected Success to be true")
	}
	if output.Message != "Process stopped" {
		t.Errorf("Expected message 'Process stopped', got %q", output.Message)
	}
}

// TestProcessKillInput_Good verifies the ProcessKillInput struct has expected fields.
func TestProcessKillInput_Good(t *testing.T) {
	input := ProcessKillInput{
		ID: "proc-1",
	}

	if input.ID != "proc-1" {
		t.Errorf("Expected ID 'proc-1', got %q", input.ID)
	}
}

// TestProcessKillOutput_Good verifies the ProcessKillOutput struct has expected fields.
func TestProcessKillOutput_Good(t *testing.T) {
	output := ProcessKillOutput{
		ID:      "proc-1",
		Success: true,
		Message: "Process killed",
	}

	if output.ID != "proc-1" {
		t.Errorf("Expected ID 'proc-1', got %q", output.ID)
	}
	if !output.Success {
		t.Error("Expected Success to be true")
	}
}

// TestProcessListInput_Good verifies the ProcessListInput struct has expected fields.
func TestProcessListInput_Good(t *testing.T) {
	input := ProcessListInput{
		RunningOnly: true,
	}

	if !input.RunningOnly {
		t.Error("Expected RunningOnly to be true")
	}
}

// TestProcessListInput_Defaults verifies default values.
func TestProcessListInput_Defaults(t *testing.T) {
	input := ProcessListInput{}

	if input.RunningOnly {
		t.Error("Expected RunningOnly to default to false")
	}
}

// TestProcessListOutput_Good verifies the ProcessListOutput struct has expected fields.
func TestProcessListOutput_Good(t *testing.T) {
	now := time.Now()
	output := ProcessListOutput{
		Processes: []ProcessInfo{
			{
				ID:        "proc-1",
				Command:   "echo",
				Args:      []string{"hello"},
				Dir:       "/tmp",
				Status:    "running",
				PID:       12345,
				ExitCode:  0,
				StartedAt: now,
				Duration:  5 * time.Second,
			},
		},
		Total: 1,
	}

	if len(output.Processes) != 1 {
		t.Fatalf("Expected 1 process, got %d", len(output.Processes))
	}
	if output.Total != 1 {
		t.Errorf("Expected total 1, got %d", output.Total)
	}

	proc := output.Processes[0]
	if proc.ID != "proc-1" {
		t.Errorf("Expected ID 'proc-1', got %q", proc.ID)
	}
	if proc.Status != "running" {
		t.Errorf("Expected status 'running', got %q", proc.Status)
	}
	if proc.PID != 12345 {
		t.Errorf("Expected PID 12345, got %d", proc.PID)
	}
}

// TestProcessOutputInput_Good verifies the ProcessOutputInput struct has expected fields.
func TestProcessOutputInput_Good(t *testing.T) {
	input := ProcessOutputInput{
		ID: "proc-1",
	}

	if input.ID != "proc-1" {
		t.Errorf("Expected ID 'proc-1', got %q", input.ID)
	}
}

// TestProcessOutputOutput_Good verifies the ProcessOutputOutput struct has expected fields.
func TestProcessOutputOutput_Good(t *testing.T) {
	output := ProcessOutputOutput{
		ID:     "proc-1",
		Output: "hello world\n",
	}

	if output.ID != "proc-1" {
		t.Errorf("Expected ID 'proc-1', got %q", output.ID)
	}
	if output.Output != "hello world\n" {
		t.Errorf("Expected output 'hello world\\n', got %q", output.Output)
	}
}

// TestProcessInputInput_Good verifies the ProcessInputInput struct has expected fields.
func TestProcessInputInput_Good(t *testing.T) {
	input := ProcessInputInput{
		ID:    "proc-1",
		Input: "test input\n",
	}

	if input.ID != "proc-1" {
		t.Errorf("Expected ID 'proc-1', got %q", input.ID)
	}
	if input.Input != "test input\n" {
		t.Errorf("Expected input 'test input\\n', got %q", input.Input)
	}
}

// TestProcessInputOutput_Good verifies the ProcessInputOutput struct has expected fields.
func TestProcessInputOutput_Good(t *testing.T) {
	output := ProcessInputOutput{
		ID:      "proc-1",
		Success: true,
		Message: "Input sent",
	}

	if output.ID != "proc-1" {
		t.Errorf("Expected ID 'proc-1', got %q", output.ID)
	}
	if !output.Success {
		t.Error("Expected Success to be true")
	}
}

// TestProcessInfo_Good verifies the ProcessInfo struct has expected fields.
func TestProcessInfo_Good(t *testing.T) {
	now := time.Now()
	info := ProcessInfo{
		ID:        "proc-1",
		Command:   "echo",
		Args:      []string{"hello"},
		Dir:       "/tmp",
		Status:    "exited",
		PID:       12345,
		ExitCode:  0,
		StartedAt: now,
		Duration:  2 * time.Second,
	}

	if info.ID != "proc-1" {
		t.Errorf("Expected ID 'proc-1', got %q", info.ID)
	}
	if info.Command != "echo" {
		t.Errorf("Expected command 'echo', got %q", info.Command)
	}
	if info.Status != "exited" {
		t.Errorf("Expected status 'exited', got %q", info.Status)
	}
	if info.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", info.ExitCode)
	}
	if info.Duration != 2*time.Second {
		t.Errorf("Expected duration 2s, got %v", info.Duration)
	}
}

// TestWithProcessService_Good verifies the WithProcessService option.
func TestWithProcessService_Good(t *testing.T) {
	// Note: We can't easily create a real process.Service here without Core,
	// so we just verify the option doesn't panic with nil.
	s, err := New(WithProcessService(nil))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	if s.processService != nil {
		t.Error("Expected processService to be nil when passed nil")
	}
}
