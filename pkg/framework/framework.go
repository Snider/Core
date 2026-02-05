// Package framework provides the Core DI/service framework.
// Import this package for cleaner access to the framework types.
//
// Usage:
//
//	import "github.com/host-uk/core/pkg/framework"
//
//	app, _ := framework.New(
//	    framework.WithServiceLock(),
//	)
package framework

import (
	"github.com/host-uk/core/pkg/framework/core"
)

// Re-export core types for cleaner imports
type (
	Core                  = core.Core
	Option                = core.Option
	Message               = core.Message
	Query                 = core.Query
	Task                  = core.Task
	QueryHandler          = core.QueryHandler
	TaskHandler           = core.TaskHandler
	Startable             = core.Startable
	Stoppable             = core.Stoppable
	Config                = core.Config
	Display               = core.Display
	WindowOption          = core.WindowOption
	Features              = core.Features
	Contract              = core.Contract
	Error                 = core.Error
	ServiceRuntime[T any] = core.ServiceRuntime[T]
	Runtime               = core.Runtime
	ServiceFactory        = core.ServiceFactory
)

// Re-export core functions
var (
	New              = core.New
	WithService      = core.WithService
	WithName         = core.WithName
	WithApp          = core.WithApp
	WithAssets       = core.WithAssets
	WithServiceLock  = core.WithServiceLock
	App              = core.App
	E                = core.E
	NewRuntime       = core.NewRuntime
	NewWithFactories = core.NewWithFactories
)

// NewServiceRuntime creates a new ServiceRuntime for a service.
func NewServiceRuntime[T any](c *Core, opts T) *ServiceRuntime[T] {
	return core.NewServiceRuntime(c, opts)
}

// ServiceFor retrieves a typed service from the core container by name.
func ServiceFor[T any](c *Core, name string) (T, error) {
	return core.ServiceFor[T](c, name)
}

// MustServiceFor retrieves a typed service or returns an error if not found.
func MustServiceFor[T any](c *Core, name string) T {
	return core.MustServiceFor[T](c, name)
}

// Action types
type (
	ActionServiceStartup  = core.ActionServiceStartup
	ActionServiceShutdown = core.ActionServiceShutdown
)
