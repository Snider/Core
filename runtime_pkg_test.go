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
