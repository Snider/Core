package config

import (
	"fmt"

	"github.com/host-uk/core/pkg/cli"
)

func addPathCommand(parent *cli.Command) {
	cmd := cli.NewCommand("path", "Show the configuration file path", "", func(cmd *cli.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		fmt.Println(cfg.Path())
		return nil
	})

	cli.WithArgs(cmd, cli.NoArgs())

	parent.AddCommand(cmd)
}
