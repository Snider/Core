package unifi

import (
	"encoding/json"
	"fmt"

	"github.com/host-uk/core/pkg/log"
)

// NetworkConf represents a UniFi network configuration entry.
type NetworkConf struct {
	ID                      string `json:"_id"`
	Name                    string `json:"name"`
	Purpose                 string `json:"purpose"`      // wan, corporate, remote-user-vpn
	IPSubnet                string `json:"ip_subnet"`    // CIDR (e.g. "10.69.1.1/24")
	VLAN                    int    `json:"vlan"`         // VLAN ID (0 = untagged)
	VLANEnabled             bool   `json:"vlan_enabled"` // Whether VLAN tagging is active
	Enabled                 bool   `json:"enabled"`
	NetworkGroup            string `json:"networkgroup"` // LAN, WAN, WAN2
	NetworkIsolationEnabled bool   `json:"network_isolation_enabled"`
	InternetAccessEnabled   bool   `json:"internet_access_enabled"`
	IsNAT                   bool   `json:"is_nat"`
	DHCPEnabled             bool   `json:"dhcpd_enabled"`
	DHCPStart               string `json:"dhcpd_start"`
	DHCPStop                string `json:"dhcpd_stop"`
	DHCPDNS1                string `json:"dhcpd_dns_1"`
	DHCPDNS2                string `json:"dhcpd_dns_2"`
	DHCPDNSEnabled          bool   `json:"dhcpd_dns_enabled"`
	MDNSEnabled             bool   `json:"mdns_enabled"`
	FirewallZoneID          string `json:"firewall_zone_id"`
	GatewayType             string `json:"gateway_type"`
	VPNType                 string `json:"vpn_type"`
	WANType                 string `json:"wan_type"` // pppoe, dhcp, static
	WANNetworkGroup         string `json:"wan_networkgroup"`
}

// networkConfResponse is the raw API response wrapper.
type networkConfResponse struct {
	Data []NetworkConf `json:"data"`
}

// GetNetworks returns all network configurations from the controller.
// Uses the raw controller API for the full networkconf data.
func (c *Client) GetNetworks(siteName string) ([]NetworkConf, error) {
	if siteName == "" {
		siteName = "default"
	}

	path := fmt.Sprintf("/api/s/%s/rest/networkconf", siteName)

	raw, err := c.api.GetJSON(path)
	if err != nil {
		return nil, log.E("unifi.GetNetworks", "failed to fetch networks", err)
	}

	var resp networkConfResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, log.E("unifi.GetNetworks", "failed to parse networks", err)
	}

	return resp.Data, nil
}
