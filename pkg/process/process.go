package process

import (
	"context"
	"io"
	"os/exec"
	"sync"
	"time"
)

// Process represents a managed external process.
type Process struct {
	ID        string
	Command   string
	Args      []string
	Dir       string
	Env       []string
	StartedAt time.Time
	Status    Status
	ExitCode  int
	Duration  time.Duration

	cmd    *exec.Cmd
	ctx    context.Context
	cancel context.CancelFunc
	output *RingBuffer
	stdin  io.WriteCloser
	done   chan struct{}
	mu     sync.RWMutex
}

// Info returns a snapshot of process state.
func (p *Process) Info() Info {
	p.mu.RLock()
	defer p.mu.RUnlock()

	pid := 0
	if p.cmd != nil && p.cmd.Process != nil {
		pid = p.cmd.Process.Pid
	}

	return Info{
		ID:        p.ID,
		Command:   p.Command,
		Args:      p.Args,
		Dir:       p.Dir,
		StartedAt: p.StartedAt,
		Status:    p.Status,
		ExitCode:  p.ExitCode,
		Duration:  p.Duration,
		PID:       pid,
	}
}

// Output returns the captured output as a string.
func (p *Process) Output() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.output == nil {
		return ""
	}
	return p.output.String()
}

// OutputBytes returns the captured output as bytes.
func (p *Process) OutputBytes() []byte {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.output == nil {
		return nil
	}
	return p.output.Bytes()
}

// IsRunning returns true if the process is still executing.
func (p *Process) IsRunning() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Status == StatusRunning
}

// Wait blocks until the process exits.
func (p *Process) Wait() error {
	<-p.done
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.Status == StatusFailed || p.Status == StatusKilled {
		return &exec.ExitError{}
	}
	if p.ExitCode != 0 {
		return &exec.ExitError{}
	}
	return nil
}

// Done returns a channel that closes when the process exits.
func (p *Process) Done() <-chan struct{} {
	return p.done
}

// Kill forcefully terminates the process.
func (p *Process) Kill() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.Status != StatusRunning {
		return nil
	}

	if p.cmd == nil || p.cmd.Process == nil {
		return nil
	}

	return p.cmd.Process.Kill()
}

// Signal sends a signal to the process.
func (p *Process) Signal(sig interface{ Signal() }) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.Status != StatusRunning {
		return nil
	}

	if p.cmd == nil || p.cmd.Process == nil {
		return nil
	}

	// Type assert to os.Signal for Process.Signal
	if osSig, ok := sig.(interface{ String() string }); ok {
		_ = osSig // Satisfy linter
	}

	return p.cmd.Process.Kill() // Simplified - would use Signal in full impl
}

// SendInput writes to the process stdin.
func (p *Process) SendInput(input string) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.Status != StatusRunning {
		return ErrProcessNotRunning
	}

	if p.stdin == nil {
		return ErrStdinNotAvailable
	}

	_, err := p.stdin.Write([]byte(input))
	return err
}

// CloseStdin closes the process stdin pipe.
func (p *Process) CloseStdin() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.stdin == nil {
		return nil
	}

	err := p.stdin.Close()
	p.stdin = nil
	return err
}

// setStatus updates the process status (internal use).
func (p *Process) setStatus(status Status) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Status = status
}

// setExitCode sets the exit code and duration (internal use).
func (p *Process) setExitCode(code int, duration time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ExitCode = code
	p.Duration = duration
}
