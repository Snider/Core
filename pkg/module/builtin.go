package module

// BuiltinIDE returns the IDE module configuration.
func BuiltinIDE() Config {
	return Config{
		Code:        "ide",
		Type:        TypeCore,
		Name:        "Core IDE",
		Version:     "0.1.0",
		Namespace:   "core",
		Description: "Integrated development environment for Core",
		Contexts:    []Context{ContextDeveloper, ContextDefault},
		Menu: []MenuItem{
			{
				ID:       "developer",
				Label:    "Developer",
				Order:    100,
				Contexts: []Context{ContextDeveloper, ContextDefault},
				Children: []MenuItem{
					{ID: "dev-new-file", Label: "New File", Accelerator: "CmdOrCtrl+N", Action: "ide:new-file", Order: 1},
					{ID: "dev-open-file", Label: "Open File...", Accelerator: "CmdOrCtrl+O", Action: "ide:open-file", Order: 2},
					{ID: "dev-save", Label: "Save", Accelerator: "CmdOrCtrl+S", Action: "ide:save", Order: 3},
					{ID: "dev-sep1", Separator: true, Order: 4},
					{ID: "dev-editor", Label: "Editor", Route: "/developer/editor", Order: 5},
					{ID: "dev-terminal", Label: "Terminal", Route: "/developer/terminal", Order: 6},
					{ID: "dev-sep2", Separator: true, Order: 7},
					{ID: "dev-run", Label: "Run", Accelerator: "CmdOrCtrl+R", Action: "ide:run", Order: 8},
					{ID: "dev-build", Label: "Build", Accelerator: "CmdOrCtrl+B", Action: "ide:build", Order: 9},
				},
			},
		},
		Routes: []Route{
			{Path: "/developer/editor", Component: "dev-edit", Title: "Editor", Contexts: []Context{ContextDeveloper, ContextDefault}},
			{Path: "/developer/terminal", Component: "dev-terminal", Title: "Terminal", Contexts: []Context{ContextDeveloper, ContextDefault}},
		},
		API: []APIEndpoint{
			{Method: "POST", Path: "/file/new", Description: "Create a new file"},
			{Method: "POST", Path: "/file/open", Description: "Open a file"},
			{Method: "POST", Path: "/file/save", Description: "Save a file"},
			{Method: "GET", Path: "/file/read", Description: "Read file content"},
			{Method: "GET", Path: "/dir/list", Description: "List directory contents"},
			{Method: "GET", Path: "/languages", Description: "Get supported languages"},
		},
	}
}

// BuiltinWorkspace returns the workspace module configuration.
func BuiltinWorkspace() Config {
	return Config{
		Code:        "workspace",
		Type:        TypeCore,
		Name:        "Workspace Manager",
		Version:     "0.1.0",
		Namespace:   "core",
		Description: "Project workspace management",
		Contexts:    []Context{ContextDeveloper, ContextDefault},
		Menu: []MenuItem{
			{
				ID:       "workspace",
				Label:    "Workspace",
				Order:    50,
				Contexts: []Context{ContextDeveloper, ContextDefault},
				Children: []MenuItem{
					{ID: "ws-new", Label: "New...", Action: "workspace:new", Order: 1},
					{ID: "ws-open", Label: "Open...", Action: "workspace:open", Order: 2},
					{ID: "ws-list", Label: "List", Action: "workspace:list", Order: 3},
				},
			},
		},
		Routes: []Route{
			{Path: "/workspace/new", Component: "workspace-new", Title: "New Workspace"},
			{Path: "/workspace/list", Component: "workspace-list", Title: "Workspaces"},
		},
	}
}

// BuiltinSystem returns the system module configuration.
func BuiltinSystem() Config {
	return Config{
		Code:        "system",
		Type:        TypeCore,
		Name:        "System",
		Version:     "0.1.0",
		Namespace:   "core",
		Description: "System information and health",
		Contexts:    []Context{ContextDefault},
		API: []APIEndpoint{
			{Method: "GET", Path: "/info", Description: "System information"},
			{Method: "GET", Path: "/health", Description: "Health check"},
			{Method: "GET", Path: "/runtime", Description: "Runtime information"},
		},
	}
}

// BuiltinNavigation returns sidebar navigation items.
func BuiltinNavigation() Config {
	return Config{
		Code:        "nav",
		Type:        TypeCore,
		Name:        "Navigation",
		Version:     "0.1.0",
		Namespace:   "core",
		Description: "Core navigation items",
		Contexts:    []Context{ContextDefault, ContextDeveloper, ContextMiner, ContextRetail},
		Menu: []MenuItem{
			{ID: "nav-dashboard", Label: "Dashboard", Route: "blockchain", Icon: "fa-house fa-regular fa-2xl shrink-0", Order: 1, Contexts: []Context{ContextDefault, ContextDeveloper, ContextMiner}},
			{ID: "nav-mining", Label: "Mining", Route: "mining", Icon: "fa-microchip fa-regular fa-2xl shrink-0", Order: 20, Contexts: []Context{ContextDefault, ContextDeveloper, ContextMiner}},
			{ID: "nav-developer", Label: "Developer", Route: "dev/edit", Icon: "fa-code fa-regular fa-2xl shrink-0", Order: 30, Contexts: []Context{ContextDefault, ContextDeveloper}},
		},
	}
}

// RegisterBuiltins registers all built-in Core modules.
func RegisterBuiltins(r *Registry) error {
	builtins := []Config{
		BuiltinNavigation(),
		BuiltinIDE(),
		BuiltinWorkspace(),
		BuiltinSystem(),
	}
	for _, cfg := range builtins {
		if err := r.Register(cfg); err != nil {
			return err
		}
	}
	return nil
}
