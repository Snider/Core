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
	// ConfigKeyVerifyTLS is the config key for whether to verify UniFi TLS certificates.
	ConfigKeyVerifyTLS = "unifi.verify_tls"

	// DefaultURL is the default UniFi controller URL.
	DefaultURL = "https://10.69.1.1"
)

// NewFromConfig creates a UniFi client using the standard config resolution:
//
//  1. ~/.core/config.yaml keys: unifi.url, unifi.user, unifi.pass, unifi.apikey, unifi.verify_tls
//  2. UNIFI_URL + UNIFI_USER + UNIFI_PASS + UNIFI_APIKEY + UNIFI_VERIFY_TLS environment variables (override config file)
//  3. Provided flag overrides (highest priority; pass empty to skip)
func NewFromConfig(flagURL, flagUser, flagPass, flagAPIKey string) (*Client, error) {
	url, user, pass, apikey, verifyTLS, err := ResolveConfig(flagURL, flagUser, flagPass, flagAPIKey)
	if err != nil {
		return nil, err
	}

	if user == "" && apikey == "" {
		return nil, log.E("unifi.NewFromConfig", "no credentials configured (set UNIFI_USER/UNIFI_PASS or UNIFI_APIKEY, or run: core unifi config)", nil)
	}

	return New(url, user, pass, apikey, verifyTLS)
}

// ResolveConfig resolves the UniFi URL and credentials from all config sources.
// Flag values take highest priority, then env vars, then config file.
func ResolveConfig(flagURL, flagUser, flagPass, flagAPIKey string) (url, user, pass, apikey string, verifyTLS bool, err error) {
	// Default to true
	verifyTLS = true

	// Start with config file values
	cfg, cfgErr := config.New()
	if cfgErr == nil {
		_ = cfg.Get(ConfigKeyURL, &url)
		_ = cfg.Get(ConfigKeyUser, &user)
		_ = cfg.Get(ConfigKeyPass, &pass)
		_ = cfg.Get(ConfigKeyAPIKey, &apikey)
		_ = cfg.Get(ConfigKeyVerifyTLS, &verifyTLS)
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
	if envVerify := os.Getenv("UNIFI_VERIFY_TLS"); envVerify != "" {
		verifyTLS = envVerify == "true" || envVerify == "1"
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

	// Default URL if nothing configured
	if url == "" {
		url = DefaultURL
	}

	return url, user, pass, apikey, verifyTLS, nil
}

// SaveConfig persists the UniFi URL and/or credentials to the config file.
// Pass setVerifyTLS=true to update the verify_tls setting.
func SaveConfig(url, user, pass, apikey string, setVerifyTLS bool, verifyTLS bool) error {
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

	if setVerifyTLS {
		if err := cfg.Set(ConfigKeyVerifyTLS, verifyTLS); err != nil {
			return log.E("unifi.SaveConfig", "failed to save verify_tls", err)
		}
	}

	return nil
}
