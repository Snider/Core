package collect

import (
	"context"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/collect"
	"github.com/host-uk/core/pkg/i18n"
)

// addProcessCommand adds the 'process' subcommand to the collect parent.
func addProcessCommand(parent *cli.Command) {
	processCmd := &cli.Command{
		Use:   "process <source> <dir>",
		Short: i18n.T("cmd.collect.process.short"),
		Long:  i18n.T("cmd.collect.process.long"),
		Args:  cli.ExactArgs(2),
		RunE: func(cmd *cli.Command, args []string) error {
			return runProcess(args[0], args[1])
		},
	}

	parent.AddCommand(processCmd)
}

func runProcess(source, dir string) error {
	cfg := newConfig()
	setupVerboseLogging(cfg)

	processor := &collect.Processor{
		Source: source,
		Dir:    dir,
	}

	if cfg.DryRun {
		cli.Info("Dry run: would process " + source + " data in " + dir)
		return nil
	}

	ctx := context.Background()
	result, err := processor.Process(ctx, cfg)
	if err != nil {
		return cli.Wrap(err, "processing failed")
	}

	printResult(result)
	return nil
}
