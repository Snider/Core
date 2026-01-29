package sources

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSourceConfig_Empty(t *testing.T) {
	cfg := SourceConfig{}
	assert.Empty(t, cfg.GitHubRepo)
	assert.Empty(t, cfg.RegistryImage)
	assert.Empty(t, cfg.CDNURL)
	assert.Empty(t, cfg.ImageName)
}

func TestSourceConfig_Complete(t *testing.T) {
	cfg := SourceConfig{
		GitHubRepo:    "owner/repo",
		RegistryImage: "ghcr.io/owner/image:v1",
		CDNURL:        "https://cdn.example.com/images",
		ImageName:     "my-image-darwin-arm64.qcow2",
	}

	assert.Equal(t, "owner/repo", cfg.GitHubRepo)
	assert.Equal(t, "ghcr.io/owner/image:v1", cfg.RegistryImage)
	assert.Equal(t, "https://cdn.example.com/images", cfg.CDNURL)
	assert.Equal(t, "my-image-darwin-arm64.qcow2", cfg.ImageName)
}

func TestImageSource_Interface(t *testing.T) {
	// Ensure both sources implement the interface
	var _ ImageSource = (*GitHubSource)(nil)
	var _ ImageSource = (*CDNSource)(nil)
}
