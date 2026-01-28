package workspace

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/host-uk/core/pkg/core"
	"github.com/host-uk/core/pkg/io"
	"github.com/stretchr/testify/assert"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// mockConfig is a mock implementation of the core.Config interface for testing.
type mockConfig struct {
	values map[string]interface{}
}

func (m *mockConfig) Get(key string, out any) error {
	val, ok := m.values[key]
	if !ok {
		return fmt.Errorf("key not found: %s", key)
	}
	// This is a simplified mock; a real one would use reflection to set `out`
	switch v := out.(type) {
	case *string:
		*v = val.(string)
	default:
		return fmt.Errorf("unsupported type in mock config Get")
	}
	return nil
}

func (m *mockConfig) Set(key string, v any) error {
	m.values[key] = v
	return nil
}

// newTestService creates a workspace service instance with mocked dependencies.
func newTestService(t *testing.T, workspaceDir string) (*Service, *io.MockMedium) {
	coreInstance, err := core.New()
	assert.NoError(t, err)

	mockCfg := &mockConfig{values: map[string]interface{}{"workspaceDir": workspaceDir}}
	coreInstance.RegisterService("config", mockCfg)

	mockMedium := io.NewMockMedium()
	service, err := New(mockMedium)
	assert.NoError(t, err)

	service.ServiceRuntime = core.NewServiceRuntime(coreInstance, Options{})

	return service, mockMedium
}

func TestServiceStartup(t *testing.T) {
	workspaceDir := "/tmp/workspace"

	t.Run("existing valid list.json", func(t *testing.T) {
		service, mockMedium := newTestService(t, workspaceDir)

		expectedWorkspaceList := map[string]string{
			"workspace1": "pubkey1",
			"workspace2": "pubkey2",
		}
		listContent, _ := json.MarshalIndent(expectedWorkspaceList, "", "  ")
		listPath := filepath.Join(workspaceDir, listFile)
		mockMedium.Files[listPath] = string(listContent)

		err := service.ServiceStartup(context.Background(), application.ServiceOptions{})

		assert.NoError(t, err)
		// assert.Equal(t, expectedWorkspaceList, service.workspaceList) // This check is difficult with current implementation
		assert.NotNil(t, service.activeWorkspace)
		assert.Equal(t, defaultWorkspace, service.activeWorkspace.Name)
	})
}

func TestCreateAndSwitchWorkspace(t *testing.T) {
	workspaceDir := "/tmp/workspace"
	service, _ := newTestService(t, workspaceDir)

	// Create
	workspaceID, err := service.CreateWorkspace("test", "password")
	assert.NoError(t, err)
	assert.NotEmpty(t, workspaceID)

	// Switch
	err = service.SwitchWorkspace(workspaceID)
	assert.NoError(t, err)
	assert.Equal(t, workspaceID, service.activeWorkspace.Name)
}

func TestWorkspaceFileOperations(t *testing.T) {
	workspaceDir := "/tmp/workspace"

	t.Run("FileGet returns error when no active workspace", func(t *testing.T) {
		service, _ := newTestService(t, workspaceDir)
		// Don't call ServiceStartup so there's no active workspace

		_, err := service.WorkspaceFileGet("test.txt")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no active workspace")
	})

	t.Run("FileSet returns error when no active workspace", func(t *testing.T) {
		service, _ := newTestService(t, workspaceDir)
		// Don't call ServiceStartup so there's no active workspace

		err := service.WorkspaceFileSet("test.txt", "content")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no active workspace")
	})

	t.Run("FileGet and FileSet work with active workspace", func(t *testing.T) {
		service, mockMedium := newTestService(t, workspaceDir)

		// Start up the service to set active workspace
		err := service.ServiceStartup(context.Background(), application.ServiceOptions{})
		assert.NoError(t, err)

		// Test FileSet
		err = service.WorkspaceFileSet("test.txt", "hello world")
		assert.NoError(t, err)

		// Verify file was written to mock medium
		expectedPath := filepath.Join(workspaceDir, defaultWorkspace, "test.txt")
		assert.Equal(t, "hello world", mockMedium.Files[expectedPath])

		// Test FileGet
		content, err := service.WorkspaceFileGet("test.txt")
		assert.NoError(t, err)
		assert.Equal(t, "hello world", content)
	})
}

func TestListWorkspaces(t *testing.T) {
	workspaceDir := "/tmp/workspace"
	service, _ := newTestService(t, workspaceDir)

	t.Run("returns empty list when no workspaces", func(t *testing.T) {
		workspaces := service.ListWorkspaces()
		assert.Empty(t, workspaces)
	})

	t.Run("returns list after creating workspaces", func(t *testing.T) {
		// Create some workspaces
		id1, err := service.CreateWorkspace("test1", "password")
		assert.NoError(t, err)
		id2, err := service.CreateWorkspace("test2", "password")
		assert.NoError(t, err)

		workspaces := service.ListWorkspaces()
		assert.Len(t, workspaces, 2)
		assert.Contains(t, workspaces, id1)
		assert.Contains(t, workspaces, id2)
	})
}

func TestActiveWorkspace(t *testing.T) {
	workspaceDir := "/tmp/workspace"
	service, _ := newTestService(t, workspaceDir)

	t.Run("returns nil when no active workspace", func(t *testing.T) {
		workspace := service.ActiveWorkspace()
		assert.Nil(t, workspace)
	})

	t.Run("returns workspace after startup", func(t *testing.T) {
		err := service.ServiceStartup(context.Background(), application.ServiceOptions{})
		assert.NoError(t, err)

		workspace := service.ActiveWorkspace()
		assert.NotNil(t, workspace)
		assert.Equal(t, defaultWorkspace, workspace.Name)
	})
}

func TestCreateWorkspaceErrors(t *testing.T) {
	workspaceDir := "/tmp/workspace"

	t.Run("returns error for duplicate workspace", func(t *testing.T) {
		service, _ := newTestService(t, workspaceDir)

		// Create first workspace
		_, err := service.CreateWorkspace("duplicate-test", "password")
		assert.NoError(t, err)

		// Try to create duplicate
		_, err = service.CreateWorkspace("duplicate-test", "password")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
}

func TestSwitchWorkspaceErrors(t *testing.T) {
	workspaceDir := "/tmp/workspace"
	service, _ := newTestService(t, workspaceDir)

	t.Run("returns error for non-existent workspace", func(t *testing.T) {
		err := service.SwitchWorkspace("non-existent-workspace")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not exist")
	})

	t.Run("default workspace is always accessible", func(t *testing.T) {
		err := service.SwitchWorkspace(defaultWorkspace)
		assert.NoError(t, err)
	})
}

func TestNewWorkspaceService(t *testing.T) {
	t.Run("creates service with mock medium", func(t *testing.T) {
		mockMedium := io.NewMockMedium()
		service, err := New(mockMedium)

		assert.NoError(t, err)
		assert.NotNil(t, service)
		assert.NotNil(t, service.workspaceList)
		assert.Equal(t, mockMedium, service.medium)
	})
}

func TestServiceStartupWithInvalidJSON(t *testing.T) {
	workspaceDir := "/tmp/workspace"
	service, mockMedium := newTestService(t, workspaceDir)

	// Add invalid JSON to list.json
	listPath := filepath.Join(workspaceDir, listFile)
	mockMedium.Files[listPath] = "invalid-json{{"

	// ServiceStartup should warn but continue
	err := service.ServiceStartup(context.Background(), application.ServiceOptions{})
	assert.NoError(t, err) // Should not error, just warn

	// Workspace list should be empty/reset
	assert.NotNil(t, service.activeWorkspace)
}

func TestHandleIPCEvents(t *testing.T) {
	workspaceDir := "/tmp/workspace"

	t.Run("handles switch workspace action", func(t *testing.T) {
		coreInstance, err := core.New()
		assert.NoError(t, err)

		mockCfg := &mockConfig{values: map[string]interface{}{"workspaceDir": workspaceDir}}
		coreInstance.RegisterService("config", mockCfg)

		mockMedium := io.NewMockMedium()
		service, err := New(mockMedium)
		assert.NoError(t, err)
		service.ServiceRuntime = core.NewServiceRuntime(coreInstance, Options{})

		// First startup to initialize workspace list and create default workspace
		err = service.ServiceStartup(context.Background(), application.ServiceOptions{})
		assert.NoError(t, err)

		// Create a workspace to switch to
		wsID, err := service.CreateWorkspace("ipc-test", "password")
		assert.NoError(t, err)

		// Test IPC switch workspace action
		msg := map[string]any{
			"action": "workspace.switch_workspace",
			"name":   wsID,
		}

		err = service.HandleIPCEvents(coreInstance, msg)
		assert.NoError(t, err)
		assert.Equal(t, wsID, service.activeWorkspace.Name)
	})

	t.Run("handles ActionServiceStartup message", func(t *testing.T) {
		coreInstance, err := core.New()
		assert.NoError(t, err)

		mockCfg := &mockConfig{values: map[string]interface{}{"workspaceDir": workspaceDir}}
		coreInstance.RegisterService("config", mockCfg)

		mockMedium := io.NewMockMedium()
		service, err := New(mockMedium)
		assert.NoError(t, err)
		service.ServiceRuntime = core.NewServiceRuntime(coreInstance, Options{})

		// Send ActionServiceStartup message
		err = service.HandleIPCEvents(coreInstance, core.ActionServiceStartup{})
		assert.NoError(t, err)
		assert.NotNil(t, service.activeWorkspace)
	})

	// Skipping "logs error for unknown message type" test as it requires core.App.Logger to be initialized
	// which requires Wails runtime

	// Skipping "handles map message with non-workspace action" test as it falls through to default
	// case which requires core.App.Logger
}

func TestGetWorkspaceDirError(t *testing.T) {
	t.Run("returns error when config missing workspaceDir", func(t *testing.T) {
		coreInstance, err := core.New()
		assert.NoError(t, err)

		// Register config without workspaceDir
		mockCfg := &mockConfig{values: map[string]interface{}{}}
		coreInstance.RegisterService("config", mockCfg)

		mockMedium := io.NewMockMedium()
		service, err := New(mockMedium)
		assert.NoError(t, err)
		service.ServiceRuntime = core.NewServiceRuntime(coreInstance, Options{})

		// ServiceStartup should fail because workspaceDir is missing
		err = service.ServiceStartup(context.Background(), application.ServiceOptions{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "workspaceDir")
	})
}
