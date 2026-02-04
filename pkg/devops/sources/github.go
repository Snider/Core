package sources

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/host-uk/core/pkg/io"
)

// GitHubSource downloads images from GitHub Releases.
type GitHubSource struct {
	config SourceConfig
}

// Compile-time interface check.
var _ ImageSource = (*GitHubSource)(nil)

// NewGitHubSource creates a new GitHub source.
func NewGitHubSource(cfg SourceConfig) *GitHubSource {
	return &GitHubSource{config: cfg}
}

// Name returns "github".
func (s *GitHubSource) Name() string {
	return "github"
}

// Available checks if gh CLI is installed and authenticated.
func (s *GitHubSource) Available() bool {
	_, err := exec.LookPath("gh")
	if err != nil {
		return false
	}
	// Check if authenticated
	cmd := exec.Command("gh", "auth", "status")
	return cmd.Run() == nil
}

// LatestVersion returns the latest release tag.
func (s *GitHubSource) LatestVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "gh", "release", "view",
		"-R", s.config.GitHubRepo,
		"--json", "tagName",
		"-q", ".tagName",
	)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("github.LatestVersion: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// Download downloads the image from the latest release.
func (s *GitHubSource) Download(ctx context.Context, m io.Medium, dest string, progress func(downloaded, total int64)) error {
	// Get release assets to find our image
	cmd := exec.CommandContext(ctx, "gh", "release", "download",
		"-R", s.config.GitHubRepo,
		"-p", s.config.ImageName,
		"-D", dest,
		"--clobber",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("github.Download: %w", err)
	}
	return nil
}
