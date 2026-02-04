package config

import (
	"fmt"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/config"
)

func addGetCommand(parent *cli.Command) {
	cmd := cli.NewCommand("get", "Get a configuration value", "", func(cmd *cli.Command, args []string) error {
		key := args[0]

		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		var value any
		if err := cfg.Get(key, &value); err != nil {
			return cli.Err("key not found: %s", key)
		}

		fmt.Println(value)
		return nil
	})

	cli.WithArgs(cmd, cli.ExactArgs(1))
	cli.WithExample(cmd, "core config get dev.editor")

	parent.AddCommand(cmd)
}

func loadConfig() (*config.Config, error) {
	cfg, err := config.New()
	if err != nil {
		return nil, cli.Wrap(err, "failed to load config")
	}
	return cfg, nil
}
