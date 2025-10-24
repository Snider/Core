package core

import (
	"context"
	"embed"
	"errors"
	"fmt"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// Service initialises a Core instance using the provided options and performs the necessary setup.
func Service(opts ...Option) *Core {
	c := &Core{
		mods: make(map[string]any),
	}
	// Apply all options (including WithService calls)
	for _, o := range opts {
		if err := o(c); err != nil {
			return nil
		}
	}
	c.once.Do(func() {
		// any one‑time initialisation you need
		instance = c
		c.initErr = nil
	})
	if c.initErr != nil {
		return nil
	}
	if c.serviceLock {
		c.servicesLocked = true
	}
	return c
}

// WithService wraps a function that registers a package or module with the provided Core instance as an Option.
func WithService(reg func(*Core) error) Option {
	return func(c *Core) error {
		return reg(c)
	}
}

// WithWails sets the Wails application instance to the Core configuration and returns an Option function.
func WithWails(app *application.App) Option {
	return func(c *Core) error {
		c.App = app
		return nil
	}
}

// WithAssets sets the provided embedded filesystem as the assets for the Core instance.
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

// ServiceStartup initializes the service during application startup by executing the ActionServiceStartup message.
func (c *Core) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	return c.ACTION(ActionServiceStartup{})
}

// ACTION processes a Message by invoking all registered handlers and returns an aggregated error if any handlers fail.
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

// RegisterAction adds a single handler function to the list of registered IPC handlers in a thread-safe manner.
func (c *Core) RegisterAction(handler func(*Core, Message) error) {
	c.ipcMu.Lock()
	c.ipcHandlers = append(c.ipcHandlers, handler)
	c.ipcMu.Unlock()
}

// RegisterActions registers multiple IPC handler functions to be executed during message processing in a thread-safe manner.
func (c *Core) RegisterActions(handlers ...func(*Core, Message) error) {
	c.ipcMu.Lock()
	c.ipcHandlers = append(c.ipcHandlers, handlers...)
	c.ipcMu.Unlock()
}

// RegisterModule inserts an API object under a unique name.
func (c *Core) RegisterModule(name string, api any) error {

	if c.servicesLocked {
		return fmt.Errorf("core: module %q is not permitted by the serviceLock setting", name)
	}

	if name == "" {
		return errors.New("core: module name cannot be empty")
	}
	c.modMu.Lock()
	defer c.modMu.Unlock()
	if _, exists := c.mods[name]; exists {
		return fmt.Errorf("core: module %q already registered", name)
	}
	c.mods[name] = api
	return nil
}

// Mod caller must type‑assert the result to the concrete API type it expects.
func (c *Core) Mod(name string) any {
	c.modMu.RLock()
	api, ok := c.mods[name]
	c.modMu.RUnlock()
	if !ok {
		return nil
	}
	return api
}

// Mod is a generic helper to get a module of expected type T.
func Mod[T any](c *Core, name string) *T {
	raw := c.Mod(name)
	typed, ok := raw.(*T)
	if !ok {
		return nil
	}
	return typed
}
