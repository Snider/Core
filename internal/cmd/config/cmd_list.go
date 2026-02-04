package config

import (
	"fmt"

	"github.com/host-uk/core/pkg/cli"
	"gopkg.in/yaml.v3"
)

func addListCommand(parent *cli.Command) {
	cmd := cli.NewCommand("list", "List all configuration values", "", func(cmd *cli.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		all := cfg.All()
		if len(all) == 0 {
			cli.Dim("No configuration values set")
			return nil
		}

		out, err := yaml.Marshal(all)
		if err != nil {
			return cli.Wrap(err, "failed to format config")
		}

		fmt.Print(string(out))
		return nil
	})

	cli.WithArgs(cmd, cli.NoArgs())

	parent.AddCommand(cmd)
}
