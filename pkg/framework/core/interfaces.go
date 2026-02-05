package core

import (
	"context"
	"embed"
	goio "io"
)

// This file defines the public API contracts (interfaces) for the services
// in the Core framework. Services depend on these interfaces, not on
// concrete implementations.

// Contract specifies the operational guarantees that the Core and its services must adhere to.
// This is used for configuring panic handling and other resilience features.
type Contract struct {
	// DontPanic, if true, instructs the Core to recover from panics and return an error instead.
	DontPanic bool
	// DisableLogging, if true, disables all logging from the Core and its services.
	DisableLogging bool
}

// Features provides a way to check if a feature is enabled.
// This is used for feature flagging and conditional logic.
type Features struct {
	// Flags is a list of enabled feature flags.
	Flags []string
}

// IsEnabled returns true if the given feature is enabled.
func (f *Features) IsEnabled(feature string) bool {
	for _, flag := range f.Flags {
		if flag == feature {
			return true
		}
	}
	return false
}

// Option is a function that configures the Core.
// This is used to apply settings and register services during initialization.
type Option func(*Core) error

// Message is the interface for all messages that can be sent through the Core's IPC system.
// Any struct can be a message, allowing for structured data to be passed between services.
// Used with ACTION for fire-and-forget broadcasts.
// For type-safe dispatch and registration, use core.Action and core.RegisterAction.
type Message interface{}

// Ask is the interface for read-only requests that return data.
// Used with QUERY (first responder) or QUERYALL (all responders).
// For type-safe dispatch and registration, use core.Query and core.RegisterQuery.
type Ask interface{}

// Task is the interface for requests that perform side effects.
// Used with PERFORM (first responder executes).
// For type-safe dispatch and registration, use core.Perform and core.RegisterTask.
type Task interface{}

// Request is a marker interface for queries and tasks that returns a result of type R.
// This allows for type inference in generic dispatch functions like core.DispatchAsk and core.DispatchTask.
type Request[R any] interface {
	// Response is a dummy method used to associate the response type R with the request type.
	// It is not intended to be called at runtime.
	Response() R
}

// QueryHandler handles Ask requests. Returns (result, handled, error).
// If handled is false, the query will be passed to the next handler.
// Deprecated: use TypedQueryHandler with core.RegisterAsk instead.
type QueryHandler func(*Core, Query) (any, bool, error)

// TaskHandler handles Task requests. Returns (result, handled, error).
// If handled is false, the task will be passed to the next handler.
// Deprecated: use TypedTaskHandler with core.RegisterTask instead.
type TaskHandler func(*Core, Task) (any, bool, error)

// TypedQueryHandler handles Ask requests of type Q returning R.
type TypedQueryHandler[Q any, R any] func(*Core, Q) (R, bool, error)

// TypedTaskHandler handles Task requests of type T returning R.
type TypedTaskHandler[T any, R any] func(*Core, T) (R, bool, error)

// Startable is an interface for services that need to perform initialization.
type Startable interface {
	OnStartup(ctx context.Context) error
}

// Stoppable is an interface for services that need to perform cleanup.
type Stoppable interface {
	OnShutdown(ctx context.Context) error
}

// Core is the central application object that manages services, assets, and communication.
type Core struct {
	App      any // GUI runtime (e.g., Wails App) - set by WithApp option
	assets   embed.FS
	Features *Features
	svc      *serviceManager
	bus      *messageBus
}

// Config provides access to application configuration.
type Config interface {
	// Get retrieves a configuration value by key and stores it in the 'out' variable.
	Get(key string, out any) error
	// Set stores a configuration value by key.
	Set(key string, v any) error
}

// WindowOption is an interface for applying configuration options to a window.
type WindowOption interface {
	Apply(any)
}

// Display provides access to windowing and visual elements.
type Display interface {
	// OpenWindow creates a new window with the given options.
	OpenWindow(opts ...WindowOption) error
}

// Workspace provides management for encrypted user workspaces.
type Workspace interface {
	// CreateWorkspace creates a new encrypted workspace.
	CreateWorkspace(identifier, password string) (string, error)
	// SwitchWorkspace changes the active workspace.
	SwitchWorkspace(name string) error
	// WorkspaceFileGet retrieves the content of a file from the active workspace.
	WorkspaceFileGet(filename string) (string, error)
	// WorkspaceFileSet saves content to a file in the active workspace.
	WorkspaceFileSet(filename, content string) error
}

// Crypt provides PGP-based encryption, signing, and key management.
type Crypt interface {
	// CreateKeyPair generates a new PGP keypair.
	CreateKeyPair(name, passphrase string) (string, error)
	// EncryptPGP encrypts data for a recipient.
	EncryptPGP(writer goio.Writer, recipientPath, data string, opts ...any) (string, error)
	// DecryptPGP decrypts a PGP message.
	DecryptPGP(recipientPath, message, passphrase string, opts ...any) (string, error)
}

// ActionServiceStartup is a message sent when the application's services are starting up.
// This provides a hook for services to perform initialization tasks.
type ActionServiceStartup struct{}

// ActionServiceShutdown is a message sent when the application is shutting down.
// This allows services to perform cleanup tasks, such as saving state or closing resources.
type ActionServiceShutdown struct{}
