package config

import (
	"github.com/host-uk/core/pkg/cli"
)

func addSetCommand(parent *cli.Command) {
	cmd := cli.NewCommand("set", "Set a configuration value", "", func(cmd *cli.Command, args []string) error {
		key := args[0]
		value := args[1]

		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		if err := cfg.Set(key, value); err != nil {
			return cli.Wrap(err, "failed to set config value")
		}

		cli.Success(key + " = " + value)
		return nil
	})

	cli.WithArgs(cmd, cli.ExactArgs(2))
	cli.WithExample(cmd, "core config set dev.editor vim")

	parent.AddCommand(cmd)
}
