package unifi

import (
	uf "github.com/unpoller/unifi/v5"

	"github.com/host-uk/core/pkg/log"
)

// DeviceInfo is a flat representation of any UniFi infrastructure device.
type DeviceInfo struct {
	Name    string
	IP      string
	Mac     string
	Model   string
	Version string
	Type    string // uap, usw, usg, udm, uxg
	Status  int    // 1 = online
}

// GetDevices returns the raw device container for a site (or all sites).
func (c *Client) GetDevices(siteName string) (*uf.Devices, error) {
	sites, err := c.getSitesForFilter(siteName)
	if err != nil {
		return nil, err
	}

	devices, err := c.api.GetDevices(sites)
	if err != nil {
		return nil, log.E("unifi.GetDevices", "failed to fetch devices", err)
	}

	return devices, nil
}

// GetDeviceList returns a flat list of all infrastructure devices,
// optionally filtered by device type (uap, usw, usg, udm, uxg).
func (c *Client) GetDeviceList(siteName, deviceType string) ([]DeviceInfo, error) {
	devices, err := c.GetDevices(siteName)
	if err != nil {
		return nil, err
	}

	var list []DeviceInfo

	if deviceType == "" || deviceType == "uap" {
		for _, d := range devices.UAPs {
			list = append(list, DeviceInfo{
				Name:    d.Name,
				IP:      d.IP,
				Mac:     d.Mac,
				Model:   d.Model,
				Version: d.Version,
				Type:    "uap",
				Status:  d.State.Int(),
			})
		}
	}

	if deviceType == "" || deviceType == "usw" {
		for _, d := range devices.USWs {
			list = append(list, DeviceInfo{
				Name:    d.Name,
				IP:      d.IP,
				Mac:     d.Mac,
				Model:   d.Model,
				Version: d.Version,
				Type:    "usw",
				Status:  d.State.Int(),
			})
		}
	}

	if deviceType == "" || deviceType == "usg" {
		for _, d := range devices.USGs {
			list = append(list, DeviceInfo{
				Name:    d.Name,
				IP:      d.IP,
				Mac:     d.Mac,
				Model:   d.Model,
				Version: d.Version,
				Type:    "usg",
				Status:  d.State.Int(),
			})
		}
	}

	if deviceType == "" || deviceType == "udm" {
		for _, d := range devices.UDMs {
			list = append(list, DeviceInfo{
				Name:    d.Name,
				IP:      d.IP,
				Mac:     d.Mac,
				Model:   d.Model,
				Version: d.Version,
				Type:    "udm",
				Status:  d.State.Int(),
			})
		}
	}

	if deviceType == "" || deviceType == "uxg" {
		for _, d := range devices.UXGs {
			list = append(list, DeviceInfo{
				Name:    d.Name,
				IP:      d.IP,
				Mac:     d.Mac,
				Model:   d.Model,
				Version: d.Version,
				Type:    "uxg",
				Status:  d.State.Int(),
			})
		}
	}

	return list, nil
}
