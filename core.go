package core

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// New initialises a Core instance using the provided options and performs the necessary setup.
// It is the primary entry point for creating a new Core application.
//
// Example:
//
//	core, err := core.New(
//		core.WithService(&MyService{}),
//		core.WithAssets(assets),
//	)
func New(opts ...Option) (*Core, error) {
	c := &Core{
		services: make(map[string]any),
		Features: &Features{},
	}
	for _, o := range opts {
		if err := o(c); err != nil {
			return nil, err
		}
	}
	c.once.Do(func() {
		c.initErr = nil
	})
	if c.initErr != nil {
		return nil, c.initErr
	}
	if c.serviceLock {
		c.servicesLocked = true
	}
	return c, nil
}

// WithService creates an Option that registers a service. It automatically discovers
// the service name from its package path and registers its IPC handler if it
// implements a method named `HandleIPCEvents`.
//
// Example:
//
//	// In myapp/services/calculator.go
//	package services
//
//	type Calculator struct{}
//
//	func (s *Calculator) Add(a, b int) int { return a + b }
//
//	// In main.go
//	import "myapp/services"
//
//	core.New(core.WithService(services.NewCalculator))
func WithService(factory func(*Core) (any, error)) Option {
	return func(c *Core) error {
		serviceInstance, err := factory(c)

		if err != nil {
			return fmt.Errorf("core: failed to create service: %w", err)
		}

		// --- Service Name Discovery ---
		typeOfService := reflect.TypeOf(serviceInstance)
		if typeOfService.Kind() == reflect.Ptr {
			typeOfService = typeOfService.Elem()
		}
		pkgPath := typeOfService.PkgPath()
		parts := strings.Split(pkgPath, "/")
		name := strings.ToLower(parts[len(parts)-1])

		// --- IPC Handler Discovery ---
		instanceValue := reflect.ValueOf(serviceInstance)
		handlerMethod := instanceValue.MethodByName("HandleIPCEvents")
		if handlerMethod.IsValid() {
			if handler, ok := handlerMethod.Interface().(func(*Core, Message) error); ok {
				c.RegisterAction(handler)
			}
		}

		return c.RegisterService(name, serviceInstance)
	}
}

// WithName creates an option that registers a service with a specific name.
// This is useful when the service name cannot be inferred from the package path,
// such as when using anonymous functions as factories.
// Note: Unlike WithService, this does not automatically discover or register
// IPC handlers. If your service needs IPC handling, implement HandleIPCEvents
// and register it manually.
func WithName(name string, factory func(*Core) (any, error)) Option {
	return func(c *Core) error {
		serviceInstance, err := factory(c)
		if err != nil {
			return fmt.Errorf("core: failed to create service '%s': %w", name, err)
		}
		return c.RegisterService(name, serviceInstance)
	}
}

// WithWails creates an Option that injects the Wails application instance into the Core.
// This is essential for services that need to interact with the Wails runtime.
func WithWails(app *application.App) Option {
	return func(c *Core) error {
		c.App = app
		return nil
	}
}

// WithAssets creates an Option that registers the application's embedded assets.
// This is necessary for the application to be able to serve its frontend.
func WithAssets(fs embed.FS) Option {
	return func(c *Core) error {
		c.assets = fs
		return nil
	}
}

// WithServiceLock creates an Option that prevents any further services from being
// registered after the Core has been initialized. This is a security measure to
// prevent late-binding of services that could have unintended consequences.
func WithServiceLock() Option {
	return func(c *Core) error {
		c.serviceLock = true
		return nil
	}
}

// --- Core Methods ---

// ServiceStartup is the entry point for the Core service's startup lifecycle.
// It is called by Wails when the application starts.
func (c *Core) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	return c.ACTION(ActionServiceStartup{})
}

// ServiceShutdown is the entry point for the Core service's shutdown lifecycle.
// It is called by Wails when the application shuts down.
func (c *Core) ServiceShutdown(ctx context.Context) error {
	return c.ACTION(ActionServiceShutdown{})
}

// ACTION dispatches a message to all registered IPC handlers.
// This is the primary mechanism for services to communicate with each other.
func (c *Core) ACTION(msg Message) error {
	c.ipcMu.RLock()
	handlers := append([]func(*Core, Message) error(nil), c.ipcHandlers...)
	c.ipcMu.RUnlock()

	var agg error
	for _, h := range handlers {
		if err := h(c, msg); err != nil {
			agg = fmt.Errorf("%w; %v", agg, err)
		}
	}
	return agg
}

// RegisterAction adds a new IPC handler to the Core.
func (c *Core) RegisterAction(handler func(*Core, Message) error) {
	c.ipcMu.Lock()
	c.ipcHandlers = append(c.ipcHandlers, handler)
	c.ipcMu.Unlock()
}

// RegisterActions adds multiple IPC handlers to the Core.
func (c *Core) RegisterActions(handlers ...func(*Core, Message) error) {
	c.ipcMu.Lock()
	c.ipcHandlers = append(c.ipcHandlers, handlers...)
	c.ipcMu.Unlock()
}

// RegisterService adds a new service to the Core.
func (c *Core) RegisterService(name string, api any) error {
	if c.servicesLocked {
		return fmt.Errorf("core: service %q is not permitted by the serviceLock setting", name)
	}
	if name == "" {
		return errors.New("core: service name cannot be empty")
	}
	c.serviceMu.Lock()
	defer c.serviceMu.Unlock()
	if _, exists := c.services[name]; exists {
		return fmt.Errorf("core: service %q already registered", name)
	}
	c.services[name] = api
	return nil
}

// Service retrieves a registered service by name.
// It returns nil if the service is not found.
func (c *Core) Service(name string) any {
	c.serviceMu.RLock()
	api, ok := c.services[name]
	c.serviceMu.RUnlock()
	if !ok {
		return nil
	}
	return api
}

// ServiceFor retrieves a registered service by name and asserts its type to the given interface T.
func ServiceFor[T any](c *Core, name string) (T, error) {
	var zero T
	raw := c.Service(name)
	if raw == nil {
		return zero, fmt.Errorf("service '%s' not found", name)
	}
	typed, ok := raw.(T)
	if !ok {
		return zero, fmt.Errorf("service '%s' is of type %T, but expected %T", name, raw, zero)
	}
	return typed, nil
}

// MustServiceFor retrieves a registered service by name and asserts its type to the given interface T.
// It panics if the service is not found or cannot be cast to T.
func MustServiceFor[T any](c *Core, name string) T {
	svc, err := ServiceFor[T](c, name)
	if err != nil {
		panic(err)
	}
	return svc
}

// App returns the global application instance.
// It panics if the Core has not been initialized.
func App() *application.App {
	if instance == nil {
		panic("core.App() called before core.Setup() was successfully initialized")
	}
	return instance.App
}

// Config returns the registered Config service.
func (c *Core) Config() Config {
	cfg := MustServiceFor[Config](c, "config")
	return cfg
}

// Display returns the registered Display service.
func (c *Core) Display() Display {
	display := MustServiceFor[Display](c, "display")
	return display
}

// Core returns the Core instance itself.
func (c *Core) Core() *Core { return c }

// Assets returns the embedded filesystem containing the application's assets.
func (c *Core) Assets() embed.FS {
	return c.assets
}
