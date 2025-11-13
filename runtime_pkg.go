package core

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// Runtime is the container that holds all instantiated services.
// Its fields are the concrete types, allowing Wails to bind them directly.
type Runtime struct {
	app  *application.App
	Core *Core
}

// ServiceFactory defines a function that creates a service instance.
type ServiceFactory func() (any, error)

// NewWithFactories creates a new Runtime instance using the provided service factories.
func NewWithFactories(app *application.App, factories map[string]ServiceFactory) (*Runtime, error) {
	services := make(map[string]any)
	coreOpts := []Option{
		WithWails(app),
	}

	for _, name := range []string{} {
		factory, ok := factories[name]
		if !ok {
			return nil, fmt.Errorf("service %s factory not provided", name)
		}
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
func NewRuntime(app *application.App) (*Runtime, error) {
	return NewWithFactories(app, map[string]ServiceFactory{})
}

// ServiceName returns the name of the service. This is used by Wails to identify the service.
func (r *Runtime) ServiceName() string {
	return "Core"
}

// ServiceStartup is called by Wails at application startup.
func (r *Runtime) ServiceStartup(ctx context.Context, options application.ServiceOptions) {
	r.Core.ServiceStartup(ctx, options)
}

// ServiceShutdown is called by Wails at application shutdown.
func (r *Runtime) ServiceShutdown(ctx context.Context) {
	if r.Core != nil {
		r.Core.ServiceShutdown(ctx)
	}
}
