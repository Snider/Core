// Package docs provides documentation management commands.
package docs

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/host-uk/core/cmd/core/cmd/shared"
	"github.com/host-uk/core/pkg/repos"
	"github.com/leaanthony/clir"
)

// Style and utility aliases
var (
	repoNameStyle = shared.RepoNameStyle
	successStyle  = shared.SuccessStyle
	errorStyle    = shared.ErrorStyle
	dimStyle      = shared.DimStyle
	headerStyle   = shared.HeaderStyle
	confirm       = shared.Confirm
)

var (
	docsFoundStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#22c55e")) // green-500

	docsMissingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6b7280")) // gray-500

	docsFileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3b82f6")) // blue-500
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

// AddDocsCommand adds the 'docs' command to the given parent command.
func AddDocsCommand(parent *clir.Cli) {
	docsCmd := parent.NewSubCommand("docs", "Documentation management")
	docsCmd.LongDescription("Manage documentation across all repos.\n" +
		"Scan for docs, check coverage, and sync to core-php/docs/packages/.")

	// Add subcommands
	addDocsSyncCommand(docsCmd)
	addDocsListCommand(docsCmd)
}

func addDocsSyncCommand(parent *clir.Command) {
	var registryPath string
	var dryRun bool
	var outputDir string

	syncCmd := parent.NewSubCommand("sync", "Sync documentation to core-php/docs/packages/")
	syncCmd.StringFlag("registry", "Path to repos.yaml", &registryPath)
	syncCmd.BoolFlag("dry-run", "Show what would be synced without copying", &dryRun)
	syncCmd.StringFlag("output", "Output directory (default: core-php/docs/packages)", &outputDir)

	syncCmd.Action(func() error {
		return runDocsSync(registryPath, outputDir, dryRun)
	})
}

func addDocsListCommand(parent *clir.Command) {
	var registryPath string

	listCmd := parent.NewSubCommand("list", "List documentation across repos")
	listCmd.StringFlag("registry", "Path to repos.yaml", &registryPath)

	listCmd.Action(func() error {
		return runDocsList(registryPath)
	})
}

// packageOutputName maps repo name to output folder name
func packageOutputName(repoName string) string {
	// core -> go (the Go framework)
	if repoName == "core" {
		return "go"
	}
	// core-admin -> admin, core-api -> api, etc.
	if strings.HasPrefix(repoName, "core-") {
		return strings.TrimPrefix(repoName, "core-")
	}
	return repoName
}

// shouldSyncRepo returns true if this repo should be synced
func shouldSyncRepo(repoName string) bool {
	// Skip core-php (it's the destination)
	if repoName == "core-php" {
		return false
	}
	// Skip template
	if repoName == "core-template" {
		return false
	}
	return true
}

func runDocsSync(registryPath string, outputDir string, dryRun bool) error {
	// Find or use provided registry
	reg, basePath, err := loadRegistry(registryPath)
	if err != nil {
		return err
	}

	// Default output to core-php/docs/packages relative to registry
	if outputDir == "" {
		outputDir = filepath.Join(basePath, "core-php", "docs", "packages")
	}

	// Scan all repos for docs
	var docsInfo []RepoDocInfo
	for _, repo := range reg.List() {
		if !shouldSyncRepo(repo.Name) {
			continue
		}
		info := scanRepoDocs(repo)
		if info.HasDocs && len(info.DocsFiles) > 0 {
			docsInfo = append(docsInfo, info)
		}
	}

	if len(docsInfo) == 0 {
		fmt.Println("No documentation found in any repos.")
		return nil
	}

	fmt.Printf("\n%s %d repo(s) with docs/ directories\n\n", dimStyle.Render("Found"), len(docsInfo))

	// Show what will be synced
	var totalFiles int
	for _, info := range docsInfo {
		totalFiles += len(info.DocsFiles)
		outName := packageOutputName(info.Name)
		fmt.Printf("  %s → %s %s\n",
			repoNameStyle.Render(info.Name),
			docsFileStyle.Render("packages/"+outName+"/"),
			dimStyle.Render(fmt.Sprintf("(%d files)", len(info.DocsFiles))))

		for _, f := range info.DocsFiles {
			fmt.Printf("    %s\n", dimStyle.Render(f))
		}
	}

	fmt.Printf("\n%s %d files from %d repos → %s\n",
		dimStyle.Render("Total:"), totalFiles, len(docsInfo), outputDir)

	if dryRun {
		fmt.Printf("\n%s\n", dimStyle.Render("Dry run - no files copied"))
		return nil
	}

	// Confirm
	fmt.Println()
	if !confirm("Sync?") {
		fmt.Println("Aborted.")
		return nil
	}

	// Sync docs
	fmt.Println()
	var synced int
	for _, info := range docsInfo {
		outName := packageOutputName(info.Name)
		repoOutDir := filepath.Join(outputDir, outName)

		// Clear existing directory
		os.RemoveAll(repoOutDir)

		if err := os.MkdirAll(repoOutDir, 0755); err != nil {
			fmt.Printf("  %s %s: %s\n", errorStyle.Render("✗"), info.Name, err)
			continue
		}

		// Copy all docs files
		docsDir := filepath.Join(info.Path, "docs")
		for _, f := range info.DocsFiles {
			src := filepath.Join(docsDir, f)
			dst := filepath.Join(repoOutDir, f)
			os.MkdirAll(filepath.Dir(dst), 0755)
			if err := copyFile(src, dst); err != nil {
				fmt.Printf("  %s %s: %s\n", errorStyle.Render("✗"), f, err)
			}
		}

		fmt.Printf("  %s %s → packages/%s/\n", successStyle.Render("✓"), info.Name, outName)
		synced++
	}

	fmt.Printf("\n%s Synced %d packages\n", successStyle.Render("Done:"), synced)

	return nil
}

func runDocsList(registryPath string) error {
	reg, _, err := loadRegistry(registryPath)
	if err != nil {
		return err
	}

	fmt.Printf("\n%-20s  %-8s  %-8s  %-10s  %s\n",
		headerStyle.Render("Repo"),
		headerStyle.Render("README"),
		headerStyle.Render("CLAUDE"),
		headerStyle.Render("CHANGELOG"),
		headerStyle.Render("docs/"),
	)
	fmt.Println(strings.Repeat("─", 70))

	var withDocs, withoutDocs int
	for _, repo := range reg.List() {
		info := scanRepoDocs(repo)

		readme := docsMissingStyle.Render("—")
		if info.Readme != "" {
			readme = docsFoundStyle.Render("✓")
		}

		claude := docsMissingStyle.Render("—")
		if info.ClaudeMd != "" {
			claude = docsFoundStyle.Render("✓")
		}

		changelog := docsMissingStyle.Render("—")
		if info.Changelog != "" {
			changelog = docsFoundStyle.Render("✓")
		}

		docsDir := docsMissingStyle.Render("—")
		if len(info.DocsFiles) > 0 {
			docsDir = docsFoundStyle.Render(fmt.Sprintf("%d files", len(info.DocsFiles)))
		}

		fmt.Printf("%-20s  %-8s  %-8s  %-10s  %s\n",
			repoNameStyle.Render(info.Name),
			readme,
			claude,
			changelog,
			docsDir,
		)

		if info.HasDocs {
			withDocs++
		} else {
			withoutDocs++
		}
	}

	fmt.Println()
	fmt.Printf("%s %d with docs, %d without\n",
		dimStyle.Render("Coverage:"),
		withDocs,
		withoutDocs,
	)

	return nil
}

func loadRegistry(registryPath string) (*repos.Registry, string, error) {
	var reg *repos.Registry
	var err error
	var basePath string

	if registryPath != "" {
		reg, err = repos.LoadRegistry(registryPath)
		if err != nil {
			return nil, "", fmt.Errorf("failed to load registry: %w", err)
		}
		basePath = filepath.Dir(registryPath)
	} else {
		registryPath, err = repos.FindRegistry()
		if err == nil {
			reg, err = repos.LoadRegistry(registryPath)
			if err != nil {
				return nil, "", fmt.Errorf("failed to load registry: %w", err)
			}
			basePath = filepath.Dir(registryPath)
		} else {
			cwd, _ := os.Getwd()
			reg, err = repos.ScanDirectory(cwd)
			if err != nil {
				return nil, "", fmt.Errorf("failed to scan directory: %w", err)
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
