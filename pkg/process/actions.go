package process

import "time"

// --- ACTION messages (broadcast via Core.ACTION) ---

// ActionProcessStarted is broadcast when a process begins execution.
type ActionProcessStarted struct {
	ID      string
	Command string
	Args    []string
	Dir     string
	PID     int
}

// ActionProcessOutput is broadcast for each line of output.
// Subscribe to this for real-time streaming.
type ActionProcessOutput struct {
	ID     string
	Line   string
	Stream Stream
}

// ActionProcessExited is broadcast when a process completes.
// Check ExitCode for success (0) or failure.
type ActionProcessExited struct {
	ID       string
	ExitCode int
	Duration time.Duration
	Error    error // Non-nil if failed to start or was killed
}

// ActionProcessKilled is broadcast when a process is terminated.
type ActionProcessKilled struct {
	ID     string
	Signal string
}
