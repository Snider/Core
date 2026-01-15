package display

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// buildMenu creates and sets the main application menu. This function is called
// during the startup of the display service.
func (s *Service) buildMenu() {
	appMenu := s.app.Menu().New()
	if runtime.GOOS == "darwin" {
		appMenu.AddRole(application.AppMenu)
	}
	appMenu.AddRole(application.FileMenu)
	appMenu.AddRole(application.ViewMenu)
	appMenu.AddRole(application.EditMenu)

	workspace := appMenu.AddSubmenu("Workspace")
	workspace.Add("New...").OnClick(func(ctx *application.Context) {
		s.handleNewWorkspace()
	})
	workspace.Add("List").OnClick(func(ctx *application.Context) {
		s.handleListWorkspaces()
	})

	// Developer menu for IDE features
	developer := appMenu.AddSubmenu("Developer")
	developer.Add("New File").SetAccelerator("CmdOrCtrl+N").OnClick(func(ctx *application.Context) {
		s.handleNewFile()
	})
	developer.Add("Open File...").SetAccelerator("CmdOrCtrl+O").OnClick(func(ctx *application.Context) {
		s.handleOpenFile()
	})
	developer.Add("Save").SetAccelerator("CmdOrCtrl+S").OnClick(func(ctx *application.Context) {
		s.handleSaveFile()
	})
	developer.AddSeparator()
	developer.Add("Editor").OnClick(func(ctx *application.Context) {
		s.handleOpenEditor()
	})
	developer.Add("Terminal").OnClick(func(ctx *application.Context) {
		s.handleOpenTerminal()
	})
	developer.AddSeparator()
	developer.Add("Run").SetAccelerator("CmdOrCtrl+R").OnClick(func(ctx *application.Context) {
		s.handleRun()
	})
	developer.Add("Build").SetAccelerator("CmdOrCtrl+B").OnClick(func(ctx *application.Context) {
		s.handleBuild()
	})

	appMenu.AddRole(application.WindowMenu)
	appMenu.AddRole(application.HelpMenu)

	s.app.Menu().Set(appMenu)
}

// handleNewWorkspace opens a window for creating a new workspace.
func (s *Service) handleNewWorkspace() {
	// Open a dedicated window for workspace creation
	// The frontend at /workspace/new handles the form
	opts := application.WebviewWindowOptions{
		Name:   "workspace-new",
		Title:  "New Workspace",
		Width:  500,
		Height: 400,
		URL:    "/workspace/new",
	}
	s.app.Window().NewWithOptions(opts)
}

// handleListWorkspaces shows a dialog with available workspaces.
func (s *Service) handleListWorkspaces() {
	// Get workspace service from core
	ws := s.Core().Service("workspace")
	if ws == nil {
		dialog := s.app.Dialog().Warning()
		dialog.SetTitle("Workspace")
		dialog.SetMessage("Workspace service not available")
		dialog.Show()
		return
	}

	// Type assert to access ListWorkspaces method
	lister, ok := ws.(interface{ ListWorkspaces() []string })
	if !ok {
		dialog := s.app.Dialog().Warning()
		dialog.SetTitle("Workspace")
		dialog.SetMessage("Unable to list workspaces")
		dialog.Show()
		return
	}

	workspaces := lister.ListWorkspaces()

	var message string
	if len(workspaces) == 0 {
		message = "No workspaces found.\n\nUse Workspace → New to create one."
	} else {
		message = fmt.Sprintf("Available Workspaces (%d):\n\n%s",
			len(workspaces),
			strings.Join(workspaces, "\n"))
	}

	dialog := s.app.Dialog().Info()
	dialog.SetTitle("Workspaces")
	dialog.SetMessage(message)
	dialog.Show()
}

// handleNewFile opens the editor with a new untitled file.
func (s *Service) handleNewFile() {
	opts := application.WebviewWindowOptions{
		Name:   "editor",
		Title:  "New File - Editor",
		Width:  1200,
		Height: 800,
		URL:    "/#/developer/editor?new=true",
	}
	s.app.Window().NewWithOptions(opts)
}

// handleOpenFile opens a file dialog to select a file, then opens it in the editor.
func (s *Service) handleOpenFile() {
	dialog := s.app.Dialog().OpenFile()
	dialog.SetTitle("Open File")
	dialog.CanChooseFiles(true)
	dialog.CanChooseDirectories(false)
	result, err := dialog.PromptForSingleSelection()
	if err != nil || result == "" {
		return
	}

	opts := application.WebviewWindowOptions{
		Name:   "editor",
		Title:  result + " - Editor",
		Width:  1200,
		Height: 800,
		URL:    "/#/developer/editor?file=" + result,
	}
	s.app.Window().NewWithOptions(opts)
}

// handleSaveFile emits a save event to the focused editor window.
func (s *Service) handleSaveFile() {
	s.app.Event().Emit("ide:save")
}

// handleOpenEditor opens a standalone editor window.
func (s *Service) handleOpenEditor() {
	opts := application.WebviewWindowOptions{
		Name:   "editor",
		Title:  "Editor",
		Width:  1200,
		Height: 800,
		URL:    "/#/developer/editor",
	}
	s.app.Window().NewWithOptions(opts)
}

// handleOpenTerminal opens a terminal window.
func (s *Service) handleOpenTerminal() {
	opts := application.WebviewWindowOptions{
		Name:   "terminal",
		Title:  "Terminal",
		Width:  800,
		Height: 500,
		URL:    "/#/developer/terminal",
	}
	s.app.Window().NewWithOptions(opts)
}

// handleRun emits a run event that the IDE service can handle.
func (s *Service) handleRun() {
	s.app.Event().Emit("ide:run")
}

// handleBuild emits a build event that the IDE service can handle.
func (s *Service) handleBuild() {
	s.app.Event().Emit("ide:build")
}
