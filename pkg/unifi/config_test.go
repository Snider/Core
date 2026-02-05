package unifi

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveConfig(t *testing.T) {
	// Clear environment variables to start clean
	os.Unsetenv("UNIFI_URL")
	os.Unsetenv("UNIFI_USER")
	os.Unsetenv("UNIFI_PASS")
	os.Unsetenv("UNIFI_APIKEY")
	os.Unsetenv("UNIFI_INSECURE")
	os.Unsetenv("CORE_CONFIG_UNIFI_URL")
	os.Unsetenv("CORE_CONFIG_UNIFI_USER")
	os.Unsetenv("CORE_CONFIG_UNIFI_PASS")
	os.Unsetenv("CORE_CONFIG_UNIFI_APIKEY")
	os.Unsetenv("CORE_CONFIG_UNIFI_INSECURE")

	// 1. Test defaults
	url, user, pass, apikey, insecure, err := ResolveConfig("", "", "", "", nil)
	assert.NoError(t, err)
	assert.Equal(t, DefaultURL, url)
	assert.Empty(t, user)
	assert.Empty(t, pass)
	assert.Empty(t, apikey)
	assert.False(t, insecure)

	// 2. Test environment variables
	t.Setenv("UNIFI_URL", "https://env.url")
	t.Setenv("UNIFI_USER", "envuser")
	t.Setenv("UNIFI_PASS", "envpass")
	t.Setenv("UNIFI_APIKEY", "envapikey")
	t.Setenv("UNIFI_INSECURE", "true")

	url, user, pass, apikey, insecure, err = ResolveConfig("", "", "", "", nil)
	assert.NoError(t, err)
	assert.Equal(t, "https://env.url", url)
	assert.Equal(t, "envuser", user)
	assert.Equal(t, "envpass", pass)
	assert.Equal(t, "envapikey", apikey)
	assert.True(t, insecure)

	// Test alternate UNIFI_INSECURE value
	t.Setenv("UNIFI_INSECURE", "1")
	_, _, _, _, insecure, _ = ResolveConfig("", "", "", "", nil)
	assert.True(t, insecure)

	// 3. Test flags (highest priority)
	trueVal := true
	url, user, pass, apikey, insecure, err = ResolveConfig("https://flag.url", "flaguser", "flagpass", "flagapikey", &trueVal)
	assert.NoError(t, err)
	assert.Equal(t, "https://flag.url", url)
	assert.Equal(t, "flaguser", user)
	assert.Equal(t, "flagpass", pass)
	assert.Equal(t, "flagapikey", apikey)
	assert.True(t, insecure)

	// 4. Flags should still override env vars
	falseVal := false
	url, user, pass, apikey, insecure, err = ResolveConfig("https://flag.url", "flaguser", "flagpass", "flagapikey", &falseVal)
	assert.NoError(t, err)
	assert.Equal(t, "https://flag.url", url)
	assert.Equal(t, "flaguser", user)
	assert.Equal(t, "flagpass", pass)
	assert.Equal(t, "flagapikey", apikey)
	assert.False(t, insecure)
}

func TestNewFromConfig(t *testing.T) {
	// Mock UniFi controller
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"meta":{"rc":"ok"}, "data": []}`)
	}))
	defer ts.Close()

	// 1. Success case
	client, err := NewFromConfig(ts.URL, "user", "pass", "", nil)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, ts.URL, client.URL())

	// 2. Error case: No credentials
	os.Unsetenv("UNIFI_USER")
	os.Unsetenv("UNIFI_APIKEY")
	client, err = NewFromConfig("", "", "", "", nil)
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "no credentials configured")
}

func TestSaveConfig(t *testing.T) {
	// Mock HOME to use temp dir for config
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Clear relevant env vars that might interfere
	os.Unsetenv("UNIFI_URL")
	os.Unsetenv("UNIFI_USER")
	os.Unsetenv("UNIFI_PASS")
	os.Unsetenv("UNIFI_APIKEY")
	os.Unsetenv("UNIFI_INSECURE")
	os.Unsetenv("CORE_CONFIG_UNIFI_URL")
	os.Unsetenv("CORE_CONFIG_UNIFI_USER")
	os.Unsetenv("CORE_CONFIG_UNIFI_PASS")
	os.Unsetenv("CORE_CONFIG_UNIFI_APIKEY")
	os.Unsetenv("CORE_CONFIG_UNIFI_INSECURE")

	err := SaveConfig("https://save.url", "saveuser", "savepass", "saveapikey", nil)
	assert.NoError(t, err)

	// Verify it saved by resolving it
	url, user, pass, apikey, insecure, err := ResolveConfig("", "", "", "", nil)
	assert.NoError(t, err)
	assert.Equal(t, "https://save.url", url)
	assert.Equal(t, "saveuser", user)
	assert.Equal(t, "savepass", pass)
	assert.Equal(t, "saveapikey", apikey)
	assert.False(t, insecure)

	// Test saving insecure true
	trueVal := true
	err = SaveConfig("", "", "", "", &trueVal)
	assert.NoError(t, err)
	_, _, _, _, insecure, _ = ResolveConfig("", "", "", "", nil)
	assert.True(t, insecure)
}
