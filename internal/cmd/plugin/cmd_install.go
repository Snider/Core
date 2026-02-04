package plugin

import (
	"context"
	"os"
	"path/filepath"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/io"
	"github.com/host-uk/core/pkg/plugin"
)

func addInstallCommand(parent *cli.Command) {
	installCmd := cli.NewCommand(
		"install <source>",
		i18n.T("Install a plugin from GitHub"),
		i18n.T("Install a plugin from a GitHub repository.\n\nSource format: org/repo or org/repo@version"),
		func(cmd *cli.Command, args []string) error {
			return runInstall(args[0])
		},
	)
	installCmd.Args = cli.ExactArgs(1)
	installCmd.Example = "  core plugin install host-uk/core-plugin-example\n  core plugin install host-uk/core-plugin-example@v1.0.0"

	parent.AddCommand(installCmd)
}

func runInstall(source string) error {
	basePath, err := pluginBasePath()
	if err != nil {
		return err
	}

	registry := plugin.NewRegistry(io.Local, basePath)
	if err := registry.Load(); err != nil {
		return err
	}

	installer := plugin.NewInstaller(io.Local, registry)

	cli.Dim("Installing plugin from " + source + "...")

	if err := installer.Install(context.Background(), source); err != nil {
		return err
	}

	_, repo, _, _ := plugin.ParseSource(source)
	cli.Success("Plugin " + repo + " installed successfully")

	return nil
}

// pluginBasePath returns the default plugin directory (~/.core/plugins/).
func pluginBasePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", cli.Wrap(err, "failed to determine home directory")
	}
	return filepath.Join(home, ".core", "plugins"), nil
}
