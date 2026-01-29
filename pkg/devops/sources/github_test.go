package sources

import (
	"testing"
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
