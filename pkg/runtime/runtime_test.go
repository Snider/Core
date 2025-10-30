package runtime

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
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
		"failing": func() (any, error) {
			return nil, errors.New("initialization failed")
		},
	}

	runtime, err := newWithFactories(factories)

	assert.Error(t, err)
	assert.Nil(t, runtime)
	assert.Contains(t, err.Error(), "failed to create service failing: initialization failed")
}

// Removed TestRuntimeOptions and TestRuntimeCore as these methods no longer exist on the Runtime struct.
