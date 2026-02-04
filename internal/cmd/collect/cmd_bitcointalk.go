package collect

import (
	"context"
	"strings"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/collect"
	"github.com/host-uk/core/pkg/i18n"
)

// BitcoinTalk command flags
var bitcointalkPages int

// addBitcoinTalkCommand adds the 'bitcointalk' subcommand to the collect parent.
func addBitcoinTalkCommand(parent *cli.Command) {
	btcCmd := &cli.Command{
		Use:   "bitcointalk <topic-id|url>",
		Short: i18n.T("cmd.collect.bitcointalk.short"),
		Long:  i18n.T("cmd.collect.bitcointalk.long"),
		Args:  cli.ExactArgs(1),
		RunE: func(cmd *cli.Command, args []string) error {
			return runBitcoinTalk(args[0])
		},
	}

	cli.IntFlag(btcCmd, &bitcointalkPages, "pages", "p", 0, i18n.T("cmd.collect.bitcointalk.flag.pages"))

	parent.AddCommand(btcCmd)
}

func runBitcoinTalk(target string) error {
	var topicID, url string

	// Determine if argument is a URL or topic ID
	if strings.HasPrefix(target, "http") {
		url = target
	} else {
		topicID = target
	}

	cfg := newConfig()
	setupVerboseLogging(cfg)

	collector := &collect.BitcoinTalkCollector{
		TopicID: topicID,
		URL:     url,
		Pages:   bitcointalkPages,
	}

	if cfg.DryRun {
		cli.Info("Dry run: would collect from BitcoinTalk topic " + target)
		return nil
	}

	ctx := context.Background()
	result, err := collector.Collect(ctx, cfg)
	if err != nil {
		return cli.Wrap(err, "bitcointalk collection failed")
	}

	printResult(result)
	return nil
}
