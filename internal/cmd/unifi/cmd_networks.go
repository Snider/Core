package unifi

import (
	"fmt"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/log"
	uf "github.com/host-uk/core/pkg/unifi"
)

// Networks command flags.
var (
	networksSite string
)

// addNetworksCommand adds the 'networks' subcommand for listing network segments.
func addNetworksCommand(parent *cli.Command) {
	cmd := &cli.Command{
		Use:   "networks",
		Short: "List network segments",
		Long:  "List all network segments configured on the UniFi controller, showing VLANs, subnets, isolation, and DHCP.",
		RunE: func(cmd *cli.Command, args []string) error {
			return runNetworks()
		},
	}

	cmd.Flags().StringVar(&networksSite, "site", "", "Site name (default: \"default\")")

	parent.AddCommand(cmd)
}

func runNetworks() error {
	client, err := uf.NewFromConfig("", "", "", "", false)
	if err != nil {
		return log.E("unifi.networks", "failed to initialise client", err)
	}

	networks, err := client.GetNetworks(networksSite)
	if err != nil {
		return log.E("unifi.networks", "failed to fetch networks", err)
	}

	if len(networks) == 0 {
		cli.Text("No networks found.")
		return nil
	}

	// Separate WANs, LANs, and VPNs
	var wans, lans, vpns []uf.NetworkConf
	for _, n := range networks {
		switch n.Purpose {
		case "wan":
			wans = append(wans, n)
		case "remote-user-vpn":
			vpns = append(vpns, n)
		default:
			lans = append(lans, n)
		}
	}

	cli.Blank()

	// WANs
	if len(wans) > 0 {
		cli.Print("  %s\n\n", infoStyle.Render("WAN Interfaces"))
		wanTable := cli.NewTable("Name", "Type", "Group", "Status")
		for _, w := range wans {
			status := successStyle.Render("enabled")
			if !w.Enabled {
				status = errorStyle.Render("disabled")
			}
			wanTable.AddRow(
				valueStyle.Render(w.Name),
				dimStyle.Render(w.WANType),
				dimStyle.Render(w.WANNetworkGroup),
				status,
			)
		}
		wanTable.Render()
		cli.Blank()
	}

	// LANs
	if len(lans) > 0 {
		cli.Print("  %s\n\n", infoStyle.Render("LAN Networks"))
		lanTable := cli.NewTable("Name", "Subnet", "VLAN", "Isolated", "Internet", "DHCP", "mDNS")
		for _, n := range lans {
			vlan := dimStyle.Render("-")
			if n.VLANEnabled {
				vlan = numberStyle.Render(fmt.Sprintf("%d", n.VLAN))
			}

			isolated := successStyle.Render("no")
			if n.NetworkIsolationEnabled {
				isolated = warningStyle.Render("yes")
			}

			internet := successStyle.Render("yes")
			if !n.InternetAccessEnabled {
				internet = errorStyle.Render("no")
			}

			dhcp := dimStyle.Render("off")
			if n.DHCPEnabled {
				dhcp = fmt.Sprintf("%s - %s", n.DHCPStart, n.DHCPStop)
			}

			mdns := dimStyle.Render("off")
			if n.MDNSEnabled {
				mdns = successStyle.Render("on")
			}

			lanTable.AddRow(
				valueStyle.Render(n.Name),
				n.IPSubnet,
				vlan,
				isolated,
				internet,
				dhcp,
				mdns,
			)
		}
		lanTable.Render()
		cli.Blank()
	}

	// VPNs
	if len(vpns) > 0 {
		cli.Print("  %s\n\n", infoStyle.Render("VPN Networks"))
		vpnTable := cli.NewTable("Name", "Subnet", "Type")
		for _, v := range vpns {
			vpnTable.AddRow(
				valueStyle.Render(v.Name),
				v.IPSubnet,
				dimStyle.Render(v.VPNType),
			)
		}
		vpnTable.Render()
		cli.Blank()
	}

	cli.Print("  %s\n\n", dimStyle.Render(fmt.Sprintf("%d networks total", len(networks))))

	return nil
}
