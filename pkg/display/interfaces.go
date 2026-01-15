package display

import (
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// App abstracts the Wails application API for testing.
type App interface {
	Window() WindowManager
	Menu() MenuManager
	Dialog() DialogManager
	SystemTray() SystemTrayManager
	Env() EnvManager
	Event() EventManager
	Logger() Logger
	Quit()
}

// WindowManager handles window creation and management.
type WindowManager interface {
	NewWithOptions(opts application.WebviewWindowOptions) *application.WebviewWindow
	GetAll() []application.Window
}

// MenuManager handles menu creation.
type MenuManager interface {
	New() *application.Menu
	Set(menu *application.Menu)
}

// DialogManager handles dialog creation.
type DialogManager interface {
	Info() *application.MessageDialog
	Warning() *application.MessageDialog
	OpenFile() *application.OpenFileDialogStruct
}

// SystemTrayManager handles system tray creation.
type SystemTrayManager interface {
	New() *application.SystemTray
}

// EnvManager provides environment information.
type EnvManager interface {
	Info() application.EnvironmentInfo
	IsDarkMode() bool
}

// EventManager handles event registration and emission.
type EventManager interface {
	OnApplicationEvent(eventType events.ApplicationEventType, handler func(*application.ApplicationEvent)) func()
	Emit(name string, data ...any) bool
}

// Logger provides logging capabilities.
type Logger interface {
	Info(message string, args ...any)
}

// wailsApp wraps a real Wails application to implement the App interface.
type wailsApp struct {
	app *application.App
}

func newWailsApp(app *application.App) App {
	return &wailsApp{app: app}
}

func (w *wailsApp) Window() WindowManager         { return &wailsWindowManager{app: w.app} }
func (w *wailsApp) Menu() MenuManager             { return &wailsMenuManager{app: w.app} }
func (w *wailsApp) Dialog() DialogManager         { return &wailsDialogManager{app: w.app} }
func (w *wailsApp) SystemTray() SystemTrayManager { return &wailsSystemTrayManager{app: w.app} }
func (w *wailsApp) Env() EnvManager               { return &wailsEnvManager{app: w.app} }
func (w *wailsApp) Event() EventManager           { return &wailsEventManager{app: w.app} }
func (w *wailsApp) Logger() Logger                { return w.app.Logger }
func (w *wailsApp) Quit()                         { w.app.Quit() }

// Wails adapter implementations

type wailsWindowManager struct{ app *application.App }

func (m *wailsWindowManager) NewWithOptions(opts application.WebviewWindowOptions) *application.WebviewWindow {
	return m.app.Window.NewWithOptions(opts)
}
func (m *wailsWindowManager) GetAll() []application.Window {
	return m.app.Window.GetAll()
}

type wailsMenuManager struct{ app *application.App }

func (m *wailsMenuManager) New() *application.Menu     { return m.app.Menu.New() }
func (m *wailsMenuManager) Set(menu *application.Menu) { m.app.Menu.Set(menu) }

type wailsDialogManager struct{ app *application.App }

func (m *wailsDialogManager) Info() *application.MessageDialog    { return m.app.Dialog.Info() }
func (m *wailsDialogManager) Warning() *application.MessageDialog { return m.app.Dialog.Warning() }
func (m *wailsDialogManager) OpenFile() *application.OpenFileDialogStruct {
	return m.app.Dialog.OpenFile()
}

type wailsSystemTrayManager struct{ app *application.App }

func (m *wailsSystemTrayManager) New() *application.SystemTray { return m.app.SystemTray.New() }

type wailsEnvManager struct{ app *application.App }

func (m *wailsEnvManager) Info() application.EnvironmentInfo { return m.app.Env.Info() }
func (m *wailsEnvManager) IsDarkMode() bool                  { return m.app.Env.IsDarkMode() }

type wailsEventManager struct{ app *application.App }

func (m *wailsEventManager) OnApplicationEvent(eventType events.ApplicationEventType, handler func(*application.ApplicationEvent)) func() {
	return m.app.Event.OnApplicationEvent(eventType, handler)
}
func (m *wailsEventManager) Emit(name string, data ...any) bool {
	return m.app.Event.Emit(name, data...)
}
