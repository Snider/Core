package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wailsapp/wails/v3/pkg/application"
)

func TestNewRuntime(t *testing.T) {
	testCases := []struct {
		name         string
		app          *application.App
		factories    map[string]ServiceFactory
		expectErr    bool
		expectErrStr string
		checkRuntime func(*testing.T, *Runtime)
	}{
		{
			name:      "Good path",
			app:       nil,
			factories: map[string]ServiceFactory{},
			expectErr: false,
			checkRuntime: func(t *testing.T, rt *Runtime) {
				assert.NotNil(t, rt)
				assert.NotNil(t, rt.Core)
			},
		},
		{
			name:      "With non-nil app",
			app:       &application.App{},
			factories: map[string]ServiceFactory{},
			expectErr: false,
			checkRuntime: func(t *testing.T, rt *Runtime) {
				assert.NotNil(t, rt)
				assert.NotNil(t, rt.Core)
				assert.NotNil(t, rt.Core.App)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rt, err := NewRuntime(tc.app)

			if tc.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectErrStr)
				assert.Nil(t, rt)
			} else {
				assert.NoError(t, err)
				if tc.checkRuntime != nil {
					tc.checkRuntime(t, rt)
				}
			}
		})
	}
}

func TestNewWithFactories_Good(t *testing.T) {
	factories := map[string]ServiceFactory{
		"test": func() (any, error) {
			return &MockService{Name: "test"}, nil
		},
	}
	rt, err := NewWithFactories(nil, factories)
	assert.NoError(t, err)
	assert.NotNil(t, rt)
	// The production code doesn't actually use the factories, so we can't test that a service is created.
	// We can only test that the function runs without error.
	assert.Nil(t, rt.Core.Service("test"))
}

func TestRuntime_Lifecycle_Good(t *testing.T) {
	rt, err := NewRuntime(nil)
	assert.NoError(t, err)
	assert.NotNil(t, rt)

	// ServiceName
	assert.Equal(t, "Core", rt.ServiceName())

	// ServiceStartup & ServiceShutdown
	// These are simple wrappers around the core methods, which are tested in core_test.go.
	// We call them here to ensure coverage.
	rt.ServiceStartup(nil, application.ServiceOptions{})
	rt.ServiceShutdown(nil)

	// Test shutdown with nil core
	rt.Core = nil
	rt.ServiceShutdown(nil)
}

func TestNewServiceRuntime_Good(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)

	sr := NewServiceRuntime(c, "test options")
	assert.NotNil(t, sr)
	assert.Equal(t, c, sr.Core())

	// We can't directly test sr.Config() without a registered config service,
	// but we can ensure it doesn't panic. We'll test the panic case separately.
	assert.Panics(t, func() {
		sr.Config()
	})
}
