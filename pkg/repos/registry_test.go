package repos

import (
	"testing"

	"github.com/host-uk/core/pkg/io"
	"github.com/stretchr/testify/assert"
)

func TestLoadRegistry(t *testing.T) {
	m := io.NewMockMedium()
	yaml := `
version: 1
org: host-uk
base_path: /tmp/repos
repos:
  core:
    type: foundation
    description: Core package
`
	_ = m.Write("/tmp/repos.yaml", yaml)

	reg, err := LoadRegistry(m, "/tmp/repos.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, reg)
	assert.Equal(t, "host-uk", reg.Org)
	assert.Equal(t, "/tmp/repos", reg.BasePath)
	assert.Equal(t, m, reg.medium)

	repo, ok := reg.Get("core")
	assert.True(t, ok)
	assert.Equal(t, "core", repo.Name)
	assert.Equal(t, "/tmp/repos/core", repo.Path)
	assert.Equal(t, reg, repo.registry)
}

func TestRepo_Exists(t *testing.T) {
	m := io.NewMockMedium()
	reg := &Registry{
		medium:   m,
		BasePath: "/tmp/repos",
		Repos:    make(map[string]*Repo),
	}
	repo := &Repo{
		Name:     "core",
		Path:     "/tmp/repos/core",
		registry: reg,
	}

	// Not exists yet
	assert.False(t, repo.Exists())

	// Create directory in mock
	_ = m.EnsureDir("/tmp/repos/core")
	assert.True(t, repo.Exists())
}

func TestRepo_IsGitRepo(t *testing.T) {
	m := io.NewMockMedium()
	reg := &Registry{
		medium:   m,
		BasePath: "/tmp/repos",
		Repos:    make(map[string]*Repo),
	}
	repo := &Repo{
		Name:     "core",
		Path:     "/tmp/repos/core",
		registry: reg,
	}

	// Not a git repo yet
	assert.False(t, repo.IsGitRepo())

	// Create .git directory in mock
	_ = m.EnsureDir("/tmp/repos/core/.git")
	assert.True(t, repo.IsGitRepo())
}
