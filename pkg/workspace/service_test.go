package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/host-uk/core/pkg/crypt/openpgp"
	core "github.com/host-uk/core/pkg/framework/core"
	"github.com/stretchr/testify/assert"
)

func TestWorkspace(t *testing.T) {
	// Setup core with crypt service
	c, _ := core.New(
		core.WithName("crypt", openpgp.New),
	)

	tempHome, _ := os.MkdirTemp("", "core-test-home")
	defer os.RemoveAll(tempHome)

	// Mock os.UserHomeDir by setting HOME env
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", oldHome)

	s_any, err := New(c)
	assert.NoError(t, err)
	s := s_any.(*Service)

	// Test CreateWorkspace
	id, err := s.CreateWorkspace("test-user", "pass123")
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	wsPath := filepath.Join(tempHome, ".core", "workspaces", id)
	assert.DirExists(t, wsPath)
	assert.DirExists(t, filepath.Join(wsPath, "keys"))
	assert.FileExists(t, filepath.Join(wsPath, "keys", "private.key"))

	// Test SwitchWorkspace
	err = s.SwitchWorkspace(id)
	assert.NoError(t, err)
	assert.Equal(t, id, s.activeWorkspace)

	// Test File operations
	filename := "secret.txt"
	content := "top secret info"
	err = s.WorkspaceFileSet(filename, content)
	assert.NoError(t, err)

	got, err := s.WorkspaceFileGet(filename)
	assert.NoError(t, err)
	assert.Equal(t, content, got)
}
