package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/host-uk/core/pkg/release"
	"github.com/leaanthony/clir"
)

// CIRelease command styles
var (
	releaseHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#3b82f6")) // blue-500

	releaseSuccessStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#22c55e")) // green-500

	releaseErrorStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#ef4444")) // red-500

	releaseDimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6b7280")) // gray-500

	releaseValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#e2e8f0")) // gray-200
)

// AddCIReleaseCommand adds the release command and its subcommands.
func AddCIReleaseCommand(app *clir.Cli) {
	releaseCmd := app.NewSubCommand("ci", "Build and publish releases")
	releaseCmd.LongDescription("Builds release artifacts, generates changelog, and publishes to GitHub.\n" +
		"Configuration can be provided via .core/release.yaml or command-line flags.")

	// Flags for the main release command
	var dryRun bool
	var version string
	var draft bool
	var prerelease bool

	releaseCmd.BoolFlag("dry-run", "Preview release without publishing", &dryRun)
	releaseCmd.StringFlag("version", "Version to release (e.g., v1.2.3)", &version)
	releaseCmd.BoolFlag("draft", "Create release as a draft", &draft)
	releaseCmd.BoolFlag("prerelease", "Mark release as a prerelease", &prerelease)

	// Default action for `core ci` - publish only (expects artifacts in dist/)
	releaseCmd.Action(func() error {
		return runCIPublish(dryRun, version, draft, prerelease)
	})

	// `release init` subcommand
	initCmd := releaseCmd.NewSubCommand("init", "Initialize release configuration")
	initCmd.LongDescription("Creates a .core/release.yaml configuration file interactively.")
	initCmd.Action(func() error {
		return runCIReleaseInit()
	})

	// `release changelog` subcommand
	changelogCmd := releaseCmd.NewSubCommand("changelog", "Generate changelog")
	changelogCmd.LongDescription("Generates a changelog from conventional commits.")
	var fromRef, toRef string
	changelogCmd.StringFlag("from", "Starting ref (default: previous tag)", &fromRef)
	changelogCmd.StringFlag("to", "Ending ref (default: HEAD)", &toRef)
	changelogCmd.Action(func() error {
		return runChangelog(fromRef, toRef)
	})

	// `release version` subcommand
	versionCmd := releaseCmd.NewSubCommand("version", "Show or set version")
	versionCmd.LongDescription("Shows the determined version or validates a version string.")
	versionCmd.Action(func() error {
		return runCIReleaseVersion()
	})
}

// runCIPublish publishes pre-built artifacts from dist/.
// It does NOT build - use `core build` first.
func runCIPublish(dryRun bool, version string, draft, prerelease bool) error {
	ctx := context.Background()

	// Get current directory
	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Load configuration
	cfg, err := release.LoadConfig(projectDir)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Apply CLI overrides
	if version != "" {
		cfg.SetVersion(version)
	}

	// Apply draft/prerelease overrides to all publishers
	if draft || prerelease {
		for i := range cfg.Publishers {
			if draft {
				cfg.Publishers[i].Draft = true
			}
			if prerelease {
				cfg.Publishers[i].Prerelease = true
			}
		}
	}

	// Print header
	fmt.Printf("%s Publishing release\n", releaseHeaderStyle.Render("CI:"))
	if dryRun {
		fmt.Printf("  %s\n", releaseDimStyle.Render("(dry-run mode)"))
	}
	fmt.Println()

	// Check for publishers
	if len(cfg.Publishers) == 0 {
		return fmt.Errorf("no publishers configured in .core/release.yaml")
	}

	// Publish pre-built artifacts
	rel, err := release.Publish(ctx, cfg, dryRun)
	if err != nil {
		fmt.Printf("%s %v\n", releaseErrorStyle.Render("Error:"), err)
		return err
	}

	// Print summary
	fmt.Println()
	fmt.Printf("%s Publish completed!\n", releaseSuccessStyle.Render("Success:"))
	fmt.Printf("  Version:   %s\n", releaseValueStyle.Render(rel.Version))
	fmt.Printf("  Artifacts: %d\n", len(rel.Artifacts))

	if !dryRun {
		for _, pub := range cfg.Publishers {
			fmt.Printf("  Published: %s\n", releaseValueStyle.Render(pub.Type))
		}
	}

	return nil
}

// runCIReleaseInit creates a release configuration interactively.
func runCIReleaseInit() error {
	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Check if config already exists
	if release.ConfigExists(projectDir) {
		fmt.Printf("%s Configuration already exists at %s\n",
			releaseDimStyle.Render("Note:"),
			release.ConfigPath(projectDir))

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Overwrite? [y/N]: ")
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	fmt.Printf("%s Creating release configuration\n", releaseHeaderStyle.Render("Init:"))
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	// Project name
	defaultName := filepath.Base(projectDir)
	fmt.Printf("Project name [%s]: ", defaultName)
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)
	if name == "" {
		name = defaultName
	}

	// Repository
	fmt.Print("GitHub repository (owner/repo): ")
	repo, _ := reader.ReadString('\n')
	repo = strings.TrimSpace(repo)

	// Create config
	cfg := release.DefaultConfig()
	cfg.Project.Name = name
	cfg.Project.Repository = repo

	// Write config
	if err := release.WriteConfig(cfg, projectDir); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Println()
	fmt.Printf("%s Configuration written to %s\n",
		releaseSuccessStyle.Render("Success:"),
		release.ConfigPath(projectDir))

	return nil
}

// runChangelog generates and prints a changelog.
func runChangelog(fromRef, toRef string) error {
	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Load config for changelog settings
	cfg, err := release.LoadConfig(projectDir)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Generate changelog
	changelog, err := release.GenerateWithConfig(projectDir, fromRef, toRef, &cfg.Changelog)
	if err != nil {
		return fmt.Errorf("failed to generate changelog: %w", err)
	}

	fmt.Println(changelog)
	return nil
}

// runCIReleaseVersion shows the determined version.
func runCIReleaseVersion() error {
	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	version, err := release.DetermineVersion(projectDir)
	if err != nil {
		return fmt.Errorf("failed to determine version: %w", err)
	}

	fmt.Printf("Version: %s\n", releaseValueStyle.Render(version))
	return nil
}
