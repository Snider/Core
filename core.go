// Package core provides the primary public API for the Core framework.
// It acts as a facade, re-exporting types and functions from the internal
// core package to provide a clean, root-level import path.
package core

import (
	// Import the internal core package which contains the actual definitions.
	core "github.com/Snider/Core/pkg/core"
)

// --- Primary Types & Constructors ---

// Core is the main application container.
type Core = core.Core

// New is the primary constructor for the Core framework.
var New = core.New

// --- Core Options ---

// WithService is a helper function to create a service option.
var WithService = core.WithService

// WithWails provides the Wails application instance to the core.
var WithWails = core.WithWails

// WithAssets provides the application's assets to the core.
var WithAssets = core.WithAssets

// WithServiceLock prevents new services from being registered after startup.
var WithServiceLock = core.WithServiceLock

// --- Service Runtime ---

// Runtime is a helper struct embedded in services to provide access to the core application.
type Runtime[T any] = core.Runtime[T]

// NewRuntime creates a new Runtime instance for a service.
func NewRuntime[T any](c *Core, opts T) *Runtime[T] {
	return core.NewRuntime(c, opts)
}

// --- Messages & Actions ---

// Message is the interface for all IPC messages.
type Message = core.Message

// ActionServiceStartup is a message sent when services should perform their startup tasks.
type ActionServiceStartup = core.ActionServiceStartup

// --- Service Interfaces (from pkg/core/interfaces.go) ---

// Config is the public interface for the configuration service.
type Config = core.Config

// Display is the public interface for the display service.
type Display = core.Display

// Docs is the public interface for the documentation service.
type Docs = core.Docs

// Crypt is the public interface for the cryptography service.
type Crypt = core.Crypt

// I18n is the public interface for the internationalization service.
type I18n = core.I18n

// Workspace is the public interface for the workspace service.
type Workspace = core.Workspace
