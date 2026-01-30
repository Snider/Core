// Package cli provides the CLI runtime and utilities.
//
// The CLI uses the Core framework for its own runtime, providing:
//   - Global singleton access via cli.App()
//   - Output service for styled terminal printing
//   - Signal handling for graceful shutdown
//   - Worker bundle spawning for commands
package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/host-uk/core/pkg/framework"
)

var (
	instance *Runtime
	once     sync.Once
)

// Runtime is the CLI's Core runtime.
type Runtime struct {
	Core   *framework.Core
	ctx    context.Context
	cancel context.CancelFunc
}

// RuntimeOptions configures the CLI runtime.
type RuntimeOptions struct {
	// AppName is the CLI application name (used in output)
	AppName string
	// Version is the CLI version string
	Version string
}

// Init initialises the global CLI runtime.
// Call this once at startup (typically in main.go).
func Init(opts RuntimeOptions) error {
	var initErr error
	once.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())

		core, err := framework.New(
			framework.WithService(NewOutputService(OutputServiceOptions{
				AppName: opts.AppName,
			})),
			framework.WithService(NewSignalService(SignalServiceOptions{
				Cancel: cancel,
			})),
			framework.WithServiceLock(),
		)
		if err != nil {
			initErr = err
			cancel()
			return
		}

		instance = &Runtime{
			Core:   core,
			ctx:    ctx,
			cancel: cancel,
		}

		// Start services
		if err := core.ServiceStartup(ctx, nil); err != nil {
			initErr = err
			return
		}
	})
	return initErr
}

// App returns the global CLI runtime.
// Panics if Init() hasn't been called.
func App() *Runtime {
	if instance == nil {
		panic("cli.App() called before cli.Init()")
	}
	return instance
}

// Context returns the CLI's root context.
// This context is cancelled on shutdown signals.
func (r *Runtime) Context() context.Context {
	return r.ctx
}

// Shutdown gracefully shuts down the CLI runtime.
func (r *Runtime) Shutdown() {
	r.cancel()
	r.Core.ServiceShutdown(r.ctx)
}

// Output returns the output service for styled printing.
func (r *Runtime) Output() *OutputService {
	return framework.MustServiceFor[*OutputService](r.Core, "output")
}

// --- Output Service ---

// OutputServiceOptions configures the output service.
type OutputServiceOptions struct {
	AppName string
}

// OutputService provides styled terminal output.
type OutputService struct {
	*framework.ServiceRuntime[OutputServiceOptions]
}

// NewOutputService creates an output service factory.
func NewOutputService(opts OutputServiceOptions) func(*framework.Core) (any, error) {
	return func(c *framework.Core) (any, error) {
		return &OutputService{
			ServiceRuntime: framework.NewServiceRuntime(c, opts),
		}, nil
	}
}

// Success prints a success message with checkmark.
func (s *OutputService) Success(msg string) {
	fmt.Println(SuccessStyle.Render(SymbolCheck + " " + msg))
}

// Error prints an error message with cross.
func (s *OutputService) Error(msg string) {
	fmt.Println(ErrorStyle.Render(SymbolCross + " " + msg))
}

// Warning prints a warning message.
func (s *OutputService) Warning(msg string) {
	fmt.Println(WarningStyle.Render(SymbolWarning + " " + msg))
}

// Info prints an info message.
func (s *OutputService) Info(msg string) {
	fmt.Println(InfoStyle.Render(SymbolInfo + " " + msg))
}

// Title prints a title/header.
func (s *OutputService) Title(msg string) {
	fmt.Println(TitleStyle.Render(msg))
}

// Dim prints dimmed/subtle text.
func (s *OutputService) Dim(msg string) {
	fmt.Println(DimStyle.Render(msg))
}

// --- Signal Service ---

// SignalServiceOptions configures the signal service.
type SignalServiceOptions struct {
	Cancel context.CancelFunc
}

// SignalService handles OS signals for graceful shutdown.
type SignalService struct {
	*framework.ServiceRuntime[SignalServiceOptions]
	sigChan chan os.Signal
}

// NewSignalService creates a signal service factory.
func NewSignalService(opts SignalServiceOptions) func(*framework.Core) (any, error) {
	return func(c *framework.Core) (any, error) {
		return &SignalService{
			ServiceRuntime: framework.NewServiceRuntime(c, opts),
			sigChan:        make(chan os.Signal, 1),
		}, nil
	}
}

// OnStartup starts listening for signals.
func (s *SignalService) OnStartup(ctx context.Context) error {
	signal.Notify(s.sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case <-s.sigChan:
			s.Opts().Cancel()
		case <-ctx.Done():
		}
	}()

	return nil
}

// OnShutdown stops listening for signals.
func (s *SignalService) OnShutdown(ctx context.Context) error {
	signal.Stop(s.sigChan)
	close(s.sigChan)
	return nil
}
