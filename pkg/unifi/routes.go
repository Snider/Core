package unifi

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/host-uk/core/pkg/log"
)

// Route represents a single entry in the UniFi gateway routing table.
type Route struct {
	Network   string `json:"pfx"`      // CIDR prefix (e.g. "10.69.1.0/24")
	NextHop   string `json:"nh"`       // Next-hop address or interface
	Interface string `json:"intf"`     // Interface name (e.g. "br0", "eth4")
	Type      string `json:"type"`     // Route type (e.g. "S" static, "C" connected, "K" kernel)
	Distance  int    `json:"distance"` // Administrative distance
	Metric    int    `json:"metric"`   // Route metric
	Uptime    int    `json:"uptime"`   // Uptime in seconds
	Selected  bool   `json:"fib"`      // Whether route is in the forwarding table
}

// routeResponse is the raw API response wrapper.
type routeResponse struct {
	Data []Route `json:"data"`
}

// GetRoutes returns the active routing table from the gateway for the given site.
// Uses the raw controller API since unpoller doesn't wrap this endpoint.
func (c *Client) GetRoutes(siteName string) ([]Route, error) {
	if siteName == "" {
		siteName = "default"
	}

	path := fmt.Sprintf("/api/s/%s/stat/routing", url.PathEscape(siteName))

	raw, err := c.api.GetJSON(path)
	if err != nil {
		return nil, log.E("unifi.GetRoutes", "failed to fetch routing table", err)
	}

	var resp routeResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, log.E("unifi.GetRoutes", "failed to parse routing table", err)
	}

	return resp.Data, nil
}

// RouteTypeName returns a human-readable name for the route type code.
func RouteTypeName(code string) string {
	switch code {
	case "S":
		return "static"
	case "C":
		return "connected"
	case "K":
		return "kernel"
	case "B":
		return "bgp"
	case "O":
		return "ospf"
	default:
		return code
	}
}
