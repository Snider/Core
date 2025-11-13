package runtime

import (
	"context"
	"fmt"

	"github.com/Snider/Core/pkg/core"
	"github.com/Snider/Core/pkg/crypt"
	"github.com/Snider/Core/pkg/workspace"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// Runtime is the container that holds all instantiated services.
// Its fields are the concrete types, allowing Wails to bind them directly.
type Runtime struct {
	app       *application.App
	Core      *core.Core
	Crypt     *crypt.Service
	Workspace *workspace.Service
}

// ServiceFactory defines a function that creates a service instance.
type ServiceFactory func() (any, error)

// NewWithFactories creates a new Runtime instance using the provided service factories.
func NewWithFactories(app *application.App, factories map[string]ServiceFactory) (*Runtime, error) {
	services := make(map[string]any)
	coreOpts := []core.Option{
		core.WithWails(app),
	}

	for _, name := range []string{"crypt", "workspace"} {
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
		coreOpts = append(coreOpts, core.WithName(name, func(c *core.Core) (any, error) { return svcCopy, nil }))
	}

	coreInstance, err := core.New(coreOpts...)
	if err != nil {
		return nil, err
	}

	// --- Type Assertions ---
	cryptSvc, ok := services["crypt"].(*crypt.Service)
	if !ok {
		return nil, fmt.Errorf("crypt service has unexpected type")
	}
	workspaceSvc, ok := services["workspace"].(*workspace.Service)
	if !ok {
		return nil, fmt.Errorf("workspace service has unexpected type")
	}

	rt := &Runtime{
		app:       app,
		Core:      coreInstance,
		Crypt:     cryptSvc,
		Workspace: workspaceSvc,
	}

	return rt, nil
}

// New creates and wires together all application services.
func New(app *application.App) (*Runtime, error) {
	return NewWithFactories(app, map[string]ServiceFactory{
		"crypt":     func() (any, error) { return crypt.New() },
		"workspace": func() (any, error) { return workspace.New() },
	})
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
