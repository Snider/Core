package ci

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/host-uk/core/pkg/release"
)

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
