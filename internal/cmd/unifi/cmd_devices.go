package unifi

import (
	"strings"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/log"
	uf "github.com/host-uk/core/pkg/unifi"
)

// Devices command flags.
var (
	devicesSite string
	devicesType string
)

// addDevicesCommand adds the 'devices' subcommand for listing infrastructure devices.
func addDevicesCommand(parent *cli.Command) {
	cmd := &cli.Command{
		Use:   "devices",
		Short: "List infrastructure devices",
		Long:  "List all infrastructure devices (APs, switches, gateways) on the UniFi network.",
		RunE: func(cmd *cli.Command, args []string) error {
			return runDevices()
		},
	}

	cmd.Flags().StringVar(&devicesSite, "site", "", "Filter by site name")
	cmd.Flags().StringVar(&devicesType, "type", "", "Filter by device type (uap, usw, usg, udm, uxg)")

	parent.AddCommand(cmd)
}

func runDevices() error {
	client, err := uf.NewFromConfig("", "", "", "", nil)
	if err != nil {
		return log.E("unifi.devices", "failed to initialise client", err)
	}

	devices, err := client.GetDeviceList(devicesSite, strings.ToLower(devicesType))
	if err != nil {
		return log.E("unifi.devices", "failed to fetch devices", err)
	}

	if len(devices) == 0 {
		cli.Text("No devices found.")
		return nil
	}

	table := cli.NewTable("Name", "IP", "MAC", "Model", "Type", "Version", "Status")

	for _, d := range devices {
		status := successStyle.Render("online")
		if d.Status != 1 {
			status = errorStyle.Render("offline")
		}

		table.AddRow(
			valueStyle.Render(d.Name),
			d.IP,
			dimStyle.Render(d.Mac),
			d.Model,
			dimStyle.Render(d.Type),
			dimStyle.Render(d.Version),
			status,
		)
	}

	cli.Blank()
	cli.Print("  %d devices\n\n", len(devices))
	table.Render()

	return nil
}
