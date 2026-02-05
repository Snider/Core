package unifi

import (
	"errors"
	"fmt"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/log"
	uf "github.com/host-uk/core/pkg/unifi"
)

// Clients command flags.
var (
	clientsSite     string
	clientsWired    bool
	clientsWireless bool
)

// addClientsCommand adds the 'clients' subcommand for listing connected clients.
func addClientsCommand(parent *cli.Command) {
	cmd := &cli.Command{
		Use:   "clients",
		Short: "List connected clients",
		Long:  "List all connected clients on the UniFi network, optionally filtered by site or connection type.",
		RunE: func(cmd *cli.Command, args []string) error {
			return runClients()
		},
	}

	cmd.Flags().StringVar(&clientsSite, "site", "", "Filter by site name")
	cmd.Flags().BoolVar(&clientsWired, "wired", false, "Show only wired clients")
	cmd.Flags().BoolVar(&clientsWireless, "wireless", false, "Show only wireless clients")

	parent.AddCommand(cmd)
}

func runClients() error {
	if clientsWired && clientsWireless {
		return log.E("unifi.clients", "conflicting flags", errors.New("--wired and --wireless cannot both be set"))
	}

	client, err := uf.NewFromConfig("", "", "", "", false)
	if err != nil {
		return log.E("unifi.clients", "failed to initialise client", err)
	}

	clients, err := client.GetClients(uf.ClientFilter{
		Site:     clientsSite,
		Wired:    clientsWired,
		Wireless: clientsWireless,
	})
	if err != nil {
		return log.E("unifi.clients", "failed to fetch clients", err)
	}

	if len(clients) == 0 {
		cli.Text("No clients found.")
		return nil
	}

	table := cli.NewTable("Name", "IP", "MAC", "Network", "Type", "Uptime")

	for _, cl := range clients {
		name := cl.Name
		if name == "" {
			name = cl.Hostname
		}
		if name == "" {
			name = "(unknown)"
		}

		connType := cl.Essid
		if cl.IsWired.Val {
			connType = "wired"
		}

		table.AddRow(
			valueStyle.Render(name),
			cl.IP,
			dimStyle.Render(cl.Mac),
			cl.Network,
			dimStyle.Render(connType),
			dimStyle.Render(formatUptime(cl.Uptime.Int())),
		)
	}

	cli.Blank()
	cli.Print("  %d clients\n\n", len(clients))
	table.Render()

	return nil
}

// formatUptime converts seconds to a human-readable duration string.
func formatUptime(seconds int) string {
	if seconds <= 0 {
		return "-"
	}

	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60

	switch {
	case days > 0:
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	case hours > 0:
		return fmt.Sprintf("%dh %dm", hours, minutes)
	default:
		return fmt.Sprintf("%dm", minutes)
	}
}
