package process

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/host-uk/core/pkg/framework"
)

// Global default service (follows i18n pattern).
var (
	defaultService atomic.Pointer[Service]
	defaultOnce    sync.Once
	defaultErr     error
)

// Default returns the global process service.
// Returns nil if not initialized.
func Default() *Service {
	return defaultService.Load()
}

// SetDefault sets the global process service.
// Thread-safe: can be called concurrently with Default().
func SetDefault(s *Service) {
	if s == nil {
		panic("process: SetDefault called with nil service")
	}
	defaultService.Store(s)
}

// Init initializes the default global service with a Core instance.
// This is typically called during application startup.
func Init(c *framework.Core) error {
	defaultOnce.Do(func() {
		factory := NewService(Options{})
		svc, err := factory(c)
		if err != nil {
			defaultErr = err
			return
		}
		defaultService.Store(svc.(*Service))
	})
	return defaultErr
}

// --- Global convenience functions ---

// Start spawns a new process using the default service.
func Start(ctx context.Context, command string, args ...string) (*Process, error) {
	svc := Default()
	if svc == nil {
		return nil, ErrServiceNotInitialized
	}
	return svc.Start(ctx, command, args...)
}

// Run executes a command and waits for completion using the default service.
func Run(ctx context.Context, command string, args ...string) (string, error) {
	svc := Default()
	if svc == nil {
		return "", ErrServiceNotInitialized
	}
	return svc.Run(ctx, command, args...)
}

// Get returns a process by ID from the default service.
func Get(id string) (*Process, error) {
	svc := Default()
	if svc == nil {
		return nil, ErrServiceNotInitialized
	}
	return svc.Get(id)
}

// List returns all processes from the default service.
func List() []*Process {
	svc := Default()
	if svc == nil {
		return nil
	}
	return svc.List()
}

// Kill terminates a process by ID using the default service.
func Kill(id string) error {
	svc := Default()
	if svc == nil {
		return ErrServiceNotInitialized
	}
	return svc.Kill(id)
}

// ErrServiceNotInitialized is returned when the service is not initialized.
var ErrServiceNotInitialized = &ServiceError{msg: "process: service not initialized"}

// ServiceError represents a service-level error.
type ServiceError struct {
	msg string
}

func (e *ServiceError) Error() string {
	return e.msg
}
