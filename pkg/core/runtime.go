package core

// App is the runtime container that holds all instantiated services.
// It holds the concrete service types so they can be registered with Wails.
//type App struct {
//	Config  *config.Service
//	Display *display.Service
//	Docs    *docs.Service
//	Crypt   *crypt.Service
//}
//
//// New creates and wires together all application services.
//func New() (*App, error) {
//	// 1. Instantiate services.
//	configSvc, err := config.New()
//	if err != nil {
//		return nil, err
//	}
//
//	displaySvc, err := display.New()
//	if err != nil {
//		return nil, err
//	}
//
//	cryptSvc, err := crypt.New()
//	if err != nil {
//		return nil, err
//	}
//
//	// 2. Inject dependencies.
//	docsSvc := docs.New(configSvc, displaySvc)
//
//	// 3. Assemble the application container.
//	app := &App{
//		Config:  configSvc,
//		Display: displaySvc,
//		Docs:    docsSvc,
//		Crypt:   cryptSvc,
//	}
//
//	return app, nil
//}
//
//func NewRuntime[T any](c *Core, opts T) *Runtime[T] {
//	return &Runtime[T]{
//		core:    c,
//		Options: opts,
//	}
//}
//
//type Runtime[T any] struct {
//	core    *Core
//	Options T
//}
//
//func (r *Runtime[T]) Core() *Core {
//	return r.core
//}
