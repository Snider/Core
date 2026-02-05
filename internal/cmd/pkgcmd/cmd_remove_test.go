package pkgcmd

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

func TestCheckRepoSafety_Clean(t *testing.T) {
	tmp := t.TempDir()
	repoPath := setupTestRepo(t, tmp, "clean-repo")

	blocked, reasons := checkRepoSafety(repoPath)
	assert.False(t, blocked)
	assert.Empty(t, reasons)
}

func TestCheckRepoSafety_UncommittedChanges(t *testing.T) {
	tmp := t.TempDir()
	repoPath := setupTestRepo(t, tmp, "dirty-repo")

	require.NoError(t, os.WriteFile(filepath.Join(repoPath, "new.txt"), []byte("data"), 0644))

	blocked, reasons := checkRepoSafety(repoPath)
	assert.True(t, blocked)
	assert.NotEmpty(t, reasons)
	assert.Contains(t, reasons[0], "uncommitted changes")
}

func TestCheckRepoSafety_Stash(t *testing.T) {
	tmp := t.TempDir()
	repoPath := setupTestRepo(t, tmp, "stash-repo")

	// Create a file, add, stash
	require.NoError(t, os.WriteFile(filepath.Join(repoPath, "stash.txt"), []byte("data"), 0644))
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = repoPath
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "stash")
	cmd.Dir = repoPath
	require.NoError(t, cmd.Run())

	blocked, reasons := checkRepoSafety(repoPath)
	assert.True(t, blocked)
	found := false
	for _, r := range reasons {
		if assert.ObjectsAreEqual("stashed", "") || len(r) > 0 {
			if contains(r, "stash") {
				found = true
			}
		}
	}
	assert.True(t, found, "expected stash warning in reasons: %v", reasons)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
