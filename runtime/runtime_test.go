package runtime_test

import (
	"testing"

	"github.com/Snider/Core/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/wailsapp/wails/v3/pkg/application"
)

func TestNew(t *testing.T) {
	testCases := []struct {
		name         string
		app          *application.App
		factories    map[string]runtime.ServiceFactory
		expectErr    bool
		expectErrStr string
		checkRuntime func(*testing.T, *runtime.Runtime)
	}{
		{
			name:      "Good path",
			app:       nil,
			factories: map[string]runtime.ServiceFactory{},
			expectErr: false,
			checkRuntime: func(t *testing.T, rt *runtime.Runtime) {
				assert.NotNil(t, rt)
				assert.NotNil(t, rt.Core)
			},
		},
		{
			name:      "With non-nil app",
			app:       &application.App{},
			factories: map[string]runtime.ServiceFactory{},
			expectErr: false,
			checkRuntime: func(t *testing.T, rt *runtime.Runtime) {
				assert.NotNil(t, rt)
				assert.NotNil(t, rt.Core)
				assert.NotNil(t, rt.Core.App)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rt, err := runtime.NewWithFactories(tc.app, tc.factories)

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
