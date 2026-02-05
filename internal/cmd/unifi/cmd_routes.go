package unifi

import (
	"fmt"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/log"
	uf "github.com/host-uk/core/pkg/unifi"
)

// Routes command flags.
var (
	routesSite string
	routesType string
)

// addRoutesCommand adds the 'routes' subcommand for listing the gateway routing table.
func addRoutesCommand(parent *cli.Command) {
	cmd := &cli.Command{
		Use:   "routes",
		Short: "List gateway routing table",
		Long:  "List the active routing table from the UniFi gateway, showing network segments and next-hop destinations.",
		RunE: func(cmd *cli.Command, args []string) error {
			return runRoutes()
		},
	}

	cmd.Flags().StringVar(&routesSite, "site", "", "Site name (default: \"default\")")
	cmd.Flags().StringVar(&routesType, "type", "", "Filter by route type (static, connected, kernel, bgp, ospf)")

	parent.AddCommand(cmd)
}

func runRoutes() error {
	client, err := uf.NewFromConfig("", "", "", "", nil)
	if err != nil {
		return log.E("unifi.routes", "failed to initialise client", err)
	}

	routes, err := client.GetRoutes(routesSite)
	if err != nil {
		return log.E("unifi.routes", "failed to fetch routes", err)
	}

	// Filter by type if requested
	if routesType != "" {
		var filtered []uf.Route
		for _, r := range routes {
			if uf.RouteTypeName(r.Type) == routesType || r.Type == routesType {
				filtered = append(filtered, r)
			}
		}
		routes = filtered
	}

	if len(routes) == 0 {
		cli.Text("No routes found.")
		return nil
	}

	table := cli.NewTable("Network", "Next Hop", "Interface", "Type", "Distance", "FIB")

	for _, r := range routes {
		typeName := uf.RouteTypeName(r.Type)

		fib := dimStyle.Render("no")
		if r.Selected {
			fib = successStyle.Render("yes")
		}

		table.AddRow(
			valueStyle.Render(r.Network),
			r.NextHop,
			dimStyle.Render(r.Interface),
			dimStyle.Render(typeName),
			fmt.Sprintf("%d", r.Distance),
			fib,
		)
	}

	cli.Blank()
	cli.Print("  %d routes\n\n", len(routes))
	table.Render()

	return nil
}
