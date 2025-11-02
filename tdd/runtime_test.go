package tdd

import (
	"testing"

	"github.com/Snider/Core/pkg/config"
	"github.com/Snider/Core/pkg/crypt"
	"github.com/Snider/Core/pkg/display"
	"github.com/Snider/Core/pkg/help"
	"github.com/Snider/Core/pkg/i18n"
	"github.com/Snider/Core/pkg/runtime"
	"github.com/Snider/Core/pkg/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// This is a helper function to create a new runtime with mock services
func newTestRuntime(app *application.App) (*runtime.Runtime, error) {
	return runtime.NewWithFactories(app, map[string]runtime.ServiceFactory{
		"config":    func() (any, error) { return &config.Service{}, nil },
		"display":   func() (any, error) { return &display.Service{}, nil },
		"help":      func() (any, error) { return &help.Service{}, nil },
		"crypt":     func() (any, error) { return &crypt.Service{}, nil },
		"i18n":      func() (any, error) { return &i18n.Service{}, nil },
		"workspace": func() (any, error) { return &workspace.Service{}, nil },
	})
}

func TestNew_Good(t *testing.T) {
	rt, err := newTestRuntime(nil)
	assert.NoError(t, err)
	assert.NotNil(t, rt)
	assert.NotNil(t, rt.Core)
	assert.NotNil(t, rt.Config)
	assert.NotNil(t, rt.Display)
	assert.NotNil(t, rt.Help)
	assert.NotNil(t, rt.Crypt)
	assert.NotNil(t, rt.I18n)
	assert.NotNil(t, rt.Workspace)
}
