package unifi

import (
	"crypto/tls"
	"net/http"

	uf "github.com/unpoller/unifi/v5"

	"github.com/host-uk/core/pkg/log"
)

// Client wraps the unpoller UniFi client with config-based auth.
type Client struct {
	api *uf.Unifi
	url string
}

// New creates a new UniFi API client for the given controller URL and credentials.
// TLS verification can be disabled via the insecure parameter (useful for self-signed certs on home lab controllers).
// WARNING: Setting insecure=true disables certificate verification and should only be used in trusted network
// environments with self-signed certificates (e.g., home lab controllers).
func New(url, user, pass, apikey string, insecure bool) (*Client, error) {
	cfg := &uf.Config{
		URL:    url,
		User:   user,
		Pass:   pass,
		APIKey: apikey,
	}

	api, err := uf.NewUnifi(cfg)
	if err != nil {
		return nil, log.E("unifi.New", "failed to create client", err)
	}

	// Only override the HTTP client if insecure mode is explicitly requested
	if insecure {
		// #nosec G402 -- InsecureSkipVerify is intentionally enabled for home lab controllers
		// with self-signed certificates. This is opt-in via the insecure parameter.
		api.Client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
					MinVersion:         tls.VersionTLS12,
				},
			},
		}
	}

	return &Client{api: api, url: url}, nil
}

// API exposes the underlying SDK client for direct access.
func (c *Client) API() *uf.Unifi { return c.api }

// URL returns the UniFi controller URL.
func (c *Client) URL() string { return c.url }
