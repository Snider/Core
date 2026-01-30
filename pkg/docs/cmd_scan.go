package docs

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/host-uk/core/pkg/i18n"
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
	var basePath string

	if registryPath != "" {
		reg, err = repos.LoadRegistry(registryPath)
		if err != nil {
			return nil, "", fmt.Errorf("%s: %w", i18n.T("i18n.fail.load", "registry"), err)
		}
		basePath = filepath.Dir(registryPath)
	} else {
		registryPath, err = repos.FindRegistry()
		if err == nil {
			reg, err = repos.LoadRegistry(registryPath)
			if err != nil {
				return nil, "", fmt.Errorf("%s: %w", i18n.T("i18n.fail.load", "registry"), err)
			}
			basePath = filepath.Dir(registryPath)
		} else {
			cwd, _ := os.Getwd()
			reg, err = repos.ScanDirectory(cwd)
			if err != nil {
				return nil, "", fmt.Errorf("%s: %w", i18n.T("i18n.fail.scan", "directory"), err)
			}
			basePath = cwd
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
	if _, err := os.Stat(readme); err == nil {
		info.Readme = readme
		info.HasDocs = true
	}

	// Check for CLAUDE.md
	claudeMd := filepath.Join(repo.Path, "CLAUDE.md")
	if _, err := os.Stat(claudeMd); err == nil {
		info.ClaudeMd = claudeMd
		info.HasDocs = true
	}

	// Check for CHANGELOG.md
	changelog := filepath.Join(repo.Path, "CHANGELOG.md")
	if _, err := os.Stat(changelog); err == nil {
		info.Changelog = changelog
		info.HasDocs = true
	}

	// Recursively scan docs/ directory for .md files
	docsDir := filepath.Join(repo.Path, "docs")
	if _, err := os.Stat(docsDir); err == nil {
		filepath.WalkDir(docsDir, func(path string, d fs.DirEntry, err error) error {
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

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
