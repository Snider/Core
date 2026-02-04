package plugin

import (
	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/io"
	"github.com/host-uk/core/pkg/plugin"
)

func addRemoveCommand(parent *cli.Command) {
	removeCmd := cli.NewCommand(
		"remove <name>",
		i18n.T("Remove an installed plugin"),
		"",
		func(cmd *cli.Command, args []string) error {
			return runRemove(args[0])
		},
	)
	removeCmd.Args = cli.ExactArgs(1)

	parent.AddCommand(removeCmd)
}

func runRemove(name string) error {
	basePath, err := pluginBasePath()
	if err != nil {
		return err
	}

	registry := plugin.NewRegistry(io.Local, basePath)
	if err := registry.Load(); err != nil {
		return err
	}

	if !cli.Confirm("Remove plugin " + name + "?") {
		cli.Dim("Cancelled")
		return nil
	}

	installer := plugin.NewInstaller(io.Local, registry)

	if err := installer.Remove(name); err != nil {
		return err
	}

	cli.Success("Plugin " + name + " removed")
	return nil
}
