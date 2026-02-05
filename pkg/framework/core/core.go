package core

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

var (
	instance   *Core
	instanceMu sync.RWMutex
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
		Features: &Features{},
		svc:      newServiceManager(),
	}
	c.bus = newMessageBus(c)

	for _, o := range opts {
		if err := o(c); err != nil {
			return nil, err
		}
	}

	c.svc.applyLock()
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

// WithApp creates an Option that injects the GUI runtime (e.g., Wails App) into the Core.
// This is essential for services that need to interact with the GUI runtime.
func WithApp(app any) Option {
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
		c.svc.enableLock()
		return nil
	}
}

// --- Core Methods ---

// ServiceStartup is the entry point for the Core service's startup lifecycle.
// It is called by the GUI runtime when the application starts.
func (c *Core) ServiceStartup(ctx context.Context, options any) error {
	startables := c.svc.getStartables()

	var agg error
	for _, s := range startables {
		if err := s.OnStartup(ctx); err != nil {
			agg = errors.Join(agg, err)
		}
	}

	if err := c.ACTION(ActionServiceStartup{}); err != nil {
		agg = errors.Join(agg, err)
	}

	return agg
}

// ServiceShutdown is the entry point for the Core service's shutdown lifecycle.
// It is called by the GUI runtime when the application shuts down.
func (c *Core) ServiceShutdown(ctx context.Context) error {
	var agg error
	if err := c.ACTION(ActionServiceShutdown{}); err != nil {
		agg = errors.Join(agg, err)
	}

	stoppables := c.svc.getStoppables()
	for i := len(stoppables) - 1; i >= 0; i-- {
		if err := stoppables[i].OnShutdown(ctx); err != nil {
			agg = errors.Join(agg, err)
		}
	}

	return agg
}

// ACTION dispatches a message to all registered IPC handlers.
// This is the primary mechanism for services to communicate with each other.
func (c *Core) ACTION(msg Message) error {
	return c.bus.action(msg)
}

// RegisterAction adds a new IPC handler to the Core.
func (c *Core) RegisterAction(handler func(*Core, Message) error) {
	c.bus.registerAction(handler)
}

// RegisterActions adds multiple IPC handlers to the Core.
func (c *Core) RegisterActions(handlers ...func(*Core, Message) error) {
	c.bus.registerActions(handlers...)
}

// QUERY dispatches a query to handlers until one responds.
// Returns (result, handled, error). If no handler responds, handled is false.
func (c *Core) QUERY(q Query) (any, bool, error) {
	return c.bus.query(q)
}

// QUERYALL dispatches a query to all handlers and collects all responses.
// Returns all results from handlers that responded.
func (c *Core) QUERYALL(q Query) ([]any, error) {
	return c.bus.queryAll(q)
}

// PERFORM dispatches a task to handlers until one executes it.
// Returns (result, handled, error). If no handler responds, handled is false.
func (c *Core) PERFORM(t Task) (any, bool, error) {
	return c.bus.perform(t)
}

// RegisterQuery adds a query handler to the Core.
func (c *Core) RegisterQuery(handler QueryHandler) {
	c.bus.registerQuery(handler)
}

// RegisterTask adds a task handler to the Core.
func (c *Core) RegisterTask(handler TaskHandler) {
	c.bus.registerTask(handler)
}

// RegisterService adds a new service to the Core.
func (c *Core) RegisterService(name string, api any) error {
	return c.svc.registerService(name, api)
}

// Service retrieves a registered service by name.
// It returns nil if the service is not found.
func (c *Core) Service(name string) any {
	return c.svc.service(name)
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
// It panics if the Core has not been initialized via SetInstance.
// This is typically used by GUI runtimes that need global access.
func App() any {
	instanceMu.RLock()
	inst := instance
	instanceMu.RUnlock()
	if inst == nil {
		panic("core.App() called before core.SetInstance()")
	}
	return inst.App
}

// SetInstance sets the global Core instance for App() access.
// This is typically called by GUI runtimes during initialization.
func SetInstance(c *Core) {
	instanceMu.Lock()
	instance = c
	instanceMu.Unlock()
}

// GetInstance returns the global Core instance, or nil if not set.
// Use this for non-panicking access to the global instance.
func GetInstance() *Core {
	instanceMu.RLock()
	inst := instance
	instanceMu.RUnlock()
	return inst
}

// ClearInstance resets the global Core instance to nil.
// This is primarily useful for testing to ensure a clean state between tests.
func ClearInstance() {
	instanceMu.Lock()
	instance = nil
	instanceMu.Unlock()
}

// Config returns the registered Config service.
func (c *Core) Config() Config {
	cfg := MustServiceFor[Config](c, "config")
	return cfg
}

// Display returns the registered Display service.
func (c *Core) Display() Display {
	d := MustServiceFor[Display](c, "display")
	return d
}

// Core returns self, implementing the CoreProvider interface.
func (c *Core) Core() *Core { return c }

// Assets returns the embedded filesystem containing the application's assets.
func (c *Core) Assets() embed.FS {
	return c.assets
}
