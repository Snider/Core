// Package process provides process management for Core.
// It allows spawning, monitoring, and controlling external processes
// with output capture and streaming support.
package process

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"
)

// Process represents a managed process.
type Process struct {
	ID        string    `json:"id"`
	Command   string    `json:"command"`
	Args      []string  `json:"args"`
	Dir       string    `json:"dir"`
	StartedAt time.Time `json:"startedAt"`
	Status    Status    `json:"status"`
	ExitCode  int       `json:"exitCode"`

	cmd    *exec.Cmd
	cancel context.CancelFunc
	output *RingBuffer
	stdin  io.WriteCloser
	mu     sync.RWMutex
}

// Status represents the process status.
type Status string

const (
	StatusRunning Status = "running"
	StatusStopped Status = "stopped"
	StatusExited  Status = "exited"
	StatusFailed  Status = "failed"
)

// RingBuffer is a fixed-size buffer that overwrites old data.
type RingBuffer struct {
	data  []byte
	size  int
	start int
	end   int
	full  bool
	mu    sync.RWMutex
}

// NewRingBuffer creates a new ring buffer with the given size.
func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		data: make([]byte, size),
		size: size,
	}
}

// Write appends data to the ring buffer.
func (rb *RingBuffer) Write(p []byte) (n int, err error) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	for _, b := range p {
		rb.data[rb.end] = b
		rb.end = (rb.end + 1) % rb.size
		if rb.full {
			rb.start = (rb.start + 1) % rb.size
		}
		if rb.end == rb.start {
			rb.full = true
		}
	}
	return len(p), nil
}

// String returns the buffer contents as a string.
func (rb *RingBuffer) String() string {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if !rb.full && rb.start == rb.end {
		return ""
	}

	if rb.full {
		result := make([]byte, rb.size)
		copy(result, rb.data[rb.start:])
		copy(result[rb.size-rb.start:], rb.data[:rb.end])
		return string(result)
	}

	return string(rb.data[rb.start:rb.end])
}

// Len returns the current length of data in the buffer.
func (rb *RingBuffer) Len() int {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.full {
		return rb.size
	}
	if rb.end >= rb.start {
		return rb.end - rb.start
	}
	return rb.size - rb.start + rb.end
}

// OutputCallback is called when a process produces output.
type OutputCallback func(processID string, output string)

// StatusCallback is called when a process status changes.
type StatusCallback func(processID string, status Status, exitCode int)

// Service manages processes.
type Service struct {
	processes      map[string]*Process
	mu             sync.RWMutex
	bufSize        int
	idCounter      int
	onOutput       OutputCallback
	onStatusChange StatusCallback
}

// New creates a new process service.
func New() *Service {
	return &Service{
		processes: make(map[string]*Process),
		bufSize:   1024 * 1024, // 1MB default buffer
	}
}

// OnOutput sets a callback for process output.
func (s *Service) OnOutput(cb OutputCallback) {
	s.onOutput = cb
}

// OnStatusChange sets a callback for process status changes.
func (s *Service) OnStatusChange(cb StatusCallback) {
	s.onStatusChange = cb
}

// SetBufferSize sets the output buffer size for new processes.
func (s *Service) SetBufferSize(size int) {
	s.bufSize = size
}

// Start starts a new process.
func (s *Service) Start(command string, args []string, dir string) (*Process, error) {
	s.mu.Lock()
	s.idCounter++
	id := fmt.Sprintf("proc-%d", s.idCounter)
	s.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = dir

	// Create output buffer
	output := NewRingBuffer(s.bufSize)

	// Set up pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	proc := &Process{
		ID:        id,
		Command:   command,
		Args:      args,
		Dir:       dir,
		StartedAt: time.Now(),
		Status:    StatusRunning,
		cmd:       cmd,
		cancel:    cancel,
		output:    output,
		stdin:     stdin,
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start process: %w", err)
	}

	// Capture output in background
	go func() {
		reader := io.MultiReader(stdout, stderr)
		scanner := bufio.NewScanner(reader)
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)
		for scanner.Scan() {
			line := scanner.Text() + "\n"
			output.Write([]byte(line))
			// Call output callback if set
			if s.onOutput != nil {
				s.onOutput(id, line)
			}
		}
	}()

	// Wait for process in background
	go func() {
		err := cmd.Wait()
		proc.mu.Lock()

		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				proc.ExitCode = exitErr.ExitCode()
				proc.Status = StatusExited
			} else {
				proc.Status = StatusFailed
			}
		} else {
			proc.ExitCode = 0
			proc.Status = StatusExited
		}

		status := proc.Status
		exitCode := proc.ExitCode
		proc.mu.Unlock()

		// Call status callback if set
		if s.onStatusChange != nil {
			s.onStatusChange(id, status, exitCode)
		}
	}()

	// Store process
	s.mu.Lock()
	s.processes[id] = proc
	s.mu.Unlock()

	return proc, nil
}

// Stop stops a running process.
func (s *Service) Stop(id string) error {
	s.mu.RLock()
	proc, ok := s.processes[id]
	s.mu.RUnlock()

	if !ok {
		return fmt.Errorf("process not found: %s", id)
	}

	proc.mu.Lock()
	defer proc.mu.Unlock()

	if proc.Status != StatusRunning {
		return fmt.Errorf("process is not running: %s", proc.Status)
	}

	proc.cancel()
	proc.Status = StatusStopped
	return nil
}

// Kill forcefully kills a process.
func (s *Service) Kill(id string) error {
	s.mu.RLock()
	proc, ok := s.processes[id]
	s.mu.RUnlock()

	if !ok {
		return fmt.Errorf("process not found: %s", id)
	}

	proc.mu.Lock()
	defer proc.mu.Unlock()

	if proc.cmd.Process == nil {
		return fmt.Errorf("process has no PID")
	}

	if err := proc.cmd.Process.Kill(); err != nil {
		return fmt.Errorf("failed to kill process: %w", err)
	}

	proc.Status = StatusStopped
	return nil
}

// Get returns a process by ID.
func (s *Service) Get(id string) (*Process, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	proc, ok := s.processes[id]
	if !ok {
		return nil, fmt.Errorf("process not found: %s", id)
	}
	return proc, nil
}

// List returns all processes.
func (s *Service) List() []*Process {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Process, 0, len(s.processes))
	for _, proc := range s.processes {
		result = append(result, proc)
	}
	return result
}

// Output returns the captured output of a process.
func (s *Service) Output(id string) (string, error) {
	s.mu.RLock()
	proc, ok := s.processes[id]
	s.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("process not found: %s", id)
	}

	return proc.output.String(), nil
}

// SendInput sends input to a process's stdin.
func (s *Service) SendInput(id string, input string) error {
	s.mu.RLock()
	proc, ok := s.processes[id]
	s.mu.RUnlock()

	if !ok {
		return fmt.Errorf("process not found: %s", id)
	}

	proc.mu.RLock()
	defer proc.mu.RUnlock()

	if proc.Status != StatusRunning {
		return fmt.Errorf("process is not running")
	}

	if proc.stdin == nil {
		return fmt.Errorf("stdin not available")
	}

	_, err := proc.stdin.Write([]byte(input))
	return err
}

// Remove removes a stopped process from the list.
func (s *Service) Remove(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	proc, ok := s.processes[id]
	if !ok {
		return fmt.Errorf("process not found: %s", id)
	}

	if proc.Status == StatusRunning {
		return fmt.Errorf("cannot remove running process")
	}

	delete(s.processes, id)
	return nil
}

// Info returns process info without the output.
type Info struct {
	ID        string    `json:"id"`
	Command   string    `json:"command"`
	Args      []string  `json:"args"`
	Dir       string    `json:"dir"`
	StartedAt time.Time `json:"startedAt"`
	Status    Status    `json:"status"`
	ExitCode  int       `json:"exitCode"`
	PID       int       `json:"pid"`
}

// Info returns info about a process.
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
		PID:       pid,
	}
}
