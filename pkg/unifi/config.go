// Package unifi provides a thin wrapper around the unpoller/unifi Go SDK
// for managing UniFi network controllers, devices, and connected clients.
//
// Authentication is resolved from config file, environment variables, or flag overrides:
//
//  1. ~/.core/config.yaml keys: unifi.url, unifi.user, unifi.pass, unifi.apikey
//  2. UNIFI_URL + UNIFI_USER + UNIFI_PASS + UNIFI_APIKEY environment variables (override config file)
//  3. Flag overrides via core unifi config --url/--user/--pass/--apikey (highest priority)
package unifi

import (
	"os"

	"github.com/host-uk/core/pkg/config"
	"github.com/host-uk/core/pkg/log"
)

const (
	// ConfigKeyURL is the config key for the UniFi controller URL.
	ConfigKeyURL = "unifi.url"
	// ConfigKeyUser is the config key for the UniFi username.
	ConfigKeyUser = "unifi.user"
	// ConfigKeyPass is the config key for the UniFi password.
	ConfigKeyPass = "unifi.pass"
	// ConfigKeyAPIKey is the config key for the UniFi API key.
	ConfigKeyAPIKey = "unifi.apikey"
	// ConfigKeyInsecure is the config key for allowing insecure TLS connections.
	ConfigKeyInsecure = "unifi.insecure"

	// DefaultURL is the default UniFi controller URL.
	DefaultURL = "https://10.69.1.1"
)

// NewFromConfig creates a UniFi client using the standard config resolution:
//
//  1. ~/.core/config.yaml keys: unifi.url, unifi.user, unifi.pass, unifi.apikey, unifi.insecure
//  2. UNIFI_URL + UNIFI_USER + UNIFI_PASS + UNIFI_APIKEY + UNIFI_INSECURE environment variables (override config file)
//  3. Provided flag overrides (highest priority; pass nil to skip)
func NewFromConfig(flagURL, flagUser, flagPass, flagAPIKey string, flagInsecure *bool) (*Client, error) {
	url, user, pass, apikey, insecure, err := ResolveConfig(flagURL, flagUser, flagPass, flagAPIKey, flagInsecure)
	if err != nil {
		return nil, err
	}

	if user == "" && apikey == "" {
		return nil, log.E("unifi.NewFromConfig", "no credentials configured (set UNIFI_USER/UNIFI_PASS or UNIFI_APIKEY, or run: core unifi config)", nil)
	}

	return New(url, user, pass, apikey, insecure)
}

// ResolveConfig resolves the UniFi URL and credentials from all config sources.
// Flag values take highest priority, then env vars, then config file.
func ResolveConfig(flagURL, flagUser, flagPass, flagAPIKey string, flagInsecure *bool) (url, user, pass, apikey string, insecure bool, err error) {
	// Start with config file values
	cfg, cfgErr := config.New()
	if cfgErr == nil {
		_ = cfg.Get(ConfigKeyURL, &url)
		_ = cfg.Get(ConfigKeyUser, &user)
		_ = cfg.Get(ConfigKeyPass, &pass)
		_ = cfg.Get(ConfigKeyAPIKey, &apikey)
		_ = cfg.Get(ConfigKeyInsecure, &insecure)
	}

	// Overlay environment variables
	if envURL := os.Getenv("UNIFI_URL"); envURL != "" {
		url = envURL
	}
	if envUser := os.Getenv("UNIFI_USER"); envUser != "" {
		user = envUser
	}
	if envPass := os.Getenv("UNIFI_PASS"); envPass != "" {
		pass = envPass
	}
	if envAPIKey := os.Getenv("UNIFI_APIKEY"); envAPIKey != "" {
		apikey = envAPIKey
	}
	if envInsecure := os.Getenv("UNIFI_INSECURE"); envInsecure != "" {
		insecure = envInsecure == "true" || envInsecure == "1"
	}

	// Overlay flag values (highest priority)
	if flagURL != "" {
		url = flagURL
	}
	if flagUser != "" {
		user = flagUser
	}
	if flagPass != "" {
		pass = flagPass
	}
	if flagAPIKey != "" {
		apikey = flagAPIKey
	}
	if flagInsecure != nil {
		insecure = *flagInsecure
	}

	// Default URL if nothing configured
	if url == "" {
		url = DefaultURL
	}

	return url, user, pass, apikey, insecure, nil
}

// SaveConfig persists the UniFi URL and/or credentials to the config file.
func SaveConfig(url, user, pass, apikey string, insecure *bool) error {
	cfg, err := config.New()
	if err != nil {
		return log.E("unifi.SaveConfig", "failed to load config", err)
	}

	if url != "" {
		if err := cfg.Set(ConfigKeyURL, url); err != nil {
			return log.E("unifi.SaveConfig", "failed to save URL", err)
		}
	}

	if user != "" {
		if err := cfg.Set(ConfigKeyUser, user); err != nil {
			return log.E("unifi.SaveConfig", "failed to save user", err)
		}
	}

	if pass != "" {
		if err := cfg.Set(ConfigKeyPass, pass); err != nil {
			return log.E("unifi.SaveConfig", "failed to save password", err)
		}
	}

	if apikey != "" {
		if err := cfg.Set(ConfigKeyAPIKey, apikey); err != nil {
			return log.E("unifi.SaveConfig", "failed to save API key", err)
		}
	}

	if insecure != nil {
		if err := cfg.Set(ConfigKeyInsecure, *insecure); err != nil {
			return log.E("unifi.SaveConfig", "failed to save insecure flag", err)
		}
	}

	return nil
}
