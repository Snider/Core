package sources

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitHubSource_Good_Available(t *testing.T) {
	src := NewGitHubSource(SourceConfig{
		GitHubRepo: "host-uk/core-images",
		ImageName:  "core-devops-darwin-arm64.qcow2",
	})

	if src.Name() != "github" {
		t.Errorf("expected name 'github', got %q", src.Name())
	}

	// Available depends on gh CLI being installed
	_ = src.Available()
}

func TestGitHubSource_Name(t *testing.T) {
	src := NewGitHubSource(SourceConfig{})
	assert.Equal(t, "github", src.Name())
}

func TestGitHubSource_Config(t *testing.T) {
	cfg := SourceConfig{
		GitHubRepo: "owner/repo",
		ImageName:  "test-image.qcow2",
	}
	src := NewGitHubSource(cfg)

	// Verify the config is stored
	assert.Equal(t, "owner/repo", src.config.GitHubRepo)
	assert.Equal(t, "test-image.qcow2", src.config.ImageName)
}

func TestGitHubSource_Good_Multiple(t *testing.T) {
	// Test creating multiple sources with different configs
	src1 := NewGitHubSource(SourceConfig{GitHubRepo: "org1/repo1", ImageName: "img1.qcow2"})
	src2 := NewGitHubSource(SourceConfig{GitHubRepo: "org2/repo2", ImageName: "img2.qcow2"})

	assert.Equal(t, "org1/repo1", src1.config.GitHubRepo)
	assert.Equal(t, "org2/repo2", src2.config.GitHubRepo)
	assert.Equal(t, "github", src1.Name())
	assert.Equal(t, "github", src2.Name())
}

func TestNewGitHubSource_Good(t *testing.T) {
	cfg := SourceConfig{
		GitHubRepo:    "host-uk/core-images",
		RegistryImage: "ghcr.io/host-uk/core-devops",
		CDNURL:        "https://cdn.example.com",
		ImageName:     "core-devops-darwin-arm64.qcow2",
	}

	src := NewGitHubSource(cfg)
	assert.NotNil(t, src)
	assert.Equal(t, "github", src.Name())
	assert.Equal(t, cfg.GitHubRepo, src.config.GitHubRepo)
}

func TestGitHubSource_InterfaceCompliance(t *testing.T) {
	// Verify GitHubSource implements ImageSource
	var _ ImageSource = (*GitHubSource)(nil)
}
