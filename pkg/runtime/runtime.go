package runtime

import (
	"fmt"

	// Import the CONCRETE implementations from the internal packages.
	"github.com/Snider/Core/pkg/config"
	"github.com/Snider/Core/pkg/crypt"
	"github.com/Snider/Core/pkg/display"
	"github.com/Snider/Core/pkg/help"
	"github.com/Snider/Core/pkg/i18n"
	"github.com/Snider/Core/pkg/workspace"
	// Import the ABSTRACT contracts (interfaces).
	"github.com/Snider/Core/pkg/core"
)

// App is the runtime container that holds all instantiated services.
// Its fields are the concrete types, allowing Wails to bind them directly.
type Runtime struct {
	Core      *core.Core
	Config    *config.Service
	Display   *display.Service
	Help      *help.Service
	Crypt     *crypt.Service
	I18n      *i18n.Service
	Workspace *workspace.Service
}

// ServiceFactory defines a function that creates a service instance.
type ServiceFactory func() (any, error)

// newWithFactories creates a new Runtime instance using the provided service factories.
func newWithFactories(factories map[string]ServiceFactory) (*Runtime, error) {
	services := make(map[string]any)
	coreOpts := []core.Option{}

	for name, factory := range factories {
		svc, err := factory()
		if err != nil {
			return nil, fmt.Errorf("failed to create service %s: %w", name, err)
		}
		services[name] = svc
		svcCopy := svc
		coreOpts = append(coreOpts, core.WithService(func(c *core.Core) (any, error) { return svcCopy, nil }))
	}

	coreInstance, err := core.New(coreOpts...)
	if err != nil {
		return nil, err
	}

	configSvc, ok := services["config"].(*config.Service)
	if !ok {
		return nil, fmt.Errorf("config service has unexpected type")
	}
	displaySvc, ok := services["display"].(*display.Service)
	if !ok {
		return nil, fmt.Errorf("display service has unexpected type")
	}
	helpSvc, ok := services["help"].(*help.Service)
	if !ok {
		return nil, fmt.Errorf("help service has unexpected type")
	}
	cryptSvc, ok := services["crypt"].(*crypt.Service)
	if !ok {
		return nil, fmt.Errorf("crypt service has unexpected type")
	}
	i18nSvc, ok := services["i18n"].(*i18n.Service)
	if !ok {
		return nil, fmt.Errorf("i18n service has unexpected type")
	}
	workspaceSvc, ok := services["workspace"].(*workspace.Service)
	if !ok {
		return nil, fmt.Errorf("workspace service has unexpected type")
	}

	app := &Runtime{
		Core:      coreInstance,
		Config:    configSvc,
		Display:   displaySvc,
		Help:      helpSvc,
		Crypt:     cryptSvc,
		I18n:      i18nSvc,
		Workspace: workspaceSvc,
	}

	return app, nil
}

// New creates and wires together all application services using static dependency injection.
func New() (*Runtime, error) {
	return newWithFactories(map[string]ServiceFactory{
		"config":    func() (any, error) { return config.New() },
		"display":   func() (any, error) { return display.New() },
		"help":      func() (any, error) { return help.New() },
		"crypt":     func() (any, error) { return crypt.New() },
		"i18n":      func() (any, error) { return i18n.New() },
		"workspace": func() (any, error) { return workspace.New() },
	})
}
