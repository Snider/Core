package plugin

import (
	"fmt"
	"path/filepath"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/io"
	"github.com/host-uk/core/pkg/plugin"
)

func addInfoCommand(parent *cli.Command) {
	infoCmd := cli.NewCommand(
		"info <name>",
		i18n.T("Show detailed plugin information"),
		"",
		func(cmd *cli.Command, args []string) error {
			return runInfo(args[0])
		},
	)
	infoCmd.Args = cli.ExactArgs(1)

	parent.AddCommand(infoCmd)
}

func runInfo(name string) error {
	basePath, err := pluginBasePath()
	if err != nil {
		return err
	}

	registry := plugin.NewRegistry(io.Local, basePath)
	if err := registry.Load(); err != nil {
		return err
	}

	cfg, ok := registry.Get(name)
	if !ok {
		return fmt.Errorf("plugin not found: %s", name)
	}

	// Try to load the manifest for extended information
	loader := plugin.NewLoader(io.Local, basePath)
	manifest, manifestErr := loader.LoadPlugin(name)

	fmt.Println()
	cli.Label("Name", cfg.Name)
	cli.Label("Version", cfg.Version)
	cli.Label("Source", cfg.Source)

	status := "disabled"
	if cfg.Enabled {
		status = "enabled"
	}
	cli.Label("Status", status)
	cli.Label("Installed", cfg.InstalledAt)
	cli.Label("Path", filepath.Join(basePath, name))

	if manifestErr == nil && manifest != nil {
		if manifest.Description != "" {
			cli.Label("Description", manifest.Description)
		}
		if manifest.Author != "" {
			cli.Label("Author", manifest.Author)
		}
		if manifest.Entrypoint != "" {
			cli.Label("Entrypoint", manifest.Entrypoint)
		}
		if manifest.MinVersion != "" {
			cli.Label("Min Version", manifest.MinVersion)
		}
		if len(manifest.Dependencies) > 0 {
			for i, dep := range manifest.Dependencies {
				if i == 0 {
					cli.Label("Dependencies", dep)
				} else {
					fmt.Printf("               %s\n", dep)
				}
			}
		}
	}

	fmt.Println()
	return nil
}
