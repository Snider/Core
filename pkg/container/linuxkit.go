package container

import (
	"bufio"
	"context"
	"fmt"
	goio "io"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/host-uk/core/pkg/io"
)

// LinuxKitManager implements the Manager interface for LinuxKit VMs.
type LinuxKitManager struct {
	state      *State
	hypervisor Hypervisor
	medium     io.Medium
}

// NewLinuxKitManager creates a new LinuxKit manager with auto-detected hypervisor.
func NewLinuxKitManager(m io.Medium) (*LinuxKitManager, error) {
	statePath, err := DefaultStatePath()
	if err != nil {
		return nil, fmt.Errorf("failed to determine state path: %w", err)
	}

	state, err := LoadState(m, statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	hypervisor, err := DetectHypervisor()
	if err != nil {
		return nil, err
	}

	return &LinuxKitManager{
		state:      state,
		hypervisor: hypervisor,
		medium:     m,
	}, nil
}

// NewLinuxKitManagerWithHypervisor creates a manager with a specific hypervisor.
func NewLinuxKitManagerWithHypervisor(m io.Medium, state *State, hypervisor Hypervisor) *LinuxKitManager {
	return &LinuxKitManager{
		state:      state,
		hypervisor: hypervisor,
		medium:     m,
	}
}

// Run starts a new LinuxKit VM from the given image.
func (m *LinuxKitManager) Run(ctx context.Context, image string, opts RunOptions) (*Container, error) {
	// Validate image exists
	if !m.medium.IsFile(image) {
		return nil, fmt.Errorf("image not found: %s", image)
	}

	// Detect image format
	format := DetectImageFormat(image)
	if format == FormatUnknown {
		return nil, fmt.Errorf("unsupported image format: %s", image)
	}

	// Generate container ID
	id, err := GenerateID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate container ID: %w", err)
	}

	// Apply defaults
	if opts.Memory <= 0 {
		opts.Memory = 1024
	}
	if opts.CPUs <= 0 {
		opts.CPUs = 1
	}
	if opts.SSHPort <= 0 {
		opts.SSHPort = 2222
	}

	// Use name or generate from ID
	name := opts.Name
	if name == "" {
		name = id[:8]
	}

	// Ensure logs directory exists
	if err := EnsureLogsDir(m.medium); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %w", err)
	}

	// Get log file path
	logPath, err := LogPath(id)
	if err != nil {
		return nil, fmt.Errorf("failed to determine log path: %w", err)
	}

	// Build hypervisor options
	hvOpts := &HypervisorOptions{
		Memory:  opts.Memory,
		CPUs:    opts.CPUs,
		LogFile: logPath,
		SSHPort: opts.SSHPort,
		Ports:   opts.Ports,
		Volumes: opts.Volumes,
		Detach:  opts.Detach,
	}

	// Build the command
	cmd, err := m.hypervisor.BuildCommand(ctx, image, hvOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to build hypervisor command: %w", err)
	}

	// Create log file
	logFile, err := os.Create(logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	// Create container record
	container := &Container{
		ID:        id,
		Name:      name,
		Image:     image,
		Status:    StatusRunning,
		StartedAt: time.Now(),
		Ports:     opts.Ports,
		Memory:    opts.Memory,
		CPUs:      opts.CPUs,
	}

	if opts.Detach {
		// Run in background
		cmd.Stdout = logFile
		cmd.Stderr = logFile

		// Start the process
		if err := cmd.Start(); err != nil {
			_ = logFile.Close()
			return nil, fmt.Errorf("failed to start VM: %w", err)
		}

		container.PID = cmd.Process.Pid

		// Save state
		if err := m.state.Add(container); err != nil {
			// Try to kill the process we just started
			_ = cmd.Process.Kill()
			_ = logFile.Close()
			return nil, fmt.Errorf("failed to save state: %w", err)
		}

		// Close log file handle (process has its own)
		_ = logFile.Close()

		// Start a goroutine to wait for process exit and update state
		go m.waitForExit(container.ID, cmd)

		return container, nil
	}

	// Run in foreground
	// Tee output to both log file and stdout
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = logFile.Close()
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		_ = logFile.Close()
		return nil, fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		_ = logFile.Close()
		return nil, fmt.Errorf("failed to start VM: %w", err)
	}

	container.PID = cmd.Process.Pid

	// Save state before waiting
	if err := m.state.Add(container); err != nil {
		_ = cmd.Process.Kill()
		_ = logFile.Close()
		return nil, fmt.Errorf("failed to save state: %w", err)
	}

	// Copy output to both log and stdout
	go func() {
		mw := goio.MultiWriter(logFile, os.Stdout)
		_, _ = goio.Copy(mw, stdout)
	}()
	go func() {
		mw := goio.MultiWriter(logFile, os.Stderr)
		_, _ = goio.Copy(mw, stderr)
	}()

	// Wait for the process to complete
	if err := cmd.Wait(); err != nil {
		container.Status = StatusError
	} else {
		container.Status = StatusStopped
	}

	_ = logFile.Close()
	if err := m.state.Update(container); err != nil {
		return container, fmt.Errorf("update container state: %w", err)
	}

	return container, nil
}

// waitForExit monitors a detached process and updates state when it exits.
func (m *LinuxKitManager) waitForExit(id string, cmd *exec.Cmd) {
	err := cmd.Wait()

	container, ok := m.state.Get(id)
	if ok {
		if err != nil {
			container.Status = StatusError
		} else {
			container.Status = StatusStopped
		}
		_ = m.state.Update(container)
	}
}

// Stop stops a running container by sending SIGTERM.
func (m *LinuxKitManager) Stop(ctx context.Context, id string) error {
	container, ok := m.state.Get(id)
	if !ok {
		return fmt.Errorf("container not found: %s", id)
	}

	if container.Status != StatusRunning {
		return fmt.Errorf("container is not running: %s", id)
	}

	// Find the process
	process, err := os.FindProcess(container.PID)
	if err != nil {
		// Process doesn't exist, update state
		container.Status = StatusStopped
		_ = m.state.Update(container)
		return nil
	}

	// Send SIGTERM
	if err := process.Signal(syscall.SIGTERM); err != nil {
		// Process might already be gone
		container.Status = StatusStopped
		_ = m.state.Update(container)
		return nil
	}

	// Honour already-cancelled contexts before waiting
	if err := ctx.Err(); err != nil {
		_ = process.Signal(syscall.SIGKILL)
		return err
	}

	// Wait for graceful shutdown with timeout
	done := make(chan struct{})
	go func() {
		_, _ = process.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Process exited gracefully
	case <-time.After(10 * time.Second):
		// Force kill
		_ = process.Signal(syscall.SIGKILL)
		<-done
	case <-ctx.Done():
		// Context cancelled
		_ = process.Signal(syscall.SIGKILL)
		return ctx.Err()
	}

	container.Status = StatusStopped
	return m.state.Update(container)
}

// List returns all known containers, verifying process state.
func (m *LinuxKitManager) List(ctx context.Context) ([]*Container, error) {
	containers := m.state.All()

	// Verify each running container's process is still alive
	for _, c := range containers {
		if c.Status == StatusRunning {
			if !isProcessRunning(c.PID) {
				c.Status = StatusStopped
				_ = m.state.Update(c)
			}
		}
	}

	return containers, nil
}

// isProcessRunning checks if a process with the given PID is still running.
func isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// On Unix, FindProcess always succeeds, so we need to send signal 0 to check
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// Logs returns a reader for the container's log output.
func (m *LinuxKitManager) Logs(ctx context.Context, id string, follow bool) (goio.ReadCloser, error) {
	_, ok := m.state.Get(id)
	if !ok {
		return nil, fmt.Errorf("container not found: %s", id)
	}

	logPath, err := LogPath(id)
	if err != nil {
		return nil, fmt.Errorf("failed to determine log path: %w", err)
	}

	if !m.medium.IsFile(logPath) {
		return nil, fmt.Errorf("no logs available for container: %s", id)
	}

	if !follow {
		// Simple case: just open and return the file
		return m.medium.Open(logPath)
	}

	// Follow mode: create a reader that tails the file
	return newFollowReader(ctx, m.medium, logPath)
}

// followReader implements goio.ReadCloser for following log files.
type followReader struct {
	file   goio.ReadCloser
	ctx    context.Context
	cancel context.CancelFunc
	reader *bufio.Reader
	medium io.Medium
	path   string
}

func newFollowReader(ctx context.Context, m io.Medium, path string) (*followReader, error) {
	file, err := m.Open(path)
	if err != nil {
		return nil, err
	}

	// Note: We don't seek here because Medium.Open doesn't guarantee Seekability.

	ctx, cancel := context.WithCancel(ctx)

	return &followReader{
		file:   file,
		ctx:    ctx,
		cancel: cancel,
		reader: bufio.NewReader(file),
		medium: m,
		path:   path,
	}, nil
}

func (f *followReader) Read(p []byte) (int, error) {
	for {
		select {
		case <-f.ctx.Done():
			return 0, goio.EOF
		default:
		}

		n, err := f.reader.Read(p)
		if n > 0 {
			return n, nil
		}
		if err != nil && err != goio.EOF {
			return 0, err
		}

		// No data available, wait a bit and try again
		select {
		case <-f.ctx.Done():
			return 0, goio.EOF
		case <-time.After(100 * time.Millisecond):
			// Reset reader to pick up new data
			f.reader.Reset(f.file)
		}
	}
}

func (f *followReader) Close() error {
	f.cancel()
	return f.file.Close()
}

// Exec executes a command inside the container via SSH.
func (m *LinuxKitManager) Exec(ctx context.Context, id string, cmd []string) error {
	container, ok := m.state.Get(id)
	if !ok {
		return fmt.Errorf("container not found: %s", id)
	}

	if container.Status != StatusRunning {
		return fmt.Errorf("container is not running: %s", id)
	}

	// Default SSH port
	sshPort := 2222

	// Build SSH command
	sshArgs := []string{
		"-p", fmt.Sprintf("%d", sshPort),
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "UserKnownHostsFile=~/.core/known_hosts",
		"-o", "LogLevel=ERROR",
		"root@localhost",
	}
	sshArgs = append(sshArgs, cmd...)

	sshCmd := exec.CommandContext(ctx, "ssh", sshArgs...)
	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr

	return sshCmd.Run()
}

// State returns the manager's state (for testing).
func (m *LinuxKitManager) State() *State {
	return m.state
}

// Hypervisor returns the manager's hypervisor (for testing).
func (m *LinuxKitManager) Hypervisor() Hypervisor {
	return m.hypervisor
}
