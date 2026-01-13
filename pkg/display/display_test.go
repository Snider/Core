package display

import (
	"testing"

	"github.com/Snider/Core/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// newTestCore creates a new core instance with essential services for testing.
func newTestCore(t *testing.T) *core.Core {
	coreInstance, err := core.New()
	require.NoError(t, err)
	return coreInstance
}

func TestNew(t *testing.T) {
	t.Run("creates service successfully", func(t *testing.T) {
		service, err := New()
		assert.NoError(t, err)
		assert.NotNil(t, service, "New() should return a non-nil service instance")
	})

	t.Run("returns independent instances", func(t *testing.T) {
		service1, err1 := New()
		service2, err2 := New()
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotSame(t, service1, service2, "New() should return different instances")
	})
}

func TestRegister(t *testing.T) {
	t.Run("registers with core successfully", func(t *testing.T) {
		coreInstance := newTestCore(t)
		service, err := Register(coreInstance)
		require.NoError(t, err)
		assert.NotNil(t, service, "Register() should return a non-nil service instance")
	})

	t.Run("returns Service type", func(t *testing.T) {
		coreInstance := newTestCore(t)
		service, err := Register(coreInstance)
		require.NoError(t, err)

		displayService, ok := service.(*Service)
		assert.True(t, ok, "Register() should return *Service type")
		assert.NotNil(t, displayService.Runtime, "Runtime should be initialized")
	})
}

func TestServiceName(t *testing.T) {
	service, err := New()
	require.NoError(t, err)

	name := service.ServiceName()
	assert.Equal(t, "github.com/Snider/Core/display", name)
}

// --- Window Option Tests ---

func TestWindowName(t *testing.T) {
	t.Run("sets window name", func(t *testing.T) {
		opt := WindowName("test-window")
		window := &Window{}

		err := opt(window)
		assert.NoError(t, err)
		assert.Equal(t, "test-window", window.Name)
	})

	t.Run("sets empty name", func(t *testing.T) {
		opt := WindowName("")
		window := &Window{}

		err := opt(window)
		assert.NoError(t, err)
		assert.Equal(t, "", window.Name)
	})
}

func TestWindowTitle(t *testing.T) {
	t.Run("sets window title", func(t *testing.T) {
		opt := WindowTitle("My Application")
		window := &Window{}

		err := opt(window)
		assert.NoError(t, err)
		assert.Equal(t, "My Application", window.Title)
	})

	t.Run("sets title with special characters", func(t *testing.T) {
		opt := WindowTitle("App - v1.0 (Beta)")
		window := &Window{}

		err := opt(window)
		assert.NoError(t, err)
		assert.Equal(t, "App - v1.0 (Beta)", window.Title)
	})
}

func TestWindowURL(t *testing.T) {
	t.Run("sets window URL", func(t *testing.T) {
		opt := WindowURL("/dashboard")
		window := &Window{}

		err := opt(window)
		assert.NoError(t, err)
		assert.Equal(t, "/dashboard", window.URL)
	})

	t.Run("sets full URL", func(t *testing.T) {
		opt := WindowURL("https://example.com/page")
		window := &Window{}

		err := opt(window)
		assert.NoError(t, err)
		assert.Equal(t, "https://example.com/page", window.URL)
	})
}

func TestWindowWidth(t *testing.T) {
	t.Run("sets window width", func(t *testing.T) {
		opt := WindowWidth(1024)
		window := &Window{}

		err := opt(window)
		assert.NoError(t, err)
		assert.Equal(t, 1024, window.Width)
	})

	t.Run("sets zero width", func(t *testing.T) {
		opt := WindowWidth(0)
		window := &Window{}

		err := opt(window)
		assert.NoError(t, err)
		assert.Equal(t, 0, window.Width)
	})

	t.Run("sets large width", func(t *testing.T) {
		opt := WindowWidth(3840)
		window := &Window{}

		err := opt(window)
		assert.NoError(t, err)
		assert.Equal(t, 3840, window.Width)
	})
}

func TestWindowHeight(t *testing.T) {
	t.Run("sets window height", func(t *testing.T) {
		opt := WindowHeight(768)
		window := &Window{}

		err := opt(window)
		assert.NoError(t, err)
		assert.Equal(t, 768, window.Height)
	})

	t.Run("sets zero height", func(t *testing.T) {
		opt := WindowHeight(0)
		window := &Window{}

		err := opt(window)
		assert.NoError(t, err)
		assert.Equal(t, 0, window.Height)
	})
}

func TestApplyOptions(t *testing.T) {
	t.Run("applies no options", func(t *testing.T) {
		window := applyOptions()
		assert.NotNil(t, window)
		assert.Equal(t, "", window.Name)
		assert.Equal(t, "", window.Title)
		assert.Equal(t, 0, window.Width)
		assert.Equal(t, 0, window.Height)
	})

	t.Run("applies single option", func(t *testing.T) {
		window := applyOptions(WindowTitle("Test"))
		assert.NotNil(t, window)
		assert.Equal(t, "Test", window.Title)
	})

	t.Run("applies multiple options", func(t *testing.T) {
		window := applyOptions(
			WindowName("main"),
			WindowTitle("My App"),
			WindowURL("/home"),
			WindowWidth(1280),
			WindowHeight(720),
		)

		assert.NotNil(t, window)
		assert.Equal(t, "main", window.Name)
		assert.Equal(t, "My App", window.Title)
		assert.Equal(t, "/home", window.URL)
		assert.Equal(t, 1280, window.Width)
		assert.Equal(t, 720, window.Height)
	})

	t.Run("handles nil options slice", func(t *testing.T) {
		window := applyOptions(nil...)
		assert.NotNil(t, window)
	})

	t.Run("applies options in order", func(t *testing.T) {
		// Later options should override earlier ones
		window := applyOptions(
			WindowTitle("First"),
			WindowTitle("Second"),
		)

		assert.NotNil(t, window)
		assert.Equal(t, "Second", window.Title)
	})
}

// --- ActionOpenWindow Tests ---

func TestActionOpenWindow(t *testing.T) {
	t.Run("creates action with options", func(t *testing.T) {
		action := ActionOpenWindow{
			WebviewWindowOptions: application.WebviewWindowOptions{
				Name:   "test",
				Title:  "Test Window",
				Width:  800,
				Height: 600,
			},
		}

		assert.Equal(t, "test", action.Name)
		assert.Equal(t, "Test Window", action.Title)
		assert.Equal(t, 800, action.Width)
		assert.Equal(t, 600, action.Height)
	})
}

// --- Integration Tests (require Wails runtime) ---

func TestOpenWindow(t *testing.T) {
	t.Run("requires Wails runtime", func(t *testing.T) {
		t.Skip("Skipping OpenWindow test - requires running Wails application instance")
	})
}

func TestNewWithStruct(t *testing.T) {
	t.Run("requires Wails runtime", func(t *testing.T) {
		t.Skip("Skipping NewWithStruct test - requires running Wails application instance")
	})
}

func TestNewWithOptions(t *testing.T) {
	t.Run("requires Wails runtime", func(t *testing.T) {
		t.Skip("Skipping NewWithOptions test - requires running Wails application instance")
	})
}

func TestNewWithURL(t *testing.T) {
	t.Run("requires Wails runtime", func(t *testing.T) {
		t.Skip("Skipping NewWithURL test - requires running Wails application instance")
	})
}

func TestServiceStartup(t *testing.T) {
	t.Run("requires Wails runtime", func(t *testing.T) {
		t.Skip("Skipping ServiceStartup test - requires running Wails application instance")
	})
}

func TestSelectDirectory(t *testing.T) {
	t.Run("requires Wails runtime", func(t *testing.T) {
		t.Skip("Skipping SelectDirectory test - requires running Wails application instance")
	})
}

func TestShowEnvironmentDialog(t *testing.T) {
	t.Run("requires Wails runtime", func(t *testing.T) {
		t.Skip("Skipping ShowEnvironmentDialog test - requires running Wails application instance")
	})
}

func TestHandleIPCEvents(t *testing.T) {
	t.Run("requires Wails runtime for full test", func(t *testing.T) {
		t.Skip("Skipping HandleIPCEvents test - requires running Wails application instance")
	})
}

func TestBuildMenu(t *testing.T) {
	t.Run("requires Wails runtime", func(t *testing.T) {
		t.Skip("Skipping buildMenu test - requires running Wails application instance")
	})
}

func TestSystemTray(t *testing.T) {
	t.Run("requires Wails runtime", func(t *testing.T) {
		t.Skip("Skipping systemTray test - requires running Wails application instance")
	})
}
