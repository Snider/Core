package workspace

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"sync"

	core "github.com/host-uk/core/pkg/framework/core"
	"github.com/host-uk/core/pkg/io"
)

// Service implements the core.Workspace interface.
type Service struct {
	core            *core.Core
	activeWorkspace string
	rootPath        string
	medium          io.Medium
	mu              sync.RWMutex
}

// New creates a new Workspace service instance.
func New(c *core.Core) (any, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, core.E("workspace.New", "failed to determine home directory", err)
	}
	rootPath := filepath.Join(home, ".core", "workspaces")

	s := &Service{
		core:     c,
		rootPath: rootPath,
		medium:   io.Local,
	}

	if err := s.medium.EnsureDir(rootPath); err != nil {
		return nil, core.E("workspace.New", "failed to ensure root directory", err)
	}

	return s, nil
}

// CreateWorkspace creates a new encrypted workspace.
// Identifier is hashed (SHA-256 as proxy for LTHN) to create the directory name.
// A PGP keypair is generated using the password.
func (s *Service) CreateWorkspace(identifier, password string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Identification (LTHN hash proxy)
	hash := sha256.Sum256([]byte(identifier))
	wsID := hex.EncodeToString(hash[:])
	wsPath := filepath.Join(s.rootPath, wsID)

	if s.medium.Exists(wsPath) {
		return "", core.E("workspace.CreateWorkspace", "workspace already exists", nil)
	}

	// 2. Directory structure
	dirs := []string{"config", "log", "data", "files", "keys"}
	for _, d := range dirs {
		if err := s.medium.EnsureDir(filepath.Join(wsPath, d)); err != nil {
			return "", core.E("workspace.CreateWorkspace", "failed to create directory: "+d, err)
		}
	}

	// 3. PGP Keypair generation
	crypt, err := s.core.Crypt()
	if err != nil {
		return "", core.E("workspace.CreateWorkspace", "failed to retrieve crypt service", err)
	}
	privKey, err := crypt.CreateKeyPair(identifier, password)
	if err != nil {
		return "", core.E("workspace.CreateWorkspace", "failed to generate keys", err)
	}

	// Save private key
	if err := s.medium.Write(filepath.Join(wsPath, "keys", "private.key"), privKey); err != nil {
		return "", core.E("workspace.CreateWorkspace", "failed to save private key", err)
	}

	return wsID, nil
}

// SwitchWorkspace changes the active workspace.
func (s *Service) SwitchWorkspace(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	wsPath := filepath.Join(s.rootPath, name)
	if !s.medium.IsDir(wsPath) {
		return core.E("workspace.SwitchWorkspace", "workspace not found: "+name, nil)
	}

	s.activeWorkspace = name
	return nil
}

// WorkspaceFileGet retrieves the content of a file from the active workspace.
// In a full implementation, this would involve decryption using the workspace key.
func (s *Service) WorkspaceFileGet(filename string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.activeWorkspace == "" {
		return "", core.E("workspace.WorkspaceFileGet", "no active workspace", nil)
	}

	path := filepath.Join(s.rootPath, s.activeWorkspace, "files", filename)
	return s.medium.Read(path)
}

// WorkspaceFileSet saves content to a file in the active workspace.
// In a full implementation, this would involve encryption using the workspace key.
func (s *Service) WorkspaceFileSet(filename, content string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.activeWorkspace == "" {
		return core.E("workspace.WorkspaceFileSet", "no active workspace", nil)
	}

	path := filepath.Join(s.rootPath, s.activeWorkspace, "files", filename)
	return s.medium.Write(path, content)
}

// HandleIPCEvents handles workspace-related IPC messages.
func (s *Service) HandleIPCEvents(c *core.Core, msg core.Message) error {
	switch m := msg.(type) {
	case map[string]any:
		action, _ := m["action"].(string)
		switch action {
		case "workspace.create":
			id, _ := m["identifier"].(string)
			pass, _ := m["password"].(string)
			_, err := s.CreateWorkspace(id, pass)
			return err
		case "workspace.switch":
			name, _ := m["name"].(string)
			return s.SwitchWorkspace(name)
		}
	}
	return nil
}

// Ensure Service implements core.Workspace.
var _ core.Workspace = (*Service)(nil)
