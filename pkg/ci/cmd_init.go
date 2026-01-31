package ci

import (
	"os"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/release"
)

func runCIReleaseInit() error {
	cwd, err := os.Getwd()
	if err != nil {
		return cli.Err("%s: %w", i18n.T("i18n.fail.get", "working directory"), err)
	}

	cli.Print("%s %s\n\n", releaseDimStyle.Render(i18n.Label("init")), i18n.T("cmd.ci.init.initializing"))

	// Check if already initialized
	if release.ConfigExists(cwd) {
		cli.Text(i18n.T("cmd.ci.init.already_initialized"))
		return nil
	}

	// Create release config
	cfg := release.DefaultConfig()
	if err := release.WriteConfig(cfg, cwd); err != nil {
		return cli.Err("%s: %w", i18n.T("i18n.fail.create", "config"), err)
	}

	cli.Blank()
	cli.Print("%s %s\n", releaseSuccessStyle.Render("v"), i18n.T("cmd.ci.init.created_config"))

	// Templates init removed as functionality not exposed

	cli.Blank()

	cli.Text(i18n.T("cmd.ci.init.next_steps"))
	cli.Print("  %s\n", i18n.T("cmd.ci.init.edit_config"))
	cli.Print("  %s\n", i18n.T("cmd.ci.init.run_ci"))

	return nil
}