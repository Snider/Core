//go:generate go run github.com/Snider/Core/pkg/updater/build

// Package updater provides functionality for self-updating Go applications.
// It supports updates from GitHub releases and generic HTTP endpoints.
package updater

import (
	"fmt"
	"net/url"
	"strings"
)

// StartupCheckMode defines the updater's behavior on startup.
type StartupCheckMode int

const (
	// NoCheck disables any checks on startup.
	NoCheck StartupCheckMode = iota
	// CheckOnStartup checks for updates on startup but does not apply them.
	CheckOnStartup
	// CheckAndUpdateOnStartup checks for and applies updates on startup.
	CheckAndUpdateOnStartup
)

// UpdateServiceConfig holds the configuration for the UpdateService.
type UpdateServiceConfig struct {
	// RepoURL is the URL to the repository for updates. It can be a GitHub
	// repository URL (e.g., "https://github.com/owner/repo") or a base URL
	// for a generic HTTP update server.
	RepoURL string
	// Channel specifies the release channel to track (e.g., "stable", "prerelease").
	// This is only used for GitHub-based updates.
	Channel string
	// CheckOnStartup determines the update behavior when the service starts.
	CheckOnStartup StartupCheckMode
	// ForceSemVerPrefix toggles whether to enforce a 'v' prefix on version tags for display.
	// If true, a 'v' prefix is added if missing. If false, it's removed if present.
	ForceSemVerPrefix bool
	// ReleaseURLFormat provides a template for constructing the download URL for a
	// release asset. The placeholder {tag} will be replaced with the release tag.
	ReleaseURLFormat string
}

// UpdateService provides a configurable interface for handling application updates.
// It can be configured to check for updates on startup and, if desired, apply
// them automatically. The service can handle updates from both GitHub releases
// and generic HTTP servers.
type UpdateService struct {
	config   UpdateServiceConfig
	isGitHub bool
	owner    string
	repo     string
}

// NewUpdateService creates and configures a new UpdateService.
// It parses the repository URL to determine if it's a GitHub repository
// and extracts the owner and repo name.
func NewUpdateService(config UpdateServiceConfig) (*UpdateService, error) {
	isGitHub := strings.Contains(config.RepoURL, "github.com")
	var owner, repo string
	var err error

	if isGitHub {
		owner, repo, err = ParseRepoURL(config.RepoURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse GitHub repo URL: %w", err)
		}
	}

	return &UpdateService{
		config:   config,
		isGitHub: isGitHub,
		owner:    owner,
		repo:     repo,
	}, nil
}

// Start initiates the update check based on the service configuration.
// It determines whether to perform a GitHub or HTTP-based update check
// based on the RepoURL. The behavior of the check is controlled by the
// CheckOnStartup setting in the configuration.
func (s *UpdateService) Start() error {
	if s.isGitHub {
		return s.startGitHubCheck()
	}
	return s.startHTTPCheck()
}

func (s *UpdateService) startGitHubCheck() error {
	switch s.config.CheckOnStartup {
	case NoCheck:
		return nil // Do nothing
	case CheckOnStartup:
		return CheckOnly(s.owner, s.repo, s.config.Channel, s.config.ForceSemVerPrefix, s.config.ReleaseURLFormat)
	case CheckAndUpdateOnStartup:
		return CheckForUpdates(s.owner, s.repo, s.config.Channel, s.config.ForceSemVerPrefix, s.config.ReleaseURLFormat)
	default:
		return fmt.Errorf("unknown startup check mode: %d", s.config.CheckOnStartup)
	}
}

func (s *UpdateService) startHTTPCheck() error {
	switch s.config.CheckOnStartup {
	case NoCheck:
		return nil // Do nothing
	case CheckOnStartup:
		return CheckOnlyHTTP(s.config.RepoURL)
	case CheckAndUpdateOnStartup:
		return CheckForUpdatesHTTP(s.config.RepoURL)
	default:
		return fmt.Errorf("unknown startup check mode: %d", s.config.CheckOnStartup)
	}
}

// ParseRepoURL extracts the owner and repository name from a GitHub URL.
// It handles standard GitHub URL formats.
func ParseRepoURL(repoURL string) (owner string, repo string, err error) {
	u, err := url.Parse(repoURL)
	if err != nil {
		return "", "", err
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid repo URL path: %s", u.Path)
	}
	return parts[0], parts[1], nil
}
