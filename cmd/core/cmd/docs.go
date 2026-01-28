package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/host-uk/core/pkg/repos"
	"github.com/charmbracelet/lipgloss"
	"github.com/leaanthony/clir"
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
	DocsDir   []string // Files in docs/ directory
}

// AddDocsCommand adds the 'docs' command to the given parent command.
func AddDocsCommand(parent *clir.Cli) {
	docsCmd := parent.NewSubCommand("docs", "Documentation management")
	docsCmd.LongDescription("Manage documentation across all repos.\n" +
		"Scan for docs, check coverage, and sync to core.help.")

	// Add subcommands
	addDocsSyncCommand(docsCmd)
	addDocsListCommand(docsCmd)
}

func addDocsSyncCommand(parent *clir.Command) {
	var registryPath string
	var dryRun bool
	var outputDir string

	syncCmd := parent.NewSubCommand("sync", "Sync documentation to output directory")
	syncCmd.StringFlag("registry", "Path to repos.yaml", &registryPath)
	syncCmd.BoolFlag("dry-run", "Show what would be synced without copying", &dryRun)
	syncCmd.StringFlag("output", "Output directory (default: ./docs-build)", &outputDir)

	syncCmd.Action(func() error {
		if outputDir == "" {
			outputDir = "./docs-build"
		}
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

func runDocsSync(registryPath string, outputDir string, dryRun bool) error {
	// Find or use provided registry
	var reg *repos.Registry
	var err error

	if registryPath != "" {
		reg, err = repos.LoadRegistry(registryPath)
		if err != nil {
			return fmt.Errorf("failed to load registry: %w", err)
		}
	} else {
		registryPath, err = repos.FindRegistry()
		if err == nil {
			reg, err = repos.LoadRegistry(registryPath)
			if err != nil {
				return fmt.Errorf("failed to load registry: %w", err)
			}
		} else {
			cwd, _ := os.Getwd()
			reg, err = repos.ScanDirectory(cwd)
			if err != nil {
				return fmt.Errorf("failed to scan directory: %w", err)
			}
		}
	}

	// Scan all repos for docs
	var docsInfo []RepoDocInfo
	for _, repo := range reg.List() {
		info := scanRepoDocs(repo)
		if info.HasDocs {
			docsInfo = append(docsInfo, info)
		}
	}

	if len(docsInfo) == 0 {
		fmt.Println("No documentation found in any repos.")
		return nil
	}

	fmt.Printf("\n%s %d repo(s) with documentation\n\n", dimStyle.Render("Found"), len(docsInfo))

	// Show what will be synced
	var totalFiles int
	for _, info := range docsInfo {
		files := countDocFiles(info)
		totalFiles += files
		fmt.Printf("  %s %s\n", repoNameStyle.Render(info.Name), dimStyle.Render(fmt.Sprintf("(%d files)", files)))

		if info.Readme != "" {
			fmt.Printf("    %s\n", docsFileStyle.Render("README.md"))
		}
		if info.ClaudeMd != "" {
			fmt.Printf("    %s\n", docsFileStyle.Render("CLAUDE.md"))
		}
		if info.Changelog != "" {
			fmt.Printf("    %s\n", docsFileStyle.Render("CHANGELOG.md"))
		}
		for _, f := range info.DocsDir {
			fmt.Printf("    %s\n", docsFileStyle.Render("docs/"+f))
		}
	}

	fmt.Printf("\n%s %d files from %d repos\n", dimStyle.Render("Total:"), totalFiles, len(docsInfo))

	if dryRun {
		fmt.Printf("\n%s\n", dimStyle.Render("Dry run - no files copied"))
		return nil
	}

	// Confirm
	fmt.Println()
	if !confirm(fmt.Sprintf("Sync to %s?", outputDir)) {
		fmt.Println("Aborted.")
		return nil
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Sync docs
	fmt.Println()
	var synced int
	for _, info := range docsInfo {
		repoOutDir := filepath.Join(outputDir, "packages", info.Name)
		if err := os.MkdirAll(repoOutDir, 0755); err != nil {
			fmt.Printf("  %s %s: %s\n", errorStyle.Render("✗"), info.Name, err)
			continue
		}

		// Copy files
		if info.Readme != "" {
			copyFile(info.Readme, filepath.Join(repoOutDir, "index.md"))
		}
		if info.ClaudeMd != "" {
			copyFile(info.ClaudeMd, filepath.Join(repoOutDir, "claude.md"))
		}
		if info.Changelog != "" {
			copyFile(info.Changelog, filepath.Join(repoOutDir, "changelog.md"))
		}
		for _, f := range info.DocsDir {
			src := filepath.Join(info.Path, "docs", f)
			dst := filepath.Join(repoOutDir, f)
			os.MkdirAll(filepath.Dir(dst), 0755)
			copyFile(src, dst)
		}

		fmt.Printf("  %s %s\n", successStyle.Render("✓"), info.Name)
		synced++
	}

	fmt.Printf("\n%s Synced %d repos to %s\n", successStyle.Render("Done:"), synced, outputDir)

	return nil
}

func runDocsList(registryPath string) error {
	// Find or use provided registry
	var reg *repos.Registry
	var err error

	if registryPath != "" {
		reg, err = repos.LoadRegistry(registryPath)
		if err != nil {
			return fmt.Errorf("failed to load registry: %w", err)
		}
	} else {
		registryPath, err = repos.FindRegistry()
		if err == nil {
			reg, err = repos.LoadRegistry(registryPath)
			if err != nil {
				return fmt.Errorf("failed to load registry: %w", err)
			}
		} else {
			cwd, _ := os.Getwd()
			reg, err = repos.ScanDirectory(cwd)
			if err != nil {
				return fmt.Errorf("failed to scan directory: %w", err)
			}
		}
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
		if len(info.DocsDir) > 0 {
			docsDir = docsFoundStyle.Render(fmt.Sprintf("%d files", len(info.DocsDir)))
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

	// Check for docs/ directory
	docsDir := filepath.Join(repo.Path, "docs")
	if entries, err := os.ReadDir(docsDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				info.DocsDir = append(info.DocsDir, e.Name())
				info.HasDocs = true
			}
		}
	}

	return info
}

func countDocFiles(info RepoDocInfo) int {
	count := len(info.DocsDir)
	if info.Readme != "" {
		count++
	}
	if info.ClaudeMd != "" {
		count++
	}
	if info.Changelog != "" {
		count++
	}
	return count
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
