package docs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/host-uk/core/pkg/i18n"
	"github.com/spf13/cobra"
)

// Flag variables for sync command
var (
	docsSyncRegistryPath string
	docsSyncDryRun       bool
	docsSyncOutputDir    string
)

var docsSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: i18n.T("cmd.docs.sync.short"),
	Long:  i18n.T("cmd.docs.sync.long"),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDocsSync(docsSyncRegistryPath, docsSyncOutputDir, docsSyncDryRun)
	},
}

func init() {
	docsSyncCmd.Flags().StringVar(&docsSyncRegistryPath, "registry", "", i18n.T("common.flag.registry"))
	docsSyncCmd.Flags().BoolVar(&docsSyncDryRun, "dry-run", false, i18n.T("cmd.docs.sync.flag.dry_run"))
	docsSyncCmd.Flags().StringVar(&docsSyncOutputDir, "output", "", i18n.T("cmd.docs.sync.flag.output"))
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
		fmt.Println(i18n.T("cmd.docs.sync.no_docs_found"))
		return nil
	}

	fmt.Printf("\n%s %s\n\n", dimStyle.Render(i18n.T("cmd.docs.sync.found_label")), i18n.T("cmd.docs.sync.repos_with_docs", map[string]interface{}{"Count": len(docsInfo)}))

	// Show what will be synced
	var totalFiles int
	for _, info := range docsInfo {
		totalFiles += len(info.DocsFiles)
		outName := packageOutputName(info.Name)
		fmt.Printf("  %s → %s %s\n",
			repoNameStyle.Render(info.Name),
			docsFileStyle.Render("packages/"+outName+"/"),
			dimStyle.Render(i18n.T("cmd.docs.sync.files_count", map[string]interface{}{"Count": len(info.DocsFiles)})))

		for _, f := range info.DocsFiles {
			fmt.Printf("    %s\n", dimStyle.Render(f))
		}
	}

	fmt.Printf("\n%s %s\n",
		dimStyle.Render(i18n.T("common.label.total")),
		i18n.T("cmd.docs.sync.total_summary", map[string]interface{}{"Files": totalFiles, "Repos": len(docsInfo), "Output": outputDir}))

	if dryRun {
		fmt.Printf("\n%s\n", dimStyle.Render(i18n.T("cmd.docs.sync.dry_run_notice")))
		return nil
	}

	// Confirm
	fmt.Println()
	if !confirm(i18n.T("cmd.docs.sync.confirm")) {
		fmt.Println(i18n.T("common.prompt.abort"))
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

	fmt.Printf("\n%s %s\n", successStyle.Render(i18n.T("common.label.done")), i18n.T("cmd.docs.sync.synced_packages", map[string]interface{}{"Count": synced}))

	return nil
}
