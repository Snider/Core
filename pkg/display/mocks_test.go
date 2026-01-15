package display

import (
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// mockApp is a mock implementation of the App interface for testing.
type mockApp struct {
	windowManager *mockWindowManager
	menuManager   *mockMenuManager
	dialogManager *mockDialogManager
	systemTrayMgr *mockSystemTrayManager
	envManager    *mockEnvManager
	eventManager  *mockEventManager
	logger        *mockLogger
	quitCalled    bool
}

func newMockApp() *mockApp {
	return &mockApp{
		windowManager: newMockWindowManager(),
		menuManager:   newMockMenuManager(),
		dialogManager: newMockDialogManager(),
		systemTrayMgr: newMockSystemTrayManager(),
		envManager:    newMockEnvManager(),
		eventManager:  newMockEventManager(),
		logger:        &mockLogger{},
	}
}

func (m *mockApp) Window() WindowManager         { return m.windowManager }
func (m *mockApp) Menu() MenuManager             { return m.menuManager }
func (m *mockApp) Dialog() DialogManager         { return m.dialogManager }
func (m *mockApp) SystemTray() SystemTrayManager { return m.systemTrayMgr }
func (m *mockApp) Env() EnvManager               { return m.envManager }
func (m *mockApp) Event() EventManager           { return m.eventManager }
func (m *mockApp) Logger() Logger                { return m.logger }
func (m *mockApp) Quit()                         { m.quitCalled = true }

// mockWindowManager tracks window creation calls.
type mockWindowManager struct {
	createdWindows []application.WebviewWindowOptions
	allWindows     []application.Window
}

func newMockWindowManager() *mockWindowManager {
	return &mockWindowManager{
		createdWindows: make([]application.WebviewWindowOptions, 0),
		allWindows:     make([]application.Window, 0),
	}
}

func (m *mockWindowManager) NewWithOptions(opts application.WebviewWindowOptions) *application.WebviewWindow {
	m.createdWindows = append(m.createdWindows, opts)
	// Return nil since we can't create a real window without Wails runtime
	return nil
}

func (m *mockWindowManager) GetAll() []application.Window {
	return m.allWindows
}

// mockMenuManager tracks menu creation calls.
type mockMenuManager struct {
	menusCreated int
	menuSet      *application.Menu
}

func newMockMenuManager() *mockMenuManager {
	return &mockMenuManager{}
}

func (m *mockMenuManager) New() *application.Menu {
	m.menusCreated++
	return nil // Can't create real menu without Wails runtime
}

func (m *mockMenuManager) Set(menu *application.Menu) {
	m.menuSet = menu
}

// mockDialogManager tracks dialog creation calls.
type mockDialogManager struct {
	infoDialogsCreated    int
	warningDialogsCreated int
}

func newMockDialogManager() *mockDialogManager {
	return &mockDialogManager{}
}

func (m *mockDialogManager) Info() *application.MessageDialog {
	m.infoDialogsCreated++
	return nil // Can't create real dialog without Wails runtime
}

func (m *mockDialogManager) Warning() *application.MessageDialog {
	m.warningDialogsCreated++
	return nil // Can't create real dialog without Wails runtime
}

func (m *mockDialogManager) OpenFile() *application.OpenFileDialogStruct {
	return nil // Can't create real dialog without Wails runtime
}

// mockSystemTrayManager tracks system tray creation calls.
type mockSystemTrayManager struct {
	traysCreated int
}

func newMockSystemTrayManager() *mockSystemTrayManager {
	return &mockSystemTrayManager{}
}

func (m *mockSystemTrayManager) New() *application.SystemTray {
	m.traysCreated++
	return nil // Can't create real system tray without Wails runtime
}

// mockEnvManager provides mock environment info.
type mockEnvManager struct {
	envInfo  application.EnvironmentInfo
	darkMode bool
}

func newMockEnvManager() *mockEnvManager {
	return &mockEnvManager{
		envInfo: application.EnvironmentInfo{
			OS:           "test-os",
			Arch:         "test-arch",
			Debug:        true,
			PlatformInfo: map[string]any{"test": "value"},
		},
		darkMode: false,
	}
}

func (m *mockEnvManager) Info() application.EnvironmentInfo {
	return m.envInfo
}

func (m *mockEnvManager) IsDarkMode() bool {
	return m.darkMode
}

// mockEventManager tracks event registration.
type mockEventManager struct {
	registeredEvents []events.ApplicationEventType
}

func newMockEventManager() *mockEventManager {
	return &mockEventManager{
		registeredEvents: make([]events.ApplicationEventType, 0),
	}
}

func (m *mockEventManager) OnApplicationEvent(eventType events.ApplicationEventType, handler func(*application.ApplicationEvent)) func() {
	m.registeredEvents = append(m.registeredEvents, eventType)
	return func() {} // Return a no-op unsubscribe function
}

func (m *mockEventManager) Emit(name string, data ...any) bool {
	return true // Pretend emission succeeded
}

// mockLogger tracks log calls.
type mockLogger struct {
	infoMessages []string
}

func (m *mockLogger) Info(message string, args ...any) {
	m.infoMessages = append(m.infoMessages, message)
}
