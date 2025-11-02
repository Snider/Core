package tdd

import (
	"fmt"
	"testing"

	"github.com/Snider/Core/pkg/core"
	"github.com/Snider/Core/pkg/workspace"
	"github.com/stretchr/testify/assert"
)

// MockMedium is a mock implementation of the io.Medium interface.
type MockMedium struct {
	files map[string]string
}

func (m *MockMedium) Read(path string) (string, error) {
	return m.FileGet(path)
}

func (m *MockMedium) Write(path, content string) error {
	return m.FileSet(path, content)
}

func (m *MockMedium) FileGet(path string) (string, error) {
	if content, ok := m.files[path]; ok {
		return content, nil
	}
	return "", fmt.Errorf("file not found: %s", path)
}

func (m *MockMedium) FileSet(path, content string) error {
	m.files[path] = content
	return nil
}

func (m *MockMedium) IsFile(path string) bool {
	_, ok := m.files[path]
	return ok
}

func (m *MockMedium) EnsureDir(path string) error {
	return nil // Assume directories always exist for the mock.
}

// MockConfig is a mock implementation of the core.Config interface.
type MockConfig struct {
	values map[string]any
}

func (m *MockConfig) Get(key string, out any) error {
	if val, ok := m.values[key]; ok {
		// This is a simplified mock. A real implementation would need reflection.
		*(out.(*string)) = val.(string)
		return nil
	}
	return fmt.Errorf("config key not found: %s", key)
}

func (m *MockConfig) Set(key string, v any) error {
	m.values[key] = v
	return nil
}

func setupWorkspaceService(t *testing.T) (*workspace.Service, *MockMedium) {
	s, err := workspace.New()
	assert.NoError(t, err)

	mockMedium := &MockMedium{files: make(map[string]string)}
	mockConfig := &MockConfig{values: make(map[string]any)}
	mockConfig.Set("workspaceDir", "/tmp/workspaces")

	// We need to create a core instance to satisfy the dependency.
	c, err := core.New()
	assert.NoError(t, err)

	s.Runtime = core.NewRuntime(c, workspace.Options{})
	s.SetMedium(mockMedium)
	s.Config = mockConfig

	return s, mockMedium
}

// TestWorkspace_Create_Good tests the happy path for creating a workspace.
func TestWorkspace_Create_Good(t *testing.T) {
	s, _ := setupWorkspaceService(t)

	_, err := s.CreateWorkspace("my-workspace", "password")
	assert.NoError(t, err)
}

// TestWorkspace_Create_Bad tests creating a workspace that already exists.
func TestWorkspace_Create_Bad(t *testing.T) {
	s, _ := setupWorkspaceService(t)

	_, err := s.CreateWorkspace("my-workspace", "password")
	assert.NoError(t, err)

	_, err = s.CreateWorkspace("my-workspace", "password")
	assert.Error(t, err)
}

// TestWorkspace_Switch_Good tests the happy path for switching workspaces.
func TestWorkspace_Switch_Good(t *testing.T) {
	s, _ := setupWorkspaceService(t)

	workspaceID, err := s.CreateWorkspace("my-workspace", "password")
	assert.NoError(t, err)

	err = s.SwitchWorkspace(workspaceID)
	assert.NoError(t, err)
}

// TestWorkspace_Switch_Bad tests switching to a non-existent workspace.
func TestWorkspace_Switch_Bad(t *testing.T) {
	s, _ := setupWorkspaceService(t)

	err := s.SwitchWorkspace("non-existent-workspace")
	assert.Error(t, err)
}

// TestWorkspace_FileGetSet_Good tests the happy path for file operations.
func TestWorkspace_FileGetSet_Good(t *testing.T) {
	s, _ := setupWorkspaceService(t)
	err := s.SwitchWorkspace("default")
	assert.NoError(t, err)

	err = s.WorkspaceFileSet("test.txt", "hello")
	assert.NoError(t, err)

	content, err := s.WorkspaceFileGet("test.txt")
	assert.NoError(t, err)
	assert.Equal(t, "hello", content)
}

// TestWorkspace_FileGet_Bad tests getting a non-existent file.
func TestWorkspace_FileGet_Bad(t *testing.T) {
	s, _ := setupWorkspaceService(t)
	err := s.SwitchWorkspace("default")
	assert.NoError(t, err)

	_, err = s.WorkspaceFileGet("non-existent-file.txt")
	assert.Error(t, err)
}

// TestWorkspace_FileGet_Ugly tests getting a file without an active workspace.
func TestWorkspace_FileGet_Ugly(t *testing.T) {
	s, _ := setupWorkspaceService(t)

	_, err := s.WorkspaceFileGet("test.txt")
	assert.Error(t, err)
}
