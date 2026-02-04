package collect

import (
	"context"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/collect"
	"github.com/host-uk/core/pkg/i18n"
)

// Market command flags
var (
	marketHistorical bool
	marketFromDate   string
)

// addMarketCommand adds the 'market' subcommand to the collect parent.
func addMarketCommand(parent *cli.Command) {
	marketCmd := &cli.Command{
		Use:   "market <coin>",
		Short: i18n.T("cmd.collect.market.short"),
		Long:  i18n.T("cmd.collect.market.long"),
		Args:  cli.ExactArgs(1),
		RunE: func(cmd *cli.Command, args []string) error {
			return runMarket(args[0])
		},
	}

	cli.BoolFlag(marketCmd, &marketHistorical, "historical", "H", false, i18n.T("cmd.collect.market.flag.historical"))
	cli.StringFlag(marketCmd, &marketFromDate, "from", "f", "", i18n.T("cmd.collect.market.flag.from"))

	parent.AddCommand(marketCmd)
}

func runMarket(coinID string) error {
	cfg := newConfig()
	setupVerboseLogging(cfg)

	collector := &collect.MarketCollector{
		CoinID:     coinID,
		Historical: marketHistorical,
		FromDate:   marketFromDate,
	}

	if cfg.DryRun {
		cli.Info("Dry run: would collect market data for " + coinID)
		return nil
	}

	ctx := context.Background()
	result, err := collector.Collect(ctx, cfg)
	if err != nil {
		return cli.Wrap(err, "market collection failed")
	}

	printResult(result)
	return nil
}
