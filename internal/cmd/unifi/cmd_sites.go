package unifi

import (
	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/log"
	uf "github.com/host-uk/core/pkg/unifi"
)

// addSitesCommand adds the 'sites' subcommand for listing UniFi sites.
func addSitesCommand(parent *cli.Command) {
	cmd := &cli.Command{
		Use:   "sites",
		Short: "List controller sites",
		Long:  "List all sites configured on the UniFi controller.",
		RunE: func(cmd *cli.Command, args []string) error {
			return runSites()
		},
	}

	parent.AddCommand(cmd)
}

func runSites() error {
	client, err := uf.NewFromConfig("", "", "", "", nil)
	if err != nil {
		return log.E("unifi.sites", "failed to initialise client", err)
	}

	sites, err := client.GetSites()
	if err != nil {
		return log.E("unifi.sites", "failed to fetch sites", err)
	}

	if len(sites) == 0 {
		cli.Text("No sites found.")
		return nil
	}

	table := cli.NewTable("Name", "Description")

	for _, s := range sites {
		table.AddRow(
			valueStyle.Render(s.Name),
			dimStyle.Render(s.Desc),
		)
	}

	cli.Blank()
	cli.Print("  %d sites\n\n", len(sites))
	table.Render()

	return nil
}
