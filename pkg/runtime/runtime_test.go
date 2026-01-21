package runtime

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Snider/Core/pkg/config"
	"github.com/Snider/Core/pkg/crypt"
	"github.com/Snider/Core/pkg/display"
	"github.com/Snider/Core/pkg/docs"
	"github.com/Snider/Core/pkg/help"
	"github.com/Snider/Core/pkg/ide"
	"github.com/Snider/Core/pkg/io"
	"github.com/Snider/Core/pkg/module"
	"github.com/Snider/Core/pkg/workspace"
)

// TestNew ensures that New correctly initializes a Runtime instance.
func TestNew(t *testing.T) {
	runtime, err := New()
	assert.NoError(t, err)
	assert.NotNil(t, runtime)

	// Assert that key services are initialized
	assert.NotNil(t, runtime.Core, "Core service should be initialized")
	assert.NotNil(t, runtime.Config, "Config service should be initialized")
	assert.NotNil(t, runtime.Display, "Display service should be initialized")
	assert.NotNil(t, runtime.Help, "Help service should be initialized")
	assert.NotNil(t, runtime.Crypt, "Crypt service should be initialized")
	assert.NotNil(t, runtime.I18n, "I18n service should be initialized")
	assert.NotNil(t, runtime.Workspace, "Workspace service should be initialized")

	// Verify services are properly wired through Core
	configFromCore := runtime.Core.Service("config")
	assert.NotNil(t, configFromCore, "Config should be registered in Core")
	assert.Equal(t, runtime.Config, configFromCore, "Config from Core should match direct reference")

	displayFromCore := runtime.Core.Service("display")
	assert.NotNil(t, displayFromCore, "Display should be registered in Core")
	assert.Equal(t, runtime.Display, displayFromCore, "Display from Core should match direct reference")

	helpFromCore := runtime.Core.Service("help")
	assert.NotNil(t, helpFromCore, "Help should be registered in Core")
	assert.Equal(t, runtime.Help, helpFromCore, "Help from Core should match direct reference")

	cryptFromCore := runtime.Core.Service("crypt")
	assert.NotNil(t, cryptFromCore, "Crypt should be registered in Core")
	assert.Equal(t, runtime.Crypt, cryptFromCore, "Crypt from Core should match direct reference")

	i18nFromCore := runtime.Core.Service("i18n")
	assert.NotNil(t, i18nFromCore, "I18n should be registered in Core")
	assert.Equal(t, runtime.I18n, i18nFromCore, "I18n from Core should match direct reference")

	workspaceFromCore := runtime.Core.Service("workspace")
	assert.NotNil(t, workspaceFromCore, "Workspace should be registered in Core")
	assert.Equal(t, runtime.Workspace, workspaceFromCore, "Workspace from Core should match direct reference")
}

// TestNewServiceInitializationError tests the error path in New.
func TestNewServiceInitializationError(t *testing.T) {
	factories := map[string]ServiceFactory{
		"config":    func() (any, error) { return config.New() },
		"display":   func() (any, error) { return display.New() },
		"docs":      func() (any, error) { return docs.New(docs.Options{BaseURL: "https://docs.lethean.io"}) },
		"help":      func() (any, error) { return help.New(help.Options{}) },
		"crypt":     func() (any, error) { return crypt.New() },
		"i18n":      func() (any, error) { return nil, errors.New("i18n service failed to initialize") }, // This factory will fail
		"ide":       func() (any, error) { return ide.New() },
		"module":    func() (any, error) { return module.NewService(module.Options{AppsDir: "apps"}) },
		"workspace": func() (any, error) { return workspace.New(io.Local) },
	}

	runtime, err := newWithFactories(factories)

	assert.Error(t, err)
	assert.Nil(t, runtime)
	assert.Contains(t, err.Error(), "failed to create service i18n: i18n service failed to initialize")
}

// TestMissingFactory tests error when a factory is not provided.
func TestMissingFactory(t *testing.T) {
	// Missing config factory
	factories := map[string]ServiceFactory{
		// "config" intentionally missing
		"display":   func() (any, error) { return display.New() },
		"docs":      func() (any, error) { return docs.New(docs.Options{BaseURL: "https://docs.lethean.io"}) },
		"help":      func() (any, error) { return help.New(help.Options{}) },
		"crypt":     func() (any, error) { return crypt.New() },
		"i18n":      func() (any, error) { return nil, nil },
		"ide":       func() (any, error) { return ide.New() },
		"module":    func() (any, error) { return module.NewService(module.Options{AppsDir: "apps"}) },
		"workspace": func() (any, error) { return workspace.New(io.Local) },
	}

	runtime, err := newWithFactories(factories)
	assert.Error(t, err)
	assert.Nil(t, runtime)
	assert.Contains(t, err.Error(), "service config factory not provided")
}

// Note: TestWrongTypeFactory removed because the core.WithService option
// requires services to implement specific interfaces (like Name() method).
// The type assertion error paths (lines 58-80) are guarded by core.New()
// which fails first for invalid service types, making those lines
// unreachable in practice. This is defensive code that protects against
// programming errors rather than runtime errors.
