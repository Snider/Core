package unifi

import (
	uf "github.com/unpoller/unifi/v5"

	"github.com/host-uk/core/pkg/log"
)

// ClientFilter controls which clients are returned.
type ClientFilter struct {
	Site     string // Filter by site name (empty = all sites)
	Wired    bool   // Show only wired clients
	Wireless bool   // Show only wireless clients
}

// GetClients returns connected clients from the UniFi controller,
// optionally filtered by site and connection type.
func (c *Client) GetClients(filter ClientFilter) ([]*uf.Client, error) {
	sites, err := c.getSitesForFilter(filter.Site)
	if err != nil {
		return nil, err
	}

	clients, err := c.api.GetClients(sites)
	if err != nil {
		return nil, log.E("unifi.GetClients", "failed to fetch clients", err)
	}

	// Apply wired/wireless filter
	if filter.Wired || filter.Wireless {
		var filtered []*uf.Client
		for _, cl := range clients {
			if filter.Wired && cl.IsWired.Val {
				filtered = append(filtered, cl)
			} else if filter.Wireless && !cl.IsWired.Val {
				filtered = append(filtered, cl)
			}
		}
		return filtered, nil
	}

	return clients, nil
}

// getSitesForFilter resolves sites by name or returns all sites.
func (c *Client) getSitesForFilter(siteName string) ([]*uf.Site, error) {
	sites, err := c.GetSites()
	if err != nil {
		return nil, err
	}

	if siteName == "" {
		return sites, nil
	}

	// Filter to matching site
	for _, s := range sites {
		if s.Name == siteName {
			return []*uf.Site{s}, nil
		}
	}

	return nil, log.E("unifi.getSitesForFilter", "site not found: "+siteName, nil)
}
