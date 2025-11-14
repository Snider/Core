package core

import (
	"embed"
	"io"
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
}

var instance *Core

// Config provides access to application configuration.
type Config interface {
	// Get retrieves a configuration value by key and stores it in the 'out' variable.
	Get(key string, out any) error
	// Set stores a configuration value by key.
	Set(key string, v any) error
}

// WindowConfig represents the configuration for a window.
type WindowConfig struct {
	Name   string
	Title  string
	URL    string
	Width  int
	Height int // Add other common window options here as needed
}

// WindowOption configures window creation.
type WindowOption interface {
	// Apply applies the window option to the given configuration.
	Apply(*WindowConfig)
}

// Display manages windows and UI.
type Display interface {
	// OpenWindow creates and displays a new window with the given options.
	OpenWindow(opts ...WindowOption) error
}

// Help manages the in-app documentation and help system.
type Help interface {
	// Show displays the main help topic.
	Show() error
	// ShowAt displays the help topic for the given anchor.
	ShowAt(anchor string) error
}

// Crypt provides cryptographic functions.
type Crypt interface {
	// EncryptPGP encrypts data using PGP and writes the result to the given writer.
	EncryptPGP(writer io.Writer, recipientPath, data string, signerPath, signerPassphrase *string) (string, error)
	// DecryptPGP decrypts a PGP message.
	DecryptPGP(recipientPath, message, passphrase string, signerPath *string) (string, error)
}

// I18n provides internationalization and localization services.
type I18n interface {
	// Translate returns the translated string for the given key.
	Translate(key string) string
	// SetLanguage changes the active language.
	SetLanguage(lang string) error
}

// Workspace manages user workspaces.
type Workspace interface {
	// CreateWorkspace creates a new workspace with the given identifier and password.
	CreateWorkspace(identifier, password string) (string, error)
	// SwitchWorkspace changes the active workspace.
	SwitchWorkspace(name string) error
	// WorkspaceFileGet retrieves the content of a file from the current workspace.
	WorkspaceFileGet(filename string) (string, error)
	// WorkspaceFileSet writes content to a file in the current workspace.
	WorkspaceFileSet(filename, content string) error
}
