package core

import "io"

// This file defines the public API contracts (interfaces) for the services
// in the Core framework. Services depend on these interfaces, not on
// concrete implementations.

// Config provides access to application configuration.
type Config interface {
	Get(path string, out any) error
	Set(path string, v any) error
}

// Display manages windows and UI.
type Display interface {
	OpenWindow(opts ...any) error // Simplified for now
}

// Docs manages the in-app documentation.
type Docs interface {
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
