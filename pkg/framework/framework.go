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
	Core                            = core.Core
	Option                          = core.Option
	Message                         = core.Message
	Query                           = core.Query
	Task                            = core.Task
	QueryHandler                    = core.QueryHandler
	TaskHandler                     = core.TaskHandler
	Startable                       = core.Startable
	Stoppable                       = core.Stoppable
	Config                          = core.Config
	Display                         = core.Display
	WindowOption                    = core.WindowOption
	Features                        = core.Features
	Contract                        = core.Contract
	Error                           = core.Error
	ServiceRuntime[T any]           = core.ServiceRuntime[T]
	Runtime                         = core.Runtime
	ServiceFactory                  = core.ServiceFactory
	Request[R any]                  = core.Request[R]
	TypedQueryHandler[Q any, R any] = core.TypedQueryHandler[Q, R]
	TypedTaskHandler[T any, R any]  = core.TypedTaskHandler[T, R]
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

// Action dispatches a message of type T to all registered IPC handlers.
func Action[T any](c *Core, msg T) error {
	return core.Action(c, msg)
}

// RegisterAction adds a type-safe IPC handler to the Core.
func RegisterAction[T any](c *Core, handler func(*Core, T) error) {
	core.RegisterAction(c, handler)
}

// Query dispatches a query to handlers until one responds, returning a typed result.
func Query[R any](c *Core, q any) (R, bool, error) {
	return core.Query[R](c, q)
}

// QueryAll dispatches a query to all handlers and collects typed responses.
func QueryAll[R any](c *Core, q any) ([]R, error) {
	return core.QueryAll[R](c, q)
}

// RegisterQuery adds a type-safe query handler to the Core.
func RegisterQuery[Q any, R any](c *Core, handler TypedQueryHandler[Q, R]) {
	core.RegisterQuery(c, handler)
}

// Perform dispatches a task to handlers until one responds, returning a typed result.
func Perform[R any](c *Core, t any) (R, bool, error) {
	return core.Perform[R](c, t)
}

// RegisterTask adds a type-safe task handler to the Core.
func RegisterTask[T any, R any](c *Core, handler TypedTaskHandler[T, R]) {
	core.RegisterTask(c, handler)
}

// DispatchQuery dispatches a query that implements Request[R], using type inference for the result.
func DispatchQuery[R any](c *Core, q Request[R]) (R, bool, error) {
	return core.DispatchQuery(c, q)
}

// DispatchTask dispatches a task that implements Request[R], using type inference for the result.
func DispatchTask[R any](c *Core, t Request[R]) (R, bool, error) {
	return core.DispatchTask(c, t)
}

// NewServiceRuntime creates a new ServiceRuntime for a service.
func NewServiceRuntime[T any](c *Core, opts T) *ServiceRuntime[T] {
	return core.NewServiceRuntime(c, opts)
}

// ServiceFor retrieves a typed service from the core container by name.
func ServiceFor[T any](c *Core, name string) (T, error) {
	return core.ServiceFor[T](c, name)
}

// MustServiceFor retrieves a typed service or returns an error if not found.
//
// Deprecated: use ServiceFor instead. This function does not panic on failure
// and is retained only for backward compatibility.
func MustServiceFor[T any](c *Core, name string) (T, error) {
	return core.MustServiceFor[T](c, name)
}

// Action types
type (
	ActionServiceStartup  = core.ActionServiceStartup
	ActionServiceShutdown = core.ActionServiceShutdown
)
