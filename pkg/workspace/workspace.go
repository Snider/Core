package workspace

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/Snider/Core/pkg/core"
	"github.com/Snider/Core/pkg/crypt/lthn"
	"github.com/Snider/Core/pkg/crypt/openpgp"
	"github.com/Snider/Core/pkg/io"
	"github.com/wailsapp/wails/v3/pkg/application"
)

const (
	defaultWorkspace = "default"
	listFile         = "list.json"
)

// Options holds configuration for the workspace service.
type Options struct{}

// Workspace represents a user's workspace.
type Workspace struct {
	Name string
	Path string
}

// Service manages user workspaces.
type Service struct {
	*core.Runtime[Options]
	activeWorkspace *Workspace
	workspaceList   map[string]string // Maps Workspace ID to Public Key
	medium          io.Medium
}

// newWorkspaceService contains the common logic for initializing a Service struct.
// It no longer takes config and medium as arguments.
func newWorkspaceService() (*Service, error) {
	s := &Service{
		workspaceList: make(map[string]string),
	}
	return s, nil
}

// New is the constructor for static dependency injection.
// It creates a Service instance without initializing the core.Runtime field.
// Dependencies are passed directly here.
func New(medium io.Medium) (*Service, error) {
	s, err := newWorkspaceService()
	if err != nil {
		return nil, err
	}
	s.medium = medium
	// Initialize the service after creation.
	// Note: ServiceStartup will now get config from s.Runtime.Config()
	if err := s.ServiceStartup(context.Background(), application.ServiceOptions{}); err != nil {
		return nil, fmt.Errorf("workspace service startup failed: %w", err)
	}
	return s, nil
}

// Register is the constructor for dynamic dependency injection (used with core.WithService).
// It creates a Service instance and initializes its core.Runtime field.
// Dependencies are injected during ServiceStartup.
func Register(c *core.Core) (any, error) {
	s, err := newWorkspaceService()
	if err != nil {
		return nil, err
	}
	s.Runtime = core.NewRuntime(c, Options{})
	return s, nil
}

// HandleIPCEvents processes IPC messages, including injecting dependencies on startup.
func (s *Service) HandleIPCEvents(c *core.Core, msg core.Message) error {
	switch m := msg.(type) {
	case map[string]any:
		if action, ok := m["action"].(string); ok && action == "display.open_windowds" {
			return nil //s.handleOpenWindowAction(m)
		}
	case core.ActionServiceStartup:
		return s.ServiceStartup(context.Background(), application.ServiceOptions{})
	default:
		c.App.Logger.Error("Display: Unknown message type", "type", fmt.Sprintf("%T", m))
	}
	return nil
}

// getWorkspaceDir retrieves the WorkspaceDir from the config service.
func (s *Service) getWorkspaceDir() (string, error) {
	var workspaceDir string
	if err := s.Config().Get("workspaceDir", &workspaceDir); err != nil {
		return "", fmt.Errorf("failed to get WorkspaceDir from config: %w", err)
	}
	return workspaceDir, nil
}

// ServiceStartup initializes the service, loading the workspace list.
func (s *Service) ServiceStartup(context.Context, application.ServiceOptions) error {
	var err error
	workspaceDir, err := s.getWorkspaceDir()
	if err != nil {
		return err
	}

	listPath := filepath.Join(workspaceDir, listFile)
	if listPath != "" {
	}
	//if s.medium.IsFile(listPath) {
	//	content, err := s.medium.FileGet(listPath)
	//	if err != nil {
	//		return fmt.Errorf("failed to read workspace list: %w", err)
	//	}
	//	if err := json.Unmarshal([]byte(content), &s.workspaceList); err != nil {
	//		fmt.Printf("Warning: could not parse workspace list: %v\n", err)
	//		s.workspaceList = make(map[string]string)
	//	}
	//}

	return s.SwitchWorkspace(defaultWorkspace)
}

// CreateWorkspace creates a new, obfuscated workspace on the local medium.
func (s *Service) CreateWorkspace(identifier, password string) (string, error) {
	workspaceDir, err := s.getWorkspaceDir()
	if err != nil {
		return "", err
	}

	realName := lthn.Hash(identifier)
	workspaceID := lthn.Hash(fmt.Sprintf("workspace/%s", realName))
	workspacePath := filepath.Join(workspaceDir, workspaceID)

	if _, exists := s.workspaceList[workspaceID]; exists {
		return "", fmt.Errorf("workspace for this identifier already exists")
	}

	dirsToCreate := []string{"config", "log", "data", "files", "keys"}
	for _, dir := range dirsToCreate {
		if err := s.medium.EnsureDir(filepath.Join(workspacePath, dir)); err != nil {
			return "", fmt.Errorf("failed to create workspace directory '%s': %w", dir, err)
		}
	}

	keyPair, err := openpgp.CreateKeyPair(workspaceID, password)
	if err != nil {
		return "", fmt.Errorf("failed to create workspace key pair: %w", err)
	}

	keyFiles := map[string]string{
		filepath.Join(workspacePath, "keys", "key.pub"):  keyPair.PublicKey,
		filepath.Join(workspacePath, "keys", "key.priv"): keyPair.PrivateKey,
	}
	for path, content := range keyFiles {
		if err := s.medium.FileSet(path, content); err != nil {
			return "", fmt.Errorf("failed to write key file %s: %w", path, err)
		}
	}

	s.workspaceList[workspaceID] = keyPair.PublicKey
	listData, err := json.MarshalIndent(s.workspaceList, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal workspace list: %w", err)
	}

	listPath := filepath.Join(workspaceDir, listFile)
	if err := s.medium.FileSet(listPath, string(listData)); err != nil {
		return "", fmt.Errorf("failed to write workspace list file: %w", err)
	}

	return workspaceID, nil
}

// SwitchWorkspace changes the active workspace.
func (s *Service) SwitchWorkspace(name string) error {
	workspaceDir, err := s.getWorkspaceDir()
	if err != nil {
		return err
	}

	if name != defaultWorkspace {
		if _, exists := s.workspaceList[name]; !exists {
			return fmt.Errorf("workspace '%s' does not exist", name)
		}
	}

	path := filepath.Join(workspaceDir, name)
	//if err := s.medium.EnsureDir(path); err != nil {
	//	return fmt.Errorf("failed to ensure workspace directory exists: %w", err)
	//}

	s.activeWorkspace = &Workspace{
		Name: name,
		Path: path,
	}

	return nil
}

// WorkspaceFileGet retrieves a file from the active workspace.
func (s *Service) WorkspaceFileGet(filename string) (string, error) {
	if s.activeWorkspace == nil {
		return "", fmt.Errorf("no active workspace")
	}
	path := filepath.Join(s.activeWorkspace.Path, filename)
	return s.medium.FileGet(path)
}

// WorkspaceFileSet writes a file to the active workspace.
func (s *Service) WorkspaceFileSet(filename, content string) (string, error) {
	if s.activeWorkspace == nil {
		return "", fmt.Errorf("no active workspace")
	}
	path := filepath.Join(s.activeWorkspace.Path, filename)
	return path, nil
	//return s.medium.FileSet(path, content)
}
