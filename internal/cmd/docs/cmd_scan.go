package docs

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/host-uk/core/internal/cmd/workspace"
	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/io"
	"github.com/host-uk/core/pkg/repos"
)

// RepoDocInfo holds documentation info for a repo
type RepoDocInfo struct {
	Name      string
	Path      string
	HasDocs   bool
	Readme    string
	ClaudeMd  string
	Changelog string
	DocsFiles []string // All files in docs/ directory (recursive)
}

func loadRegistry(registryPath string) (*repos.Registry, string, error) {
	var reg *repos.Registry
	var err error
	var registryDir string

	if registryPath != "" {
		reg, err = repos.LoadRegistry(io.Local, registryPath)
		if err != nil {
			return nil, "", cli.Wrap(err, i18n.T("i18n.fail.load", "registry"))
		}
		registryDir = filepath.Dir(registryPath)
	} else {
		registryPath, err = repos.FindRegistry(io.Local)
		if err == nil {
			reg, err = repos.LoadRegistry(io.Local, registryPath)
			if err != nil {
				return nil, "", cli.Wrap(err, i18n.T("i18n.fail.load", "registry"))
			}
			registryDir = filepath.Dir(registryPath)
		} else {
			cwd, _ := os.Getwd()
			reg, err = repos.ScanDirectory(io.Local, cwd)
			if err != nil {
				return nil, "", cli.Wrap(err, i18n.T("i18n.fail.scan", "directory"))
			}
			registryDir = cwd
		}
	}

	// Load workspace config to respect packages_dir
	wsConfig, err := workspace.LoadConfig(registryDir)
	if err != nil {
		return nil, "", cli.Wrap(err, i18n.T("i18n.fail.load", "workspace config"))
	}

	basePath := registryDir

	if wsConfig != nil && wsConfig.PackagesDir != "" && wsConfig.PackagesDir != "./packages" {
		pkgDir := wsConfig.PackagesDir

		// Expand ~
		if strings.HasPrefix(pkgDir, "~/") {
			home, _ := os.UserHomeDir()
			pkgDir = filepath.Join(home, pkgDir[2:])
		}

		if !filepath.IsAbs(pkgDir) {
			pkgDir = filepath.Join(registryDir, pkgDir)
		}
		basePath = pkgDir

		// Update repo paths if they were relative to registry
		// This ensures consistency when packages_dir overrides the default
		reg.BasePath = basePath
		for _, repo := range reg.Repos {
			repo.Path = filepath.Join(basePath, repo.Name)
		}
	}

	return reg, basePath, nil
}

func scanRepoDocs(repo *repos.Repo) RepoDocInfo {
	info := RepoDocInfo{
		Name: repo.Name,
		Path: repo.Path,
	}

	// Check for README.md
	readme := filepath.Join(repo.Path, "README.md")
	if io.Local.IsFile(readme) {
		info.Readme = readme
		info.HasDocs = true
	}

	// Check for CLAUDE.md
	claudeMd := filepath.Join(repo.Path, "CLAUDE.md")
	if io.Local.IsFile(claudeMd) {
		info.ClaudeMd = claudeMd
		info.HasDocs = true
	}

	// Check for CHANGELOG.md
	changelog := filepath.Join(repo.Path, "CHANGELOG.md")
	if io.Local.IsFile(changelog) {
		info.Changelog = changelog
		info.HasDocs = true
	}

	// Recursively scan docs/ directory for .md files
	docsDir := filepath.Join(repo.Path, "docs")
	// Check if directory exists by listing it
	if _, err := io.Local.List(docsDir); err == nil {
		_ = filepath.WalkDir(docsDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			// Skip plans/ directory
			if d.IsDir() && d.Name() == "plans" {
				return filepath.SkipDir
			}
			// Skip non-markdown files
			if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
				return nil
			}
			// Get relative path from docs/
			relPath, _ := filepath.Rel(docsDir, path)
			info.DocsFiles = append(info.DocsFiles, relPath)
			info.HasDocs = true
			return nil
		})
	}

	return info
}
