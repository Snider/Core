// Package cli provides the CLI runtime and utilities.
package cli

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/host-uk/core/pkg/io"
	"golang.org/x/term"
)

// Mode represents the CLI execution mode.
type Mode int

const (
	// ModeInteractive indicates TTY attached with coloured output.
	ModeInteractive Mode = iota
	// ModePipe indicates stdout is piped, colours disabled.
	ModePipe
	// ModeDaemon indicates headless execution, log-only output.
	ModeDaemon
)

// String returns the string representation of the Mode.
func (m Mode) String() string {
	switch m {
	case ModeInteractive:
		return "interactive"
	case ModePipe:
		return "pipe"
	case ModeDaemon:
		return "daemon"
	default:
		return "unknown"
	}
}

// DetectMode determines the execution mode based on environment.
// Checks CORE_DAEMON env var first, then TTY status.
func DetectMode() Mode {
	if os.Getenv("CORE_DAEMON") == "1" {
		return ModeDaemon
	}
	if !IsTTY() {
		return ModePipe
	}
	return ModeInteractive
}

// IsTTY returns true if stdout is a terminal.
func IsTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// IsStdinTTY returns true if stdin is a terminal.
func IsStdinTTY() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// IsStderrTTY returns true if stderr is a terminal.
func IsStderrTTY() bool {
	return term.IsTerminal(int(os.Stderr.Fd()))
}

// --- PID File Management ---

// PIDFile manages a process ID file for single-instance enforcement.
type PIDFile struct {
	medium io.Medium
	path   string
	mu     sync.Mutex
}

// NewPIDFile creates a PID file manager.
func NewPIDFile(m io.Medium, path string) *PIDFile {
	return &PIDFile{medium: m, path: path}
}

// Acquire writes the current PID to the file.
// Returns error if another instance is running.
func (p *PIDFile) Acquire() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if PID file exists
	if data, err := p.medium.Read(p.path); err == nil {
		pid, err := strconv.Atoi(data)
		if err == nil && pid > 0 {
			// Check if process is still running
			if process, err := os.FindProcess(pid); err == nil {
				if err := process.Signal(syscall.Signal(0)); err == nil {
					return fmt.Errorf("another instance is running (PID %d)", pid)
				}
			}
		}
		// Stale PID file, remove it
		_ = p.medium.Delete(p.path)
	}

	// Ensure directory exists
	if dir := filepath.Dir(p.path); dir != "." {
		if err := p.medium.EnsureDir(dir); err != nil {
			return fmt.Errorf("failed to create PID directory: %w", err)
		}
	}

	// Write current PID
	pid := os.Getpid()
	if err := p.medium.Write(p.path, strconv.Itoa(pid)); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	return nil
}

// Release removes the PID file.
func (p *PIDFile) Release() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.medium.Delete(p.path)
}

// Path returns the PID file path.
func (p *PIDFile) Path() string {
	return p.path
}

// --- Health Check Server ---

// HealthServer provides a minimal HTTP health check endpoint.
type HealthServer struct {
	addr     string
	server   *http.Server
	listener net.Listener
	mu       sync.Mutex
	ready    bool
	checks   []HealthCheck
}

// HealthCheck is a function that returns nil if healthy.
type HealthCheck func() error

// NewHealthServer creates a health check server.
func NewHealthServer(addr string) *HealthServer {
	return &HealthServer{
		addr:  addr,
		ready: true,
	}
}

// AddCheck registers a health check function.
func (h *HealthServer) AddCheck(check HealthCheck) {
	h.mu.Lock()
	h.checks = append(h.checks, check)
	h.mu.Unlock()
}

// SetReady sets the readiness status.
func (h *HealthServer) SetReady(ready bool) {
	h.mu.Lock()
	h.ready = ready
	h.mu.Unlock()
}

// Start begins serving health check endpoints.
// Endpoints:
//   - /health - liveness probe (always 200 if server is up)
//   - /ready  - readiness probe (200 if ready, 503 if not)
func (h *HealthServer) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		h.mu.Lock()
		checks := h.checks
		h.mu.Unlock()

		for _, check := range checks {
			if err := check(); err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = fmt.Fprintf(w, "unhealthy: %v\n", err)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "ok")
	})

	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		h.mu.Lock()
		ready := h.ready
		h.mu.Unlock()

		if !ready {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = fmt.Fprintln(w, "not ready")
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "ready")
	})

	listener, err := net.Listen("tcp", h.addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", h.addr, err)
	}

	h.listener = listener
	h.server = &http.Server{Handler: mux}

	go func() {
		if err := h.server.Serve(listener); err != http.ErrServerClosed {
			LogError(fmt.Sprintf("health server error: %v", err))
		}
	}()

	return nil
}

// Stop gracefully shuts down the health server.
func (h *HealthServer) Stop(ctx context.Context) error {
	if h.server == nil {
		return nil
	}
	return h.server.Shutdown(ctx)
}

// Addr returns the actual address the server is listening on.
// Useful when using port 0 for dynamic port assignment.
func (h *HealthServer) Addr() string {
	if h.listener != nil {
		return h.listener.Addr().String()
	}
	return h.addr
}

// --- Daemon Runner ---

// DaemonOptions configures daemon mode execution.
type DaemonOptions struct {
	// Medium is the storage backend for PID files.
	// Defaults to io.Local if not set.
	Medium io.Medium

	// PIDFile path for single-instance enforcement.
	// Leave empty to skip PID file management.
	PIDFile string

	// ShutdownTimeout is the maximum time to wait for graceful shutdown.
	// Default: 30 seconds.
	ShutdownTimeout time.Duration

	// HealthAddr is the address for health check endpoints.
	// Example: ":8080", "127.0.0.1:9000"
	// Leave empty to disable health checks.
	HealthAddr string

	// HealthChecks are additional health check functions.
	HealthChecks []HealthCheck

	// OnReload is called when SIGHUP is received.
	// Use for config reloading. Leave nil to ignore SIGHUP.
	OnReload func() error
}

// Daemon manages daemon lifecycle.
type Daemon struct {
	opts    DaemonOptions
	pid     *PIDFile
	health  *HealthServer
	reload  chan struct{}
	running bool
	mu      sync.Mutex
}

// NewDaemon creates a daemon runner with the given options.
func NewDaemon(opts DaemonOptions) *Daemon {
	if opts.ShutdownTimeout == 0 {
		opts.ShutdownTimeout = 30 * time.Second
	}
	if opts.Medium == nil {
		opts.Medium = io.Local
	}

	d := &Daemon{
		opts:   opts,
		reload: make(chan struct{}, 1),
	}

	if opts.PIDFile != "" {
		d.pid = NewPIDFile(opts.Medium, opts.PIDFile)
	}

	if opts.HealthAddr != "" {
		d.health = NewHealthServer(opts.HealthAddr)
		for _, check := range opts.HealthChecks {
			d.health.AddCheck(check)
		}
	}

	return d
}

// Start initialises the daemon (PID file, health server).
// Call this after cli.Init().
func (d *Daemon) Start() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.running {
		return fmt.Errorf("daemon already running")
	}

	// Acquire PID file
	if d.pid != nil {
		if err := d.pid.Acquire(); err != nil {
			return err
		}
	}

	// Start health server
	if d.health != nil {
		if err := d.health.Start(); err != nil {
			if d.pid != nil {
				_ = d.pid.Release()
			}
			return err
		}
	}

	d.running = true
	return nil
}

// Run blocks until the context is cancelled or a signal is received.
// Handles graceful shutdown with the configured timeout.
func (d *Daemon) Run(ctx context.Context) error {
	d.mu.Lock()
	if !d.running {
		d.mu.Unlock()
		return fmt.Errorf("daemon not started - call Start() first")
	}
	d.mu.Unlock()

	// Wait for context cancellation (from signal handler)
	<-ctx.Done()

	return d.Stop()
}

// Stop performs graceful shutdown.
func (d *Daemon) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.running {
		return nil
	}

	var errs []error

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), d.opts.ShutdownTimeout)
	defer cancel()

	// Stop health server
	if d.health != nil {
		d.health.SetReady(false)
		if err := d.health.Stop(shutdownCtx); err != nil {
			errs = append(errs, fmt.Errorf("health server: %w", err))
		}
	}

	// Release PID file
	if d.pid != nil {
		if err := d.pid.Release(); err != nil && !os.IsNotExist(err) {
			errs = append(errs, fmt.Errorf("pid file: %w", err))
		}
	}

	d.running = false

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}
	return nil
}

// SetReady sets the daemon readiness status for health checks.
func (d *Daemon) SetReady(ready bool) {
	if d.health != nil {
		d.health.SetReady(ready)
	}
}

// HealthAddr returns the health server address, or empty if disabled.
func (d *Daemon) HealthAddr() string {
	if d.health != nil {
		return d.health.Addr()
	}
	return ""
}

// --- Convenience Functions ---

// Run blocks until context is cancelled or signal received.
// Simple helper for daemon mode without advanced features.
//
//	cli.Init(cli.Options{AppName: "myapp"})
//	defer cli.Shutdown()
//	cli.Run(cli.Context())
func Run(ctx context.Context) error {
	mustInit()
	<-ctx.Done()
	return ctx.Err()
}

// RunWithTimeout wraps Run with a graceful shutdown timeout.
// The returned function should be deferred to replace cli.Shutdown().
//
//	cli.Init(cli.Options{AppName: "myapp"})
//	shutdown := cli.RunWithTimeout(30 * time.Second)
//	defer shutdown()
//	cli.Run(cli.Context())
func RunWithTimeout(timeout time.Duration) func() {
	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// Create done channel for shutdown completion
		done := make(chan struct{})
		go func() {
			Shutdown()
			close(done)
		}()

		select {
		case <-done:
			// Clean shutdown
		case <-ctx.Done():
			// Timeout - force exit
			LogWarn("shutdown timeout exceeded, forcing exit")
		}
	}
}
