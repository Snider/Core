package workspace

import (
	"github.com/Snider/Core/config"
	"github.com/Snider/Core/filesystem"
)

const (
	defaultWorkspace = "default"
	listFile         = "list.json"
)

// Workspace represents a user's workspace.
type Workspace struct {
	Name string
	Path string
}

// Service manages user workspaces.
type Service struct {
	config          *config.Config
	activeWorkspace *Workspace
	workspaceList   map[string]string // Maps Workspace ID to Public Key
	medium          filesystem.Medium
}
