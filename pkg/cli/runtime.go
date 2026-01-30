// Package cli provides the CLI runtime and utilities.
//
// The CLI uses the Core framework for its own runtime. Usage is simple:
//
//	cli.Init(cli.Options{AppName: "core"})
//	defer cli.Shutdown()
//
//	cli.Success("Done!")
//	cli.Error("Failed")
//	if cli.Confirm("Proceed?") { ... }
//
//	// When you need the Core instance
//	c := cli.Core()
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
	instance *runtime
	once     sync.Once
)

// runtime is the CLI's internal Core runtime.
type runtime struct {
	core   *framework.Core
	ctx    context.Context
	cancel context.CancelFunc
}

// Options configures the CLI runtime.
type Options struct {
	AppName string
	Version string
}

// Init initialises the global CLI runtime.
// Call this once at startup (typically in main.go).
func Init(opts Options) error {
	var initErr error
	once.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())

		c, err := framework.New(
			framework.WithName("signal", newSignalService(cancel)),
			framework.WithServiceLock(),
		)
		if err != nil {
			initErr = err
			cancel()
			return
		}

		instance = &runtime{
			core:   c,
			ctx:    ctx,
			cancel: cancel,
		}

		if err := c.ServiceStartup(ctx, nil); err != nil {
			initErr = err
			return
		}
	})
	return initErr
}

func mustInit() {
	if instance == nil {
		panic("cli not initialised - call cli.Init() first")
	}
}

// --- Core Access ---

// Core returns the CLI's framework Core instance.
func Core() *framework.Core {
	mustInit()
	return instance.core
}

// Context returns the CLI's root context.
// Cancelled on SIGINT/SIGTERM.
func Context() context.Context {
	mustInit()
	return instance.ctx
}

// Shutdown gracefully shuts down the CLI.
func Shutdown() {
	if instance == nil {
		return
	}
	instance.cancel()
	instance.core.ServiceShutdown(instance.ctx)
}

// --- Output Functions ---

// Success prints a success message with checkmark.
func Success(msg string) {
	fmt.Println(SuccessStyle.Render(SymbolCheck + " " + msg))
}

// Error prints an error message with cross.
func Error(msg string) {
	fmt.Println(ErrorStyle.Render(SymbolCross + " " + msg))
}

// Warning prints a warning message.
func Warning(msg string) {
	fmt.Println(WarningStyle.Render(SymbolWarning + " " + msg))
}

// Info prints an info message.
func Info(msg string) {
	fmt.Println(InfoStyle.Render(SymbolInfo + " " + msg))
}

// Title prints a title/header.
func Title(msg string) {
	fmt.Println(TitleStyle.Render(msg))
}

// Dim prints dimmed/subtle text.
func Dim(msg string) {
	fmt.Println(DimStyle.Render(msg))
}

// --- Signal Service (internal) ---

type signalService struct {
	cancel  context.CancelFunc
	sigChan chan os.Signal
}

func newSignalService(cancel context.CancelFunc) func(*framework.Core) (any, error) {
	return func(c *framework.Core) (any, error) {
		return &signalService{
			cancel:  cancel,
			sigChan: make(chan os.Signal, 1),
		}, nil
	}
}

func (s *signalService) OnStartup(ctx context.Context) error {
	signal.Notify(s.sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case <-s.sigChan:
			s.cancel()
		case <-ctx.Done():
		}
	}()

	return nil
}

func (s *signalService) OnShutdown(ctx context.Context) error {
	signal.Stop(s.sigChan)
	close(s.sigChan)
	return nil
}
