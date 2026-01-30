// Package process provides process management with Core IPC integration.
//
// The process package enables spawning, monitoring, and controlling external
// processes with output streaming via the Core ACTION system.
//
// # Getting Started
//
//	// Register with Core
//	core, _ := framework.New(
//	    framework.WithName("process", process.NewService(process.Options{})),
//	)
//
//	// Get service and run a process
//	svc := framework.MustServiceFor[*process.Service](core, "process")
//	proc, _ := svc.Start(ctx, "go", "test", "./...")
//
// # Listening for Events
//
// Process events are broadcast via Core.ACTION:
//
//	core.RegisterAction(func(c *framework.Core, msg framework.Message) error {
//	    switch m := msg.(type) {
//	    case process.ActionProcessOutput:
//	        fmt.Print(m.Line)
//	    case process.ActionProcessExited:
//	        fmt.Printf("Exit code: %d\n", m.ExitCode)
//	    }
//	    return nil
//	})
package process

import "time"

// Status represents the process lifecycle state.
type Status string

const (
	// StatusPending indicates the process is queued but not yet started.
	StatusPending Status = "pending"
	// StatusRunning indicates the process is actively executing.
	StatusRunning Status = "running"
	// StatusExited indicates the process completed (check ExitCode).
	StatusExited Status = "exited"
	// StatusFailed indicates the process could not be started.
	StatusFailed Status = "failed"
	// StatusKilled indicates the process was terminated by signal.
	StatusKilled Status = "killed"
)

// Stream identifies the output source.
type Stream string

const (
	// StreamStdout is standard output.
	StreamStdout Stream = "stdout"
	// StreamStderr is standard error.
	StreamStderr Stream = "stderr"
)

// RunOptions configures process execution.
type RunOptions struct {
	// Command is the executable to run.
	Command string
	// Args are the command arguments.
	Args []string
	// Dir is the working directory (empty = current).
	Dir string
	// Env are additional environment variables (KEY=VALUE format).
	Env []string
	// DisableCapture disables output buffering.
	// By default, output is captured to a ring buffer.
	DisableCapture bool
}

// Info provides a snapshot of process state without internal fields.
type Info struct {
	ID        string        `json:"id"`
	Command   string        `json:"command"`
	Args      []string      `json:"args"`
	Dir       string        `json:"dir"`
	StartedAt time.Time     `json:"startedAt"`
	Status    Status        `json:"status"`
	ExitCode  int           `json:"exitCode"`
	Duration  time.Duration `json:"duration"`
	PID       int           `json:"pid"`
}
