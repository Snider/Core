package display

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// buildMenu creates and sets the main application menu.
func (s *Service) buildMenu() {
	appMenu := s.Core().App.Menu.New()
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

	// Add brand-specific menu items
	//if s.brand == DeveloperHub {
	//	appMenu.AddSubmenu("Developer")
	//}

	appMenu.AddRole(application.WindowMenu)
	appMenu.AddRole(application.HelpMenu)

	s.Core().App.Menu.Set(appMenu)
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
	s.Core().App.Window.NewWithOptions(opts)
}

// handleListWorkspaces shows a dialog with available workspaces.
func (s *Service) handleListWorkspaces() {
	// Get workspace service from core
	ws := s.Core().Service("workspace")
	if ws == nil {
		dialog := s.Core().App.Dialog.Warning()
		dialog.SetTitle("Workspace")
		dialog.SetMessage("Workspace service not available")
		dialog.Show()
		return
	}

	// Type assert to access ListWorkspaces method
	lister, ok := ws.(interface{ ListWorkspaces() []string })
	if !ok {
		dialog := s.Core().App.Dialog.Warning()
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

	dialog := s.Core().App.Dialog.Info()
	dialog.SetTitle("Workspaces")
	dialog.SetMessage(message)
	dialog.Show()
}
