// Package php provides Laravel/PHP development environment management.
package php

import (
	"bufio"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/host-uk/core/pkg/cli"
)

// Service represents a managed development service.
type Service interface {
	// Name returns the service name.
	Name() string
	// Start starts the service.
	Start(ctx context.Context) error
	// Stop stops the service gracefully.
	Stop() error
	// Logs returns a reader for the service logs.
	Logs(follow bool) (io.ReadCloser, error)
	// Status returns the current service status.
	Status() ServiceStatus
}

// ServiceStatus represents the status of a service.
type ServiceStatus struct {
	Name    string
	Running bool
	PID     int
	Port    int
	Error   error
}

// baseService provides common functionality for all services.
type baseService struct {
	name      string
	port      int
	dir       string
	cmd       *exec.Cmd
	logFile   *os.File
	logPath   string
	mu        sync.RWMutex
	running   bool
	lastError error
}

func (s *baseService) Name() string {
	return s.name
}

func (s *baseService) Status() ServiceStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := ServiceStatus{
		Name:    s.name,
		Running: s.running,
		Port:    s.port,
		Error:   s.lastError,
	}

	if s.cmd != nil && s.cmd.Process != nil {
		status.PID = s.cmd.Process.Pid
	}

	return status
}

func (s *baseService) Logs(follow bool) (io.ReadCloser, error) {
	if s.logPath == "" {
		return nil, cli.Err("no log file available for %s", s.name)
	}

	file, err := os.Open(s.logPath)
	if err != nil {
		return nil, cli.WrapVerb(err, "open", "log file")
	}

	if !follow {
		return file, nil
	}

	// For follow mode, return a tailing reader
	return newTailReader(file), nil
}

func (s *baseService) startProcess(ctx context.Context, cmdName string, args []string, env []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return cli.Err("%s is already running", s.name)
	}

	// Create log file
	logDir := filepath.Join(s.dir, ".core", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return cli.WrapVerb(err, "create", "log directory")
	}

	s.logPath = filepath.Join(logDir, cli.Sprintf("%s.log", strings.ToLower(s.name)))
	logFile, err := os.OpenFile(s.logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return cli.WrapVerb(err, "create", "log file")
	}
	s.logFile = logFile

	// Create command
	s.cmd = exec.CommandContext(ctx, cmdName, args...)
	s.cmd.Dir = s.dir
	s.cmd.Stdout = logFile
	s.cmd.Stderr = logFile
	s.cmd.Env = append(os.Environ(), env...)

	// Set platform-specific process attributes for clean shutdown
	setSysProcAttr(s.cmd)

	if err := s.cmd.Start(); err != nil {
		_ = logFile.Close()
		s.lastError = err
		return cli.WrapVerb(err, "start", s.name)
	}

	s.running = true
	s.lastError = nil

	// Monitor process in background
	go func() {
		err := s.cmd.Wait()
		s.mu.Lock()
		s.running = false
		if err != nil {
			s.lastError = err
		}
		if s.logFile != nil {
			_ = s.logFile.Close()
		}
		s.mu.Unlock()
	}()

	return nil
}

func (s *baseService) stopProcess() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running || s.cmd == nil || s.cmd.Process == nil {
		return nil
	}

	// Send termination signal to process (group on Unix)
	_ = signalProcessGroup(s.cmd, termSignal())

	// Wait for graceful shutdown with timeout
	done := make(chan struct{})
	go func() {
		_ = s.cmd.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Process exited gracefully
	case <-time.After(5 * time.Second):
		// Force kill
		_ = signalProcessGroup(s.cmd, killSignal())
	}

	s.running = false
	return nil
}

// FrankenPHPService manages the FrankenPHP/Octane server.
type FrankenPHPService struct {
	baseService
	https     bool
	httpsPort int
	certFile  string
	keyFile   string
}

// NewFrankenPHPService creates a new FrankenPHP service.
func NewFrankenPHPService(dir string, opts FrankenPHPOptions) *FrankenPHPService {
	port := opts.Port
	if port == 0 {
		port = 8000
	}
	httpsPort := opts.HTTPSPort
	if httpsPort == 0 {
		httpsPort = 443
	}

	return &FrankenPHPService{
		baseService: baseService{
			name: "FrankenPHP",
			port: port,
			dir:  dir,
		},
		https:     opts.HTTPS,
		httpsPort: httpsPort,
		certFile:  opts.CertFile,
		keyFile:   opts.KeyFile,
	}
}

// FrankenPHPOptions configures the FrankenPHP service.
type FrankenPHPOptions struct {
	Port      int
	HTTPSPort int
	HTTPS     bool
	CertFile  string
	KeyFile   string
}

// Start launches the FrankenPHP Octane server.
func (s *FrankenPHPService) Start(ctx context.Context) error {
	args := []string{
		"artisan", "octane:start",
		"--server=frankenphp",
		cli.Sprintf("--port=%d", s.port),
		"--no-interaction",
	}

	if s.https && s.certFile != "" && s.keyFile != "" {
		args = append(args,
			cli.Sprintf("--https-port=%d", s.httpsPort),
			cli.Sprintf("--https-certificate=%s", s.certFile),
			cli.Sprintf("--https-certificate-key=%s", s.keyFile),
		)
	}

	return s.startProcess(ctx, "php", args, nil)
}

// Stop terminates the FrankenPHP server process.
func (s *FrankenPHPService) Stop() error {
	return s.stopProcess()
}

// ViteService manages the Vite development server.
type ViteService struct {
	baseService
	packageManager string
}

// NewViteService creates a new Vite service.
func NewViteService(dir string, opts ViteOptions) *ViteService {
	port := opts.Port
	if port == 0 {
		port = 5173
	}

	pm := opts.PackageManager
	if pm == "" {
		pm = DetectPackageManager(dir)
	}

	return &ViteService{
		baseService: baseService{
			name: "Vite",
			port: port,
			dir:  dir,
		},
		packageManager: pm,
	}
}

// ViteOptions configures the Vite service.
type ViteOptions struct {
	Port           int
	PackageManager string
}

// Start launches the Vite development server.
func (s *ViteService) Start(ctx context.Context) error {
	var cmdName string
	var args []string

	switch s.packageManager {
	case "bun":
		cmdName = "bun"
		args = []string{"run", "dev"}
	case "pnpm":
		cmdName = "pnpm"
		args = []string{"run", "dev"}
	case "yarn":
		cmdName = "yarn"
		args = []string{"dev"}
	default:
		cmdName = "npm"
		args = []string{"run", "dev"}
	}

	return s.startProcess(ctx, cmdName, args, nil)
}

// Stop terminates the Vite development server.
func (s *ViteService) Stop() error {
	return s.stopProcess()
}

// HorizonService manages Laravel Horizon.
type HorizonService struct {
	baseService
}

// NewHorizonService creates a new Horizon service.
func NewHorizonService(dir string) *HorizonService {
	return &HorizonService{
		baseService: baseService{
			name: "Horizon",
			port: 0, // Horizon doesn't expose a port directly
			dir:  dir,
		},
	}
}

// Start launches the Laravel Horizon queue worker.
func (s *HorizonService) Start(ctx context.Context) error {
	return s.startProcess(ctx, "php", []string{"artisan", "horizon"}, nil)
}

// Stop terminates Horizon using its terminate command.
func (s *HorizonService) Stop() error {
	// Horizon has its own terminate command
	cmd := exec.Command("php", "artisan", "horizon:terminate")
	cmd.Dir = s.dir
	_ = cmd.Run() // Ignore errors, will also kill via signal

	return s.stopProcess()
}

// ReverbService manages Laravel Reverb WebSocket server.
type ReverbService struct {
	baseService
}

// NewReverbService creates a new Reverb service.
func NewReverbService(dir string, opts ReverbOptions) *ReverbService {
	port := opts.Port
	if port == 0 {
		port = 8080
	}

	return &ReverbService{
		baseService: baseService{
			name: "Reverb",
			port: port,
			dir:  dir,
		},
	}
}

// ReverbOptions configures the Reverb service.
type ReverbOptions struct {
	Port int
}

// Start launches the Laravel Reverb WebSocket server.
func (s *ReverbService) Start(ctx context.Context) error {
	args := []string{
		"artisan", "reverb:start",
		cli.Sprintf("--port=%d", s.port),
	}

	return s.startProcess(ctx, "php", args, nil)
}

// Stop terminates the Reverb WebSocket server.
func (s *ReverbService) Stop() error {
	return s.stopProcess()
}

// RedisService manages a local Redis server.
type RedisService struct {
	baseService
	configFile string
}

// NewRedisService creates a new Redis service.
func NewRedisService(dir string, opts RedisOptions) *RedisService {
	port := opts.Port
	if port == 0 {
		port = 6379
	}

	return &RedisService{
		baseService: baseService{
			name: "Redis",
			port: port,
			dir:  dir,
		},
		configFile: opts.ConfigFile,
	}
}

// RedisOptions configures the Redis service.
type RedisOptions struct {
	Port       int
	ConfigFile string
}

// Start launches the Redis server.
func (s *RedisService) Start(ctx context.Context) error {
	args := []string{
		"--port", cli.Sprintf("%d", s.port),
		"--daemonize", "no",
	}

	if s.configFile != "" {
		args = []string{s.configFile}
		args = append(args, "--port", cli.Sprintf("%d", s.port), "--daemonize", "no")
	}

	return s.startProcess(ctx, "redis-server", args, nil)
}

// Stop terminates Redis using the shutdown command.
func (s *RedisService) Stop() error {
	// Try graceful shutdown via redis-cli
	cmd := exec.Command("redis-cli", "-p", cli.Sprintf("%d", s.port), "shutdown", "nosave")
	_ = cmd.Run() // Ignore errors

	return s.stopProcess()
}

// tailReader wraps a file and provides tailing functionality.
type tailReader struct {
	file   *os.File
	reader *bufio.Reader
	closed bool
	mu     sync.RWMutex
}

func newTailReader(file *os.File) *tailReader {
	return &tailReader{
		file:   file,
		reader: bufio.NewReader(file),
	}
}

func (t *tailReader) Read(p []byte) (n int, err error) {
	t.mu.RLock()
	if t.closed {
		t.mu.RUnlock()
		return 0, io.EOF
	}
	t.mu.RUnlock()

	n, err = t.reader.Read(p)
	if err == io.EOF {
		// Wait a bit and try again (tailing behavior)
		time.Sleep(100 * time.Millisecond)
		return 0, nil
	}
	return n, err
}

func (t *tailReader) Close() error {
	t.mu.Lock()
	t.closed = true
	t.mu.Unlock()
	return t.file.Close()
}
