package runtime

import (
	// Import the CONCRETE implementations from the internal packages.
	"github.com/Snider/Core/pkg/config"
	"github.com/Snider/Core/pkg/crypt"
	"github.com/Snider/Core/pkg/display"
	"github.com/Snider/Core/pkg/help"
)

// App is the runtime container that holds all instantiated services.
// It holds the concrete service types so they can be registered with Wails.
type App struct {
	Config  *config.Service
	Display *display.Service
	Help    *help.Service
	Crypt   *crypt.Service
}

// New creates and wires together all application services.
func New() (*App, error) {
	// 1. Instantiate services.
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

	// 2. Inject dependencies.
	helpSvc, _ := help.New(configSvc, displaySvc)

	// 3. Assemble the application container.
	app := &App{
		Config:  configSvc,
		Display: displaySvc,
		Help:    helpSvc,
		Crypt:   cryptSvc,
	}

	return app, nil
}
