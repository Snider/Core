package core

import (
	"context"
	"fmt"
	"sort"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// ServiceRuntime is a helper struct embedded in services to provide access to the core application.
// It is generic and can be parameterized with a service-specific options struct.
type ServiceRuntime[T any] struct {
	core *Core
	opts T
}

// NewServiceRuntime creates a new ServiceRuntime instance for a service.
// This is typically called by a service's constructor.
func NewServiceRuntime[T any](c *Core, opts T) *ServiceRuntime[T] {
	return &ServiceRuntime[T]{
		core: c,
		opts: opts,
	}
}

// Core returns the central core instance, providing access to all registered services.
func (r *ServiceRuntime[T]) Core() *Core {
	return r.core
}

// Config returns the registered Config service from the core application.
// This is a convenience method for accessing the application's configuration.
func (r *ServiceRuntime[T]) Config() Config {
	return r.core.Config()
}

// Runtime is the container that holds all instantiated services.
// Its fields are the concrete types, allowing Wails to bind them directly.
// This struct is the primary entry point for the Wails application.
type Runtime struct {
	app  *application.App
	Core *Core
}

// ServiceFactory defines a function that creates a service instance.
// This is used to decouple the service creation from the runtime initialization.
type ServiceFactory func() (any, error)

// NewWithFactories creates a new Runtime instance using the provided service factories.
// This is the most flexible way to create a new Runtime, as it allows for
// the registration of any number of services.
func NewWithFactories(app *application.App, factories map[string]ServiceFactory) (*Runtime, error) {
	services := make(map[string]any)
	coreOpts := []Option{
		WithWails(app),
	}

	names := make([]string, 0, len(factories))
	for name := range factories {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		factory := factories[name]
		svc, err := factory()
		if err != nil {
			return nil, fmt.Errorf("failed to create service %s: %w", name, err)
		}
		services[name] = svc
		svcCopy := svc
		coreOpts = append(coreOpts, WithName(name, func(c *Core) (any, error) { return svcCopy, nil }))
	}

	coreInstance, err := New(coreOpts...)
	if err != nil {
		return nil, err
	}

	// --- Type Assertions ---

	rt := &Runtime{
		app:  app,
		Core: coreInstance,
	}

	return rt, nil
}

// NewRuntime creates and wires together all application services.
// This is the simplest way to create a new Runtime, but it does not allow for
// the registration of any custom services.
func NewRuntime(app *application.App) (*Runtime, error) {
	return NewWithFactories(app, map[string]ServiceFactory{})
}

// ServiceName returns the name of the service. This is used by Wails to identify the service.
func (r *Runtime) ServiceName() string {
	return "Core"
}

// ServiceStartup is called by Wails at application startup.
// This is where the Core's startup lifecycle is initiated.
func (r *Runtime) ServiceStartup(ctx context.Context, options application.ServiceOptions) {
	r.Core.ServiceStartup(ctx, options)
}

// ServiceShutdown is called by Wails at application shutdown.
// This is where the Core's shutdown lifecycle is initiated.
func (r *Runtime) ServiceShutdown(ctx context.Context) {
	if r.Core != nil {
		r.Core.ServiceShutdown(ctx)
	}
}
