package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/host-uk/core/pkg/log"
	"github.com/host-uk/core/pkg/process"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ProcessStartInput contains parameters for starting a new process.
type ProcessStartInput struct {
	Command string   `json:"command"`        // The command to run
	Args    []string `json:"args,omitempty"` // Command arguments
	Dir     string   `json:"dir,omitempty"`  // Working directory
	Env     []string `json:"env,omitempty"`  // Environment variables (KEY=VALUE format)
}

// ProcessStartOutput contains the result of starting a process.
type ProcessStartOutput struct {
	ID        string    `json:"id"`
	PID       int       `json:"pid"`
	Command   string    `json:"command"`
	Args      []string  `json:"args"`
	StartedAt time.Time `json:"startedAt"`
}

// ProcessStopInput contains parameters for gracefully stopping a process.
type ProcessStopInput struct {
	ID string `json:"id"` // Process ID to stop
}

// ProcessStopOutput contains the result of stopping a process.
type ProcessStopOutput struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// ProcessKillInput contains parameters for force killing a process.
type ProcessKillInput struct {
	ID string `json:"id"` // Process ID to kill
}

// ProcessKillOutput contains the result of killing a process.
type ProcessKillOutput struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// ProcessListInput contains parameters for listing processes.
type ProcessListInput struct {
	RunningOnly bool `json:"running_only,omitempty"` // If true, only return running processes
}

// ProcessListOutput contains the list of processes.
type ProcessListOutput struct {
	Processes []ProcessInfo `json:"processes"`
	Total     int           `json:"total"`
}

// ProcessInfo represents information about a process.
type ProcessInfo struct {
	ID        string        `json:"id"`
	Command   string        `json:"command"`
	Args      []string      `json:"args"`
	Dir       string        `json:"dir"`
	Status    string        `json:"status"`
	PID       int           `json:"pid"`
	ExitCode  int           `json:"exitCode"`
	StartedAt time.Time     `json:"startedAt"`
	Duration  time.Duration `json:"duration"`
}

// ProcessOutputInput contains parameters for getting process output.
type ProcessOutputInput struct {
	ID string `json:"id"` // Process ID
}

// ProcessOutputOutput contains the captured output of a process.
type ProcessOutputOutput struct {
	ID     string `json:"id"`
	Output string `json:"output"`
}

// ProcessInputInput contains parameters for sending input to a process.
type ProcessInputInput struct {
	ID    string `json:"id"`    // Process ID
	Input string `json:"input"` // Input to send to stdin
}

// ProcessInputOutput contains the result of sending input to a process.
type ProcessInputOutput struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// registerProcessTools adds process management tools to the MCP server.
// Returns false if process service is not available.
func (s *Service) registerProcessTools(server *mcp.Server) bool {
	if s.processService == nil {
		return false
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "process_start",
		Description: "Start a new external process. Returns process ID for tracking.",
	}, s.processStart)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "process_stop",
		Description: "Gracefully stop a running process by ID.",
	}, s.processStop)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "process_kill",
		Description: "Force kill a process by ID. Use when process_stop doesn't work.",
	}, s.processKill)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "process_list",
		Description: "List all managed processes. Use running_only=true for only active processes.",
	}, s.processList)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "process_output",
		Description: "Get the captured output of a process by ID.",
	}, s.processOutput)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "process_input",
		Description: "Send input to a running process stdin.",
	}, s.processInput)

	return true
}

// processStart handles the process_start tool call.
func (s *Service) processStart(ctx context.Context, req *mcp.CallToolRequest, input ProcessStartInput) (*mcp.CallToolResult, ProcessStartOutput, error) {
	s.logger.Security("MCP tool execution", "tool", "process_start", "command", input.Command, "args", input.Args, "dir", input.Dir, "user", log.Username())

	if input.Command == "" {
		return nil, ProcessStartOutput{}, fmt.Errorf("command cannot be empty")
	}

	opts := process.RunOptions{
		Command: input.Command,
		Args:    input.Args,
		Dir:     input.Dir,
		Env:     input.Env,
	}

	proc, err := s.processService.StartWithOptions(ctx, opts)
	if err != nil {
		log.Error("mcp: process start failed", "command", input.Command, "err", err)
		return nil, ProcessStartOutput{}, fmt.Errorf("failed to start process: %w", err)
	}

	info := proc.Info()
	return nil, ProcessStartOutput{
		ID:        proc.ID,
		PID:       info.PID,
		Command:   proc.Command,
		Args:      proc.Args,
		StartedAt: proc.StartedAt,
	}, nil
}

// processStop handles the process_stop tool call.
func (s *Service) processStop(ctx context.Context, req *mcp.CallToolRequest, input ProcessStopInput) (*mcp.CallToolResult, ProcessStopOutput, error) {
	s.logger.Security("MCP tool execution", "tool", "process_stop", "id", input.ID, "user", log.Username())

	if input.ID == "" {
		return nil, ProcessStopOutput{}, fmt.Errorf("id cannot be empty")
	}

	proc, err := s.processService.Get(input.ID)
	if err != nil {
		log.Error("mcp: process stop failed", "id", input.ID, "err", err)
		return nil, ProcessStopOutput{}, fmt.Errorf("process not found: %w", err)
	}

	// For graceful stop, we use Kill() which sends SIGKILL
	// A more sophisticated implementation could use SIGTERM first
	if err := proc.Kill(); err != nil {
		log.Error("mcp: process stop kill failed", "id", input.ID, "err", err)
		return nil, ProcessStopOutput{}, fmt.Errorf("failed to stop process: %w", err)
	}

	return nil, ProcessStopOutput{
		ID:      input.ID,
		Success: true,
		Message: "Process stop signal sent",
	}, nil
}

// processKill handles the process_kill tool call.
func (s *Service) processKill(ctx context.Context, req *mcp.CallToolRequest, input ProcessKillInput) (*mcp.CallToolResult, ProcessKillOutput, error) {
	s.logger.Security("MCP tool execution", "tool", "process_kill", "id", input.ID, "user", log.Username())

	if input.ID == "" {
		return nil, ProcessKillOutput{}, fmt.Errorf("id cannot be empty")
	}

	if err := s.processService.Kill(input.ID); err != nil {
		log.Error("mcp: process kill failed", "id", input.ID, "err", err)
		return nil, ProcessKillOutput{}, fmt.Errorf("failed to kill process: %w", err)
	}

	return nil, ProcessKillOutput{
		ID:      input.ID,
		Success: true,
		Message: "Process killed",
	}, nil
}

// processList handles the process_list tool call.
func (s *Service) processList(ctx context.Context, req *mcp.CallToolRequest, input ProcessListInput) (*mcp.CallToolResult, ProcessListOutput, error) {
	s.logger.Info("MCP tool execution", "tool", "process_list", "running_only", input.RunningOnly, "user", log.Username())

	var procs []*process.Process
	if input.RunningOnly {
		procs = s.processService.Running()
	} else {
		procs = s.processService.List()
	}

	result := make([]ProcessInfo, len(procs))
	for i, p := range procs {
		info := p.Info()
		result[i] = ProcessInfo{
			ID:        info.ID,
			Command:   info.Command,
			Args:      info.Args,
			Dir:       info.Dir,
			Status:    string(info.Status),
			PID:       info.PID,
			ExitCode:  info.ExitCode,
			StartedAt: info.StartedAt,
			Duration:  info.Duration,
		}
	}

	return nil, ProcessListOutput{
		Processes: result,
		Total:     len(result),
	}, nil
}

// processOutput handles the process_output tool call.
func (s *Service) processOutput(ctx context.Context, req *mcp.CallToolRequest, input ProcessOutputInput) (*mcp.CallToolResult, ProcessOutputOutput, error) {
	s.logger.Info("MCP tool execution", "tool", "process_output", "id", input.ID, "user", log.Username())

	if input.ID == "" {
		return nil, ProcessOutputOutput{}, fmt.Errorf("id cannot be empty")
	}

	output, err := s.processService.Output(input.ID)
	if err != nil {
		log.Error("mcp: process output failed", "id", input.ID, "err", err)
		return nil, ProcessOutputOutput{}, fmt.Errorf("failed to get process output: %w", err)
	}

	return nil, ProcessOutputOutput{
		ID:     input.ID,
		Output: output,
	}, nil
}

// processInput handles the process_input tool call.
func (s *Service) processInput(ctx context.Context, req *mcp.CallToolRequest, input ProcessInputInput) (*mcp.CallToolResult, ProcessInputOutput, error) {
	s.logger.Security("MCP tool execution", "tool", "process_input", "id", input.ID, "user", log.Username())

	if input.ID == "" {
		return nil, ProcessInputOutput{}, fmt.Errorf("id cannot be empty")
	}
	if input.Input == "" {
		return nil, ProcessInputOutput{}, fmt.Errorf("input cannot be empty")
	}

	proc, err := s.processService.Get(input.ID)
	if err != nil {
		log.Error("mcp: process input get failed", "id", input.ID, "err", err)
		return nil, ProcessInputOutput{}, fmt.Errorf("process not found: %w", err)
	}

	if err := proc.SendInput(input.Input); err != nil {
		log.Error("mcp: process input send failed", "id", input.ID, "err", err)
		return nil, ProcessInputOutput{}, fmt.Errorf("failed to send input: %w", err)
	}

	return nil, ProcessInputOutput{
		ID:      input.ID,
		Success: true,
		Message: "Input sent successfully",
	}, nil
}
