package plugin

import (
	"context"
	"fmt"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/io"
	"github.com/host-uk/core/pkg/plugin"
)

var updateAll bool

func addUpdateCommand(parent *cli.Command) {
	updateCmd := cli.NewCommand(
		"update [name]",
		i18n.T("Update a plugin or all plugins"),
		i18n.T("Update a specific plugin to the latest version, or use --all to update all installed plugins."),
		func(cmd *cli.Command, args []string) error {
			if updateAll {
				return runUpdateAll()
			}
			if len(args) == 0 {
				return fmt.Errorf("plugin name required (or use --all)")
			}
			return runUpdate(args[0])
		},
	)

	cli.BoolFlag(updateCmd, &updateAll, "all", "a", false, i18n.T("Update all installed plugins"))

	parent.AddCommand(updateCmd)
}

func runUpdate(name string) error {
	basePath, err := pluginBasePath()
	if err != nil {
		return err
	}

	registry := plugin.NewRegistry(io.Local, basePath)
	if err := registry.Load(); err != nil {
		return err
	}

	installer := plugin.NewInstaller(io.Local, registry)

	cli.Dim("Updating " + name + "...")

	if err := installer.Update(context.Background(), name); err != nil {
		return err
	}

	cli.Success("Plugin " + name + " updated successfully")
	return nil
}

func runUpdateAll() error {
	basePath, err := pluginBasePath()
	if err != nil {
		return err
	}

	registry := plugin.NewRegistry(io.Local, basePath)
	if err := registry.Load(); err != nil {
		return err
	}

	plugins := registry.List()
	if len(plugins) == 0 {
		cli.Dim("No plugins installed")
		return nil
	}

	installer := plugin.NewInstaller(io.Local, registry)
	ctx := context.Background()

	var updated, failed int
	for _, p := range plugins {
		cli.Dim("Updating " + p.Name + "...")
		if err := installer.Update(ctx, p.Name); err != nil {
			cli.Errorf("Failed to update %s: %v", p.Name, err)
			failed++
			continue
		}
		cli.Success(p.Name + " updated")
		updated++
	}

	fmt.Println()
	cli.Dim(fmt.Sprintf("%d updated, %d failed", updated, failed))
	return nil
}
