package docs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/leaanthony/clir"
)

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
