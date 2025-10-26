package workspace

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/Snider/Core/pkg/core"
	"github.com/Snider/Core/pkg/crypt/lthn"
	"github.com/Snider/Core/pkg/crypt/openpgp"
	"github.com/Snider/Core/pkg/io"
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
	config          core.Config
	activeWorkspace *Workspace
	workspaceList   map[string]string // Maps Workspace ID to Public Key
	medium          io.Medium
}

// New is the factory function for the core.WithService pattern.
func New(c *core.Core) (any, error) {
	configSvc := core.ServiceFor[core.Config](c, "config")
	// The medium is not a registered service, so we use the local one by default.
	medium := io.Local

	s := &Service{
		Runtime:       core.NewRuntime(c, Options{}),
		config:        configSvc,
		workspaceList: make(map[string]string),
		medium:        medium,
	}

	// Initialize the service after creation.
	if err := s.ServiceStartup(); err != nil {
		return nil, fmt.Errorf("workspace service startup failed: %w", err)
	}

	return s, nil
}

// ServiceStartup initializes the service, loading the workspace list.
func (s *Service) ServiceStartup() error {
	// The config service might not be fully available during initial registration,
	// so we check for nil.
	if s.config == nil {
		return nil
	}

	listPath := filepath.Join(s.config.WorkspacesDir, listFile)

	if s.medium.IsFile(listPath) {
		content, err := s.medium.FileGet(listPath)
		if err != nil {
			return fmt.Errorf("failed to read workspace list: %w", err)
		}
		if err := json.Unmarshal([]byte(content), &s.workspaceList); err != nil {
			fmt.Printf("Warning: could not parse workspace list: %v\n", err)
			s.workspaceList = make(map[string]string)
		}
	}

	return s.SwitchWorkspace(defaultWorkspace)
}

// CreateWorkspace creates a new, obfuscated workspace on the local medium.
func (s *Service) CreateWorkspace(identifier, password string) (string, error) {
	realName := lthn.Hash(identifier)
	workspaceID := lthn.Hash(fmt.Sprintf("workspace/%s", realName))
	workspacePath := filepath.Join(s.config.WorkspacesDir, workspaceID)

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

	listPath := filepath.Join(s.config.WorkspacesDir, listFile)
	if err := s.medium.FileSet(listPath, string(listData)); err != nil {
		return "", fmt.Errorf("failed to write workspace list file: %w", err)
	}

	return workspaceID, nil
}

// SwitchWorkspace changes the active workspace.
func (s *Service) SwitchWorkspace(name string) error {
	if name != defaultWorkspace {
		if _, exists := s.workspaceList[name]; !exists {
			return fmt.Errorf("workspace '%s' does not exist", name)
		}
	}

	path := filepath.Join(s.config.WorkspacesDir, name)
	if err := s.medium.EnsureDir(path); err != nil {
		return fmt.Errorf("failed to ensure workspace directory exists: %w", err)
	}

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
func (s *Service) WorkspaceFileSet(filename, content string) error {
	if s.activeWorkspace == nil {
		return fmt.Errorf("no active workspace")
	}
	path := filepath.Join(s.activeWorkspace.Path, filename)
	return s.medium.FileSet(path, content)
}

// Ensure Service implements the core.Workspace interface.
var _ core.Workspace = (*Service)(nil)
