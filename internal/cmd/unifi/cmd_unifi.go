// Package unifi provides CLI commands for managing a UniFi network controller.
//
// Commands:
//   - config: Configure UniFi connection (URL, credentials)
//   - clients: List connected clients
//   - devices: List infrastructure devices
//   - sites: List controller sites
//   - networks: List network segments and VLANs
//   - routes: List gateway routing table
package unifi

import (
	"github.com/host-uk/core/pkg/cli"
)

func init() {
	cli.RegisterCommands(AddUniFiCommands)
}

// Style aliases from shared package.
var (
	successStyle = cli.SuccessStyle
	errorStyle   = cli.ErrorStyle
	warningStyle = cli.WarningStyle
	dimStyle     = cli.DimStyle
	valueStyle   = cli.ValueStyle
	numberStyle  = cli.NumberStyle
	infoStyle    = cli.InfoStyle
)

// AddUniFiCommands registers the 'unifi' command and all subcommands.
func AddUniFiCommands(root *cli.Command) {
	unifiCmd := &cli.Command{
		Use:   "unifi",
		Short: "UniFi network management",
		Long:  "Manage sites, devices, and connected clients on your UniFi controller.",
	}
	root.AddCommand(unifiCmd)

	addConfigCommand(unifiCmd)
	addClientsCommand(unifiCmd)
	addDevicesCommand(unifiCmd)
	addNetworksCommand(unifiCmd)
	addRoutesCommand(unifiCmd)
	addSitesCommand(unifiCmd)
}
