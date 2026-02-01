package dev

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/repos"
	"github.com/host-uk/core/pkg/workspace"
)

// loadRegistryWithConfig loads the registry and applies workspace configuration.
func loadRegistryWithConfig(registryPath string) (*repos.Registry, string, error) {
	var reg *repos.Registry
	var err error
	var registryDir string

	if registryPath != "" {
		reg, err = repos.LoadRegistry(registryPath)
		if err != nil {
			return nil, "", cli.Wrap(err, "failed to load registry")
		}
		cli.Print("%s %s\n\n", dimStyle.Render(i18n.Label("registry")), registryPath)
		registryDir = filepath.Dir(registryPath)
	} else {
		registryPath, err = repos.FindRegistry()
		if err == nil {
			reg, err = repos.LoadRegistry(registryPath)
			if err != nil {
				return nil, "", cli.Wrap(err, "failed to load registry")
			}
			cli.Print("%s %s\n\n", dimStyle.Render(i18n.Label("registry")), registryPath)
			registryDir = filepath.Dir(registryPath)
		} else {
			// Fallback: scan current directory
			cwd, _ := os.Getwd()
			reg, err = repos.ScanDirectory(cwd)
			if err != nil {
				return nil, "", cli.Wrap(err, "failed to scan directory")
			}
			cli.Print("%s %s\n\n", dimStyle.Render(i18n.T("cmd.dev.scanning_label")), cwd)
			registryDir = cwd
		}
	}
	// Load workspace config to respect packages_dir (only if config exists)
	if wsConfig, err := workspace.LoadConfig(registryDir); err == nil && wsConfig != nil {
		if wsConfig.PackagesDir != "" {
			pkgDir := wsConfig.PackagesDir
			// Expand ~
			if strings.HasPrefix(pkgDir, "~/") {
				home, _ := os.UserHomeDir()
				pkgDir = filepath.Join(home, pkgDir[2:])
			}
			if !filepath.IsAbs(pkgDir) {
				pkgDir = filepath.Join(registryDir, pkgDir)
			}

			// Update repo paths
			for _, repo := range reg.Repos {
				repo.Path = filepath.Join(pkgDir, repo.Name)
			}
		}
	}

	return reg, registryDir, nil
}
