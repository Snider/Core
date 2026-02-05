package workspace

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRepo(t *testing.T, dir, name string) string {
	t.Helper()
	repoPath := filepath.Join(dir, name)
	require.NoError(t, os.MkdirAll(repoPath, 0755))

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
		{"git", "commit", "--allow-empty", "-m", "initial"},
	}
	for _, c := range cmds {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Dir = repoPath
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "cmd %v failed: %s", c, string(out))
	}
	return repoPath
}

func TestTaskWorkspacePath(t *testing.T) {
	path := taskWorkspacePath("/home/user/Code/host-uk", 101, 343)
	assert.Equal(t, "/home/user/Code/host-uk/.core/workspace/p101/i343", path)
}

func TestCreateWorktree_Good(t *testing.T) {
	tmp := t.TempDir()
	repoPath := setupTestRepo(t, tmp, "test-repo")
	worktreePath := filepath.Join(tmp, "workspace", "test-repo")

	err := createWorktree(t.Context(), repoPath, worktreePath, "issue/123")
	require.NoError(t, err)

	// Verify worktree exists
	assert.DirExists(t, worktreePath)
	assert.FileExists(t, filepath.Join(worktreePath, ".git"))

	// Verify branch
	branch := gitOutput(worktreePath, "rev-parse", "--abbrev-ref", "HEAD")
	assert.Equal(t, "issue/123", trimNL(branch))
}

func TestCreateWorktree_BranchExists(t *testing.T) {
	tmp := t.TempDir()
	repoPath := setupTestRepo(t, tmp, "test-repo")

	// Create branch first
	cmd := exec.Command("git", "branch", "issue/456")
	cmd.Dir = repoPath
	require.NoError(t, cmd.Run())

	worktreePath := filepath.Join(tmp, "workspace", "test-repo")
	err := createWorktree(t.Context(), repoPath, worktreePath, "issue/456")
	require.NoError(t, err)

	assert.DirExists(t, worktreePath)
}

func TestCheckWorkspaceSafety_Clean(t *testing.T) {
	tmp := t.TempDir()
	wsPath := filepath.Join(tmp, "workspace")
	require.NoError(t, os.MkdirAll(wsPath, 0755))

	repoPath := setupTestRepo(t, tmp, "origin-repo")
	worktreePath := filepath.Join(wsPath, "origin-repo")
	require.NoError(t, createWorktree(t.Context(), repoPath, worktreePath, "test-branch"))

	dirty, reasons := checkWorkspaceSafety(wsPath)
	assert.False(t, dirty)
	assert.Empty(t, reasons)
}

func TestCheckWorkspaceSafety_Dirty(t *testing.T) {
	tmp := t.TempDir()
	wsPath := filepath.Join(tmp, "workspace")
	require.NoError(t, os.MkdirAll(wsPath, 0755))

	repoPath := setupTestRepo(t, tmp, "origin-repo")
	worktreePath := filepath.Join(wsPath, "origin-repo")
	require.NoError(t, createWorktree(t.Context(), repoPath, worktreePath, "test-branch"))

	// Create uncommitted file
	require.NoError(t, os.WriteFile(filepath.Join(worktreePath, "dirty.txt"), []byte("dirty"), 0644))

	dirty, reasons := checkWorkspaceSafety(wsPath)
	assert.True(t, dirty)
	assert.Contains(t, reasons[0], "uncommitted changes")
}

func TestEpicBranchName(t *testing.T) {
	assert.Equal(t, "epic/101", epicBranchName(101))
	assert.Equal(t, "epic/42", epicBranchName(42))
}

func trimNL(s string) string {
	return s[:len(s)-1]
}
