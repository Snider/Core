package unifi

import (
	uf "github.com/unpoller/unifi/v5"

	"github.com/host-uk/core/pkg/log"
)

// GetSites returns all sites from the UniFi controller.
func (c *Client) GetSites() ([]*uf.Site, error) {
	sites, err := c.api.GetSites()
	if err != nil {
		return nil, log.E("unifi.GetSites", "failed to fetch sites", err)
	}

	return sites, nil
}
