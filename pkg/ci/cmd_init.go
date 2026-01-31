package ci

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/release"
)

// runCIReleaseInit creates a release configuration interactively.
func runCIReleaseInit() error {
	projectDir, err := os.Getwd()
	if err != nil {
		return cli.WrapVerb(err, "get", "working directory")
	}

	// Check if config already exists
	if release.ConfigExists(projectDir) {
		cli.Print("%s %s %s\n",
			releaseDimStyle.Render(i18n.Label("note")),
			i18n.T("cmd.ci.init.config_exists"),
			release.ConfigPath(projectDir))

		reader := bufio.NewReader(os.Stdin)
		cli.Print("%s", i18n.T("cmd.ci.init.overwrite_prompt"))
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			cli.Text(i18n.T("common.prompt.abort"))
			return nil
		}
	}

	cli.Print("%s %s\n", releaseHeaderStyle.Render(i18n.T("cmd.ci.label.init")), i18n.T("cmd.ci.init.creating"))
	cli.Line("")

	reader := bufio.NewReader(os.Stdin)

	// Project name
	defaultName := filepath.Base(projectDir)
	cli.Print("%s [%s]: ", i18n.T("cmd.ci.init.project_name"), defaultName)
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)
	if name == "" {
		name = defaultName
	}

	// Repository
	cli.Print("%s ", i18n.T("cmd.ci.init.github_repo"))
	repo, _ := reader.ReadString('\n')
	repo = strings.TrimSpace(repo)

	// Create config
	cfg := release.DefaultConfig()
	cfg.Project.Name = name
	cfg.Project.Repository = repo

	// Write config
	if err := release.WriteConfig(cfg, projectDir); err != nil {
		return cli.WrapVerb(err, "write", "config")
	}

	cli.Line("")
	cli.Print("%s %s %s\n",
		releaseSuccessStyle.Render(i18n.T("i18n.done.pass")),
		i18n.T("cmd.ci.init.config_written"),
		release.ConfigPath(projectDir))

	return nil
}
