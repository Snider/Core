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
func New(url, user, pass, apikey string, insecure bool) (*Client, error) {
	cfg := &uf.Config{
		URL:    url,
		User:   user,
		Pass:   pass,
		APIKey: apikey,
	}

	// Skip TLS verification if requested (e.g. for self-signed certs)
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecure,
				MinVersion:         tls.VersionTLS12,
			},
		},
	}

	api, err := uf.NewUnifi(cfg)
	if err != nil {
		return nil, log.E("unifi.New", "failed to create client", err)
	}

	// Override the HTTP client to skip TLS verification
	api.Client = httpClient

	return &Client{api: api, url: url}, nil
}

// API exposes the underlying SDK client for direct access.
func (c *Client) API() *uf.Unifi { return c.api }

// URL returns the UniFi controller URL.
func (c *Client) URL() string { return c.url }
