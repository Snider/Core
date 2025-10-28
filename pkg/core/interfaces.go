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

type Contract struct {
	DontPanic      bool
	DisableLogging bool
}
type Option func(*Core) error
type Message interface{}
type Core struct {
	once           sync.Once
	initErr        error
	App            *application.App
	assets         embed.FS
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
	Get(key string, out any) error
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
	Apply(*WindowConfig)
}

// Display manages windows and UI.
type Display interface {
	OpenWindow(opts ...WindowOption) error
}

// Help manages the in-app documentation and help system.
type Help interface {
	Show() error
	ShowAt(anchor string) error
}

// Crypt provides cryptographic functions.
type Crypt interface {
	EncryptPGP(writer io.Writer, recipientPath, data string, signerPath, signerPassphrase *string) (string, error)
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
	CreateWorkspace(identifier, password string) (string, error)
	SwitchWorkspace(name string) error
	WorkspaceFileGet(filename string) (string, error)
	WorkspaceFileSet(filename, content string) error
}
