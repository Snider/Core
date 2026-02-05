package unifi

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveConfig(t *testing.T) {
	// Set env vars
	os.Setenv("UNIFI_URL", "https://env-url")
	os.Setenv("UNIFI_USER", "env-user")
	os.Setenv("UNIFI_PASS", "env-pass")
	os.Setenv("UNIFI_VERIFY_TLS", "false")
	defer func() {
		os.Unsetenv("UNIFI_URL")
		os.Unsetenv("UNIFI_USER")
		os.Unsetenv("UNIFI_PASS")
		os.Unsetenv("UNIFI_VERIFY_TLS")
	}()

	url, user, pass, apikey, verifyTLS, err := ResolveConfig("", "", "", "", nil)
	assert.NoError(t, err)
	assert.Equal(t, "https://env-url", url)
	assert.Equal(t, "env-user", user)
	assert.Equal(t, "env-pass", pass)
	assert.Equal(t, "", apikey)
	assert.False(t, verifyTLS)

	// Flag overrides
	url, user, pass, apikey, verifyTLS, err = ResolveConfig("https://flag-url", "flag-user", "flag-pass", "flag-apikey", nil)
	assert.NoError(t, err)
	assert.Equal(t, "https://flag-url", url)
	assert.Equal(t, "flag-user", user)
	assert.Equal(t, "flag-pass", pass)
	assert.Equal(t, "flag-apikey", apikey)
	// Env var for verifyTLS still applies if not overridden in ResolveConfig (which it isn't currently via flags)
	assert.False(t, verifyTLS)
}
