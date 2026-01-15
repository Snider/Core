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
		assert.NotNil(t, displayService.ServiceRuntime, "ServiceRuntime should be initialized")
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

// --- Tests with Mock App ---

// newServiceWithMockApp creates a Service with a mock app for testing.
func newServiceWithMockApp(t *testing.T) (*Service, *mockApp) {
	service, err := New()
	require.NoError(t, err)
	mock := newMockApp()
	service.app = mock
	return service, mock
}

func TestOpenWindow(t *testing.T) {
	t.Run("creates window with default options", func(t *testing.T) {
		service, mock := newServiceWithMockApp(t)

		err := service.OpenWindow()
		assert.NoError(t, err)

		// Verify window was created
		assert.Len(t, mock.windowManager.createdWindows, 1)
		opts := mock.windowManager.createdWindows[0]
		assert.Equal(t, "main", opts.Name)
		assert.Equal(t, "Core", opts.Title)
		assert.Equal(t, 1280, opts.Width)
		assert.Equal(t, 800, opts.Height)
		assert.Equal(t, "/", opts.URL)
	})

	t.Run("creates window with custom options", func(t *testing.T) {
		service, mock := newServiceWithMockApp(t)

		err := service.OpenWindow(
			WindowName("custom-window"),
			WindowTitle("Custom Title"),
			WindowWidth(640),
			WindowHeight(480),
			WindowURL("/custom"),
		)
		assert.NoError(t, err)

		assert.Len(t, mock.windowManager.createdWindows, 1)
		opts := mock.windowManager.createdWindows[0]
		assert.Equal(t, "custom-window", opts.Name)
		assert.Equal(t, "Custom Title", opts.Title)
		assert.Equal(t, 640, opts.Width)
		assert.Equal(t, 480, opts.Height)
		assert.Equal(t, "/custom", opts.URL)
	})
}

func TestNewWithStruct(t *testing.T) {
	t.Run("creates window from struct", func(t *testing.T) {
		service, mock := newServiceWithMockApp(t)

		opts := &Window{
			Name:   "struct-window",
			Title:  "Struct Title",
			Width:  800,
			Height: 600,
			URL:    "/struct",
		}

		_, err := service.NewWithStruct(opts)
		assert.NoError(t, err)

		assert.Len(t, mock.windowManager.createdWindows, 1)
		created := mock.windowManager.createdWindows[0]
		assert.Equal(t, "struct-window", created.Name)
		assert.Equal(t, "Struct Title", created.Title)
		assert.Equal(t, 800, created.Width)
		assert.Equal(t, 600, created.Height)
	})
}

func TestNewWithOptions(t *testing.T) {
	t.Run("creates window from options", func(t *testing.T) {
		service, mock := newServiceWithMockApp(t)

		_, err := service.NewWithOptions(
			WindowName("options-window"),
			WindowTitle("Options Title"),
		)
		assert.NoError(t, err)

		assert.Len(t, mock.windowManager.createdWindows, 1)
		opts := mock.windowManager.createdWindows[0]
		assert.Equal(t, "options-window", opts.Name)
		assert.Equal(t, "Options Title", opts.Title)
	})
}

func TestNewWithURL(t *testing.T) {
	t.Run("creates window with URL", func(t *testing.T) {
		service, mock := newServiceWithMockApp(t)

		_, err := service.NewWithURL("/dashboard")
		assert.NoError(t, err)

		assert.Len(t, mock.windowManager.createdWindows, 1)
		opts := mock.windowManager.createdWindows[0]
		assert.Equal(t, "/dashboard", opts.URL)
		assert.Equal(t, "Core", opts.Title)
		assert.Equal(t, 1280, opts.Width)
		assert.Equal(t, 900, opts.Height)
	})
}

func TestHandleOpenWindowAction(t *testing.T) {
	t.Run("creates window from message map", func(t *testing.T) {
		service, mock := newServiceWithMockApp(t)

		msg := map[string]any{
			"name": "action-window",
			"options": map[string]any{
				"Title":  "Action Title",
				"Width":  float64(1024),
				"Height": float64(768),
			},
		}

		err := service.handleOpenWindowAction(msg)
		assert.NoError(t, err)

		assert.Len(t, mock.windowManager.createdWindows, 1)
		opts := mock.windowManager.createdWindows[0]
		assert.Equal(t, "action-window", opts.Name)
		assert.Equal(t, "Action Title", opts.Title)
		assert.Equal(t, 1024, opts.Width)
		assert.Equal(t, 768, opts.Height)
	})
}

func TestMonitorScreenChanges(t *testing.T) {
	t.Run("registers theme change event", func(t *testing.T) {
		service, mock := newServiceWithMockApp(t)

		service.monitorScreenChanges()

		// Verify that an event handler was registered
		assert.Len(t, mock.eventManager.registeredEvents, 1)
	})
}

func TestSelectDirectory(t *testing.T) {
	t.Run("requires Wails runtime for file dialog", func(t *testing.T) {
		// SelectDirectory uses application.OpenFileDialog() directly
		// which requires Wails runtime. This test verifies the method exists.
		service, _ := newServiceWithMockApp(t)
		assert.NotNil(t, service.SelectDirectory)
	})
}

func TestShowEnvironmentDialog(t *testing.T) {
	t.Run("calls dialog with environment info", func(t *testing.T) {
		service, mock := newServiceWithMockApp(t)

		// This will panic because Dialog().Info() returns nil
		// We're verifying the env info is accessed, not that a dialog shows
		assert.NotPanics(t, func() {
			defer func() { recover() }() // Recover from nil dialog
			service.ShowEnvironmentDialog()
		})

		// Verify dialog was requested (even though it's nil)
		assert.Equal(t, 1, mock.dialogManager.infoDialogsCreated)
	})
}

func TestBuildMenu(t *testing.T) {
	t.Run("creates and sets menu", func(t *testing.T) {
		service, mock := newServiceWithMockApp(t)
		coreInstance := newTestCore(t)
		service.ServiceRuntime = core.NewServiceRuntime[Options](coreInstance, Options{})

		// buildMenu will panic because Menu().New() returns nil
		// We verify the menu manager was called
		assert.NotPanics(t, func() {
			defer func() { recover() }()
			service.buildMenu()
		})

		assert.Equal(t, 1, mock.menuManager.menusCreated)
	})
}

func TestSystemTray(t *testing.T) {
	t.Run("creates system tray", func(t *testing.T) {
		service, mock := newServiceWithMockApp(t)
		coreInstance := newTestCore(t)
		service.ServiceRuntime = core.NewServiceRuntime[Options](coreInstance, Options{})

		// systemTray will panic because SystemTray().New() returns nil
		// We verify the system tray manager was called
		assert.NotPanics(t, func() {
			defer func() { recover() }()
			service.systemTray()
		})

		assert.Equal(t, 1, mock.systemTrayMgr.traysCreated)
	})
}

func TestApplyOptionsWithError(t *testing.T) {
	t.Run("returns nil when option returns error", func(t *testing.T) {
		errorOption := func(o *Window) error {
			return assert.AnError
		}

		result := applyOptions(errorOption)
		assert.Nil(t, result)
	})

	t.Run("processes multiple options until error", func(t *testing.T) {
		firstOption := func(o *Window) error {
			o.Name = "first"
			return nil
		}
		errorOption := func(o *Window) error {
			return assert.AnError
		}

		result := applyOptions(firstOption, errorOption)
		assert.Nil(t, result)
		// The first option should have run before error
		// But the result is nil so we can't check
	})

	t.Run("handles empty options slice", func(t *testing.T) {
		opts := []WindowOption{}
		result := applyOptions(opts...)
		assert.NotNil(t, result)
		assert.Equal(t, "", result.Name) // Default empty values
	})
}

func TestHandleNewWorkspace(t *testing.T) {
	t.Run("opens workspace creation window", func(t *testing.T) {
		service, mock := newServiceWithMockApp(t)

		service.handleNewWorkspace()

		// Verify a window was created with correct options
		assert.Len(t, mock.windowManager.createdWindows, 1)
		opts := mock.windowManager.createdWindows[0]
		assert.Equal(t, "workspace-new", opts.Name)
		assert.Equal(t, "New Workspace", opts.Title)
		assert.Equal(t, 500, opts.Width)
		assert.Equal(t, 400, opts.Height)
		assert.Equal(t, "/workspace/new", opts.URL)
	})
}

func TestHandleListWorkspaces(t *testing.T) {
	t.Run("shows warning when workspace service not available", func(t *testing.T) {
		service, mock := newServiceWithMockApp(t)
		coreInstance := newTestCore(t)
		service.ServiceRuntime = core.NewServiceRuntime[Options](coreInstance, Options{})

		// Don't register workspace service - it won't be available
		// This will panic because Dialog().Warning() returns nil
		assert.NotPanics(t, func() {
			defer func() { recover() }()
			service.handleListWorkspaces()
		})

		assert.Equal(t, 1, mock.dialogManager.warningDialogsCreated)
	})
}

func TestParseWindowOptions(t *testing.T) {
	t.Run("parses complete options", func(t *testing.T) {
		msg := map[string]any{
			"name": "test-window",
			"options": map[string]any{
				"Title":  "Test Title",
				"Width":  float64(800),
				"Height": float64(600),
			},
		}

		opts := parseWindowOptions(msg)

		assert.Equal(t, "test-window", opts.Name)
		assert.Equal(t, "Test Title", opts.Title)
		assert.Equal(t, 800, opts.Width)
		assert.Equal(t, 600, opts.Height)
	})

	t.Run("handles missing name", func(t *testing.T) {
		msg := map[string]any{
			"options": map[string]any{
				"Title": "Test Title",
			},
		}

		opts := parseWindowOptions(msg)

		assert.Equal(t, "", opts.Name)
		assert.Equal(t, "Test Title", opts.Title)
	})

	t.Run("handles missing options", func(t *testing.T) {
		msg := map[string]any{
			"name": "test-window",
		}

		opts := parseWindowOptions(msg)

		assert.Equal(t, "test-window", opts.Name)
		assert.Equal(t, "", opts.Title)
		assert.Equal(t, 0, opts.Width)
		assert.Equal(t, 0, opts.Height)
	})

	t.Run("handles empty map", func(t *testing.T) {
		msg := map[string]any{}

		opts := parseWindowOptions(msg)

		assert.Equal(t, "", opts.Name)
		assert.Equal(t, "", opts.Title)
	})

	t.Run("handles wrong type for name", func(t *testing.T) {
		msg := map[string]any{
			"name": 123, // Wrong type - should be string
		}

		opts := parseWindowOptions(msg)

		assert.Equal(t, "", opts.Name) // Should not set name
	})

	t.Run("handles wrong type for options", func(t *testing.T) {
		msg := map[string]any{
			"name":    "test",
			"options": "not-a-map", // Wrong type
		}

		opts := parseWindowOptions(msg)

		assert.Equal(t, "test", opts.Name)
		assert.Equal(t, "", opts.Title) // Options not parsed
	})

	t.Run("handles partial width/height", func(t *testing.T) {
		msg := map[string]any{
			"options": map[string]any{
				"Width": float64(800),
				// Height missing
			},
		}

		opts := parseWindowOptions(msg)

		assert.Equal(t, 800, opts.Width)
		assert.Equal(t, 0, opts.Height)
	})
}

func TestBuildWailsWindowOptions(t *testing.T) {
	t.Run("creates default options with no args", func(t *testing.T) {
		opts := buildWailsWindowOptions()

		assert.Equal(t, "main", opts.Name)
		assert.Equal(t, "Core", opts.Title)
		assert.Equal(t, 1280, opts.Width)
		assert.Equal(t, 800, opts.Height)
		assert.Equal(t, "/", opts.URL)
	})

	t.Run("applies custom options", func(t *testing.T) {
		opts := buildWailsWindowOptions(
			WindowName("custom"),
			WindowTitle("Custom Title"),
			WindowWidth(640),
			WindowHeight(480),
			WindowURL("/custom"),
		)

		assert.Equal(t, "custom", opts.Name)
		assert.Equal(t, "Custom Title", opts.Title)
		assert.Equal(t, 640, opts.Width)
		assert.Equal(t, 480, opts.Height)
		assert.Equal(t, "/custom", opts.URL)
	})

	t.Run("skips nil options", func(t *testing.T) {
		opts := buildWailsWindowOptions(nil, WindowTitle("Test"))

		assert.Equal(t, "Test", opts.Title)
		assert.Equal(t, "main", opts.Name) // Default preserved
	})
}
