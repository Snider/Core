package plugin

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	core "github.com/host-uk/core/pkg/framework/core"
	"github.com/host-uk/core/pkg/io"
)

// Installer handles plugin installation from GitHub.
type Installer struct {
	medium   io.Medium
	registry *Registry
}

// NewInstaller creates a new plugin installer.
func NewInstaller(m io.Medium, registry *Registry) *Installer {
	return &Installer{
		medium:   m,
		registry: registry,
	}
}

// Install downloads and installs a plugin from GitHub.
// The source format is "org/repo" or "org/repo@version".
func (i *Installer) Install(ctx context.Context, source string) error {
	org, repo, version, err := ParseSource(source)
	if err != nil {
		return core.E("plugin.Installer.Install", "invalid source", err)
	}

	// Check if already installed
	if _, exists := i.registry.Get(repo); exists {
		return core.E("plugin.Installer.Install", "plugin already installed: "+repo, nil)
	}

	// Clone the repository
	pluginDir := filepath.Join(i.registry.basePath, repo)
	if err := i.medium.EnsureDir(pluginDir); err != nil {
		return core.E("plugin.Installer.Install", "failed to create plugin directory", err)
	}

	if err := i.cloneRepo(ctx, org, repo, version, pluginDir); err != nil {
		return core.E("plugin.Installer.Install", "failed to clone repository", err)
	}

	// Load and validate manifest
	manifestPath := filepath.Join(pluginDir, "plugin.json")
	manifest, err := LoadManifest(i.medium, manifestPath)
	if err != nil {
		// Clean up on failure
		_ = i.medium.DeleteAll(pluginDir)
		return core.E("plugin.Installer.Install", "failed to load manifest", err)
	}

	if err := manifest.Validate(); err != nil {
		_ = i.medium.DeleteAll(pluginDir)
		return core.E("plugin.Installer.Install", "invalid manifest", err)
	}

	// Resolve version
	if version == "" {
		version = manifest.Version
	}

	// Register in the registry
	cfg := &PluginConfig{
		Name:        manifest.Name,
		Version:     version,
		Source:      fmt.Sprintf("github:%s/%s", org, repo),
		Enabled:     true,
		InstalledAt: time.Now().UTC().Format(time.RFC3339),
	}

	if err := i.registry.Add(cfg); err != nil {
		return core.E("plugin.Installer.Install", "failed to register plugin", err)
	}

	if err := i.registry.Save(); err != nil {
		return core.E("plugin.Installer.Install", "failed to save registry", err)
	}

	return nil
}

// Update updates a plugin to the latest version.
func (i *Installer) Update(ctx context.Context, name string) error {
	cfg, ok := i.registry.Get(name)
	if !ok {
		return core.E("plugin.Installer.Update", "plugin not found: "+name, nil)
	}

	// Parse the source to get org/repo
	source := strings.TrimPrefix(cfg.Source, "github:")
	pluginDir := filepath.Join(i.registry.basePath, name)

	// Pull latest changes
	cmd := exec.CommandContext(ctx, "git", "-C", pluginDir, "pull", "--ff-only")
	if output, err := cmd.CombinedOutput(); err != nil {
		return core.E("plugin.Installer.Update", "failed to pull updates: "+strings.TrimSpace(string(output)), err)
	}

	// Reload manifest to get updated version
	manifestPath := filepath.Join(pluginDir, "plugin.json")
	manifest, err := LoadManifest(i.medium, manifestPath)
	if err != nil {
		return core.E("plugin.Installer.Update", "failed to read updated manifest", err)
	}

	// Update registry
	cfg.Version = manifest.Version
	if err := i.registry.Save(); err != nil {
		return core.E("plugin.Installer.Update", "failed to save registry", err)
	}

	_ = source // used for context
	return nil
}

// Remove uninstalls a plugin by removing its files and registry entry.
func (i *Installer) Remove(name string) error {
	if _, ok := i.registry.Get(name); !ok {
		return core.E("plugin.Installer.Remove", "plugin not found: "+name, nil)
	}

	// Delete plugin directory
	pluginDir := filepath.Join(i.registry.basePath, name)
	if i.medium.Exists(pluginDir) {
		if err := i.medium.DeleteAll(pluginDir); err != nil {
			return core.E("plugin.Installer.Remove", "failed to delete plugin files", err)
		}
	}

	// Remove from registry
	if err := i.registry.Remove(name); err != nil {
		return core.E("plugin.Installer.Remove", "failed to unregister plugin", err)
	}

	if err := i.registry.Save(); err != nil {
		return core.E("plugin.Installer.Remove", "failed to save registry", err)
	}

	return nil
}

// cloneRepo clones a GitHub repository using the gh CLI.
func (i *Installer) cloneRepo(ctx context.Context, org, repo, version, dest string) error {
	repoURL := fmt.Sprintf("%s/%s", org, repo)

	args := []string{"repo", "clone", repoURL, dest}
	if version != "" {
		args = append(args, "--", "--branch", version)
	}

	cmd := exec.CommandContext(ctx, "gh", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %s", err, strings.TrimSpace(string(output)))
	}

	return nil
}

// ParseSource parses a plugin source string into org, repo, and version.
// Accepted formats:
//   - "org/repo" -> org="org", repo="repo", version=""
//   - "org/repo@v1.0" -> org="org", repo="repo", version="v1.0"
func ParseSource(source string) (org, repo, version string, err error) {
	if source == "" {
		return "", "", "", core.E("plugin.ParseSource", "source is empty", nil)
	}

	// Split off version if present
	atIdx := strings.LastIndex(source, "@")
	path := source
	if atIdx != -1 {
		path = source[:atIdx]
		version = source[atIdx+1:]
		if version == "" {
			return "", "", "", core.E("plugin.ParseSource", "version is empty after @", nil)
		}
	}

	// Split org/repo
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", "", core.E("plugin.ParseSource", "source must be in format org/repo[@version]", nil)
	}

	return parts[0], parts[1], version, nil
}
