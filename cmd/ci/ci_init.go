package ci

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/release"
)

// runCIReleaseInit creates a release configuration interactively.
func runCIReleaseInit() error {
	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("common.error.failed", map[string]any{"Action": "get working directory"}), err)
	}

	// Check if config already exists
	if release.ConfigExists(projectDir) {
		fmt.Printf("%s %s %s\n",
			releaseDimStyle.Render(i18n.T("common.label.note")),
			i18n.T("cmd.ci.init.config_exists"),
			release.ConfigPath(projectDir))

		reader := bufio.NewReader(os.Stdin)
		fmt.Print(i18n.T("cmd.ci.init.overwrite_prompt"))
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println(i18n.T("cli.confirm.abort"))
			return nil
		}
	}

	fmt.Printf("%s %s\n", releaseHeaderStyle.Render(i18n.T("cmd.ci.label.init")), i18n.T("cmd.ci.init.creating"))
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	// Project name
	defaultName := filepath.Base(projectDir)
	fmt.Printf("%s [%s]: ", i18n.T("cmd.ci.init.project_name"), defaultName)
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)
	if name == "" {
		name = defaultName
	}

	// Repository
	fmt.Printf("%s ", i18n.T("cmd.ci.init.github_repo"))
	repo, _ := reader.ReadString('\n')
	repo = strings.TrimSpace(repo)

	// Create config
	cfg := release.DefaultConfig()
	cfg.Project.Name = name
	cfg.Project.Repository = repo

	// Write config
	if err := release.WriteConfig(cfg, projectDir); err != nil {
		return fmt.Errorf("%s: %w", i18n.T("common.error.failed", map[string]any{"Action": "write config"}), err)
	}

	fmt.Println()
	fmt.Printf("%s %s %s\n",
		releaseSuccessStyle.Render(i18n.T("common.label.success")),
		i18n.T("cmd.ci.init.config_written"),
		release.ConfigPath(projectDir))

	return nil
}
