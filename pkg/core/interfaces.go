package core

import (
	"context"
	"embed"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
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
type Message interface{}

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
	once           sync.Once
	initErr        error
	App            *application.App
	assets         embed.FS
	Features       *Features
	serviceLock    bool
	ipcMu          sync.RWMutex
	ipcHandlers    []func(*Core, Message) error
	serviceMu      sync.RWMutex
	services       map[string]any
	servicesLocked bool
	startables     []Startable
	stoppables     []Stoppable
}

var instance *Core

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

// ActionServiceStartup is a message sent when the application's services are starting up.
// This provides a hook for services to perform initialization tasks.
type ActionServiceStartup struct{}

// ActionServiceShutdown is a message sent when the application is shutting down.
// This allows services to perform cleanup tasks, such as saving state or closing resources.
type ActionServiceShutdown struct{}
