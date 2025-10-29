package runtime

import (
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
	Core    *core.Core
	Config  *config.Service
	Display *display.Service
	Help    *help.Service
	Crypt   *crypt.Service
	I18n    *i18n.Service
	//IO        core.IO // IO is a library, not a service, so it's not injected here directly.
	Workspace *workspace.Service
}

// New creates and wires together all application services using static dependency injection.
// This is the composition root for the static initialization modality.
func New() (*Runtime, error) {
	// 1. Instantiate services that have no direct service dependencies (or only simple ones).
	configSvc, err := config.New()
	if err != nil {
		return nil, err
	}

	displaySvc, err := display.New()
	if err != nil {
		return nil, err
	}

	cryptSvc, err := crypt.New()
	if err != nil {
		return nil, err
	}

	// 2. Instantiate services that have dependencies and inject them.
	// i18n needs config
	i18nSvc, err := i18n.New()
	if err != nil {
		return nil, err
	}

	// help needs config and display
	helpSvc, err := help.New()
	if err != nil {
		return nil, err
	}

	// workspace needs config and io.Medium (io.Local is a concrete instance)
	workspaceSvc, err := workspace.New()
	if err != nil {
		return nil, err
	}

	// 3. Assemble the application container, exposing the concrete types.
	app := &Runtime{
		Config:  configSvc,
		Display: displaySvc,
		Help:    helpSvc,
		Crypt:   cryptSvc,
		I18n:    i18nSvc,
		//IO:        io.Local, // Assign io.Local directly
		Workspace: workspaceSvc,
	}

	return app, nil
}
