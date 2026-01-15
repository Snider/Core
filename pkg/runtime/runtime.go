package runtime

import (
	"fmt"

	// Import the CONCRETE implementations from the internal packages.
	"github.com/Snider/Core/pkg/config"
	"github.com/Snider/Core/pkg/crypt"
	"github.com/Snider/Core/pkg/display"
	"github.com/Snider/Core/pkg/docs"
	"github.com/Snider/Core/pkg/help"
	"github.com/Snider/Core/pkg/i18n"
	"github.com/Snider/Core/pkg/ide"
	"github.com/Snider/Core/pkg/io"
	"github.com/Snider/Core/pkg/module"
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
	Docs      *docs.Service
	Help      *help.Service
	Crypt     *crypt.Service
	I18n      *i18n.Service
	IDE       *ide.Service
	Module    *module.Service
	Workspace *workspace.Service
}

// ServiceFactory defines a function that creates a service instance.
type ServiceFactory func() (any, error)

// newWithFactories creates a new Runtime instance using the provided service factories.
func newWithFactories(factories map[string]ServiceFactory) (*Runtime, error) {
	services := make(map[string]any)
	coreOpts := []core.Option{}

	for _, name := range []string{"config", "display", "docs", "help", "crypt", "i18n", "ide", "module", "workspace"} {
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
	docsSvc, ok := services["docs"].(*docs.Service)
	if !ok {
		return nil, fmt.Errorf("docs service has unexpected type")
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
	ideSvc, ok := services["ide"].(*ide.Service)
	if !ok {
		return nil, fmt.Errorf("ide service has unexpected type")
	}
	moduleSvc, ok := services["module"].(*module.Service)
	if !ok {
		return nil, fmt.Errorf("module service has unexpected type")
	}
	workspaceSvc, ok := services["workspace"].(*workspace.Service)
	if !ok {
		return nil, fmt.Errorf("workspace service has unexpected type")
	}

	// Set core reference for services that need it
	docsSvc.SetCore(coreInstance)

	// Set up ServiceRuntime for workspace (needs Config access)
	workspaceSvc.ServiceRuntime = core.NewServiceRuntime(coreInstance, workspace.Options{})

	// Set up ServiceRuntime for IDE
	ideSvc.ServiceRuntime = core.NewServiceRuntime(coreInstance, ide.Options{})

	// Set up ServiceRuntime for Module and register builtins
	moduleSvc.ServiceRuntime = core.NewServiceRuntime(coreInstance, module.Options{})
	module.RegisterBuiltins(moduleSvc.Registry())

	app := &Runtime{
		Core:      coreInstance,
		Config:    configSvc,
		Display:   displaySvc,
		Docs:      docsSvc,
		Help:      helpSvc,
		Crypt:     cryptSvc,
		I18n:      i18nSvc,
		IDE:       ideSvc,
		Module:    moduleSvc,
		Workspace: workspaceSvc,
	}

	return app, nil
}

// New creates and wires together all application services using static dependency injection.
func New() (*Runtime, error) {
	return newWithFactories(map[string]ServiceFactory{
		"config":    func() (any, error) { return config.New() },
		"display":   func() (any, error) { return display.New() },
		"docs":      func() (any, error) { return docs.New(docs.Options{BaseURL: "https://docs.lethean.io"}) },
		"help":      func() (any, error) { return help.New(help.Options{}) },
		"crypt":     func() (any, error) { return crypt.New() },
		"i18n":      func() (any, error) { return i18n.New() },
		"ide":       func() (any, error) { return ide.New() },
		"module":    func() (any, error) { return module.NewService(module.Options{AppsDir: "apps"}) },
		"workspace": func() (any, error) { return workspace.New(io.Local) },
	})
}
