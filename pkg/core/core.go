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
		name := parts[len(parts)-1]

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

func WithWails(app *application.App) Option {
	return func(c *Core) error {
		c.App = app
		return nil
	}
}

func WithAssets(fs embed.FS) Option {
	return func(c *Core) error {
		c.assets = fs
		return nil
	}
}

func WithServiceLock() Option {
	return func(c *Core) error {
		c.serviceLock = true
		return nil
	}
}

// --- Core Methods ---

func (c *Core) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	return c.ACTION(ActionServiceStartup{})
}

func (c *Core) ServiceShutdown(ctx context.Context) error {
	return c.ACTION(ActionServiceShutdown{})
}

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

func (c *Core) RegisterAction(handler func(*Core, Message) error) {
	c.ipcMu.Lock()
	c.ipcHandlers = append(c.ipcHandlers, handler)
	c.ipcMu.Unlock()
}

func (c *Core) RegisterActions(handlers ...func(*Core, Message) error) {
	c.ipcMu.Lock()
	c.ipcHandlers = append(c.ipcHandlers, handlers...)
	c.ipcMu.Unlock()
}

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

func (c *Core) Core() *Core { return c }
