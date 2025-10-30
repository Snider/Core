package sftp

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTest creates a temporary home directory and a dummy known_hosts file
// to prevent tests from failing in CI environments where the file doesn't exist.
func setupTest(t *testing.T) {
	t.Helper()
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	sshDir := filepath.Join(homeDir, ".ssh")
	err := os.Mkdir(sshDir, 0700)
	require.NoError(t, err)
	knownHostsFile := filepath.Join(sshDir, "known_hosts")
	err = os.WriteFile(knownHostsFile, []byte{}, 0600)
	require.NoError(t, err)
}

func TestNew(t *testing.T) {
	setupTest(t)
	// Provide a dummy ConnectionConfig for testing.
	// Since we are not setting up a real SFTP server, we expect an error during connection.
	cfg := ConnectionConfig{
		Host: "localhost",
		Port: "22",
		User: "testuser",
		// No password or keyfile provided, so connection should fail.
	}

	service, err := New(cfg)
	assert.Error(t, err)
	assert.Nil(t, service, "New() should return a nil service instance on connection error")
	assert.Contains(t, err.Error(), "no authentication method provided", "Expected authentication error")
}

func TestNew_InvalidHost(t *testing.T) {
	setupTest(t)
	cfg := ConnectionConfig{
		Host:     "non-resolvable-host.domain.invalid",
		Port:     "22",
		User:     "testuser",
		Password: "password",
	}

	service, err := New(cfg)
	assert.Error(t, err)
	assert.Nil(t, service)
	assert.Contains(t, err.Error(), "lookup non-resolvable-host.domain.invalid")
}

func TestNew_InvalidPort(t *testing.T) {
	setupTest(t)
	cfg := ConnectionConfig{
		Host:     "localhost",
		Port:     "99999", // Invalid port number
		User:     "testuser",
		Password: "password",
	}

	service, err := New(cfg)
	assert.Error(t, err)
	assert.Nil(t, service)
	assert.Contains(t, err.Error(), "invalid port")
}

func TestNew_ConnectionTimeout(t *testing.T) {
	setupTest(t)
	cfg := ConnectionConfig{
		Host:     "192.0.2.0", // Non-routable IP to simulate timeout
		Port:     "22",
		User:     "testuser",
		Password: "password",
		Timeout:  100 * time.Millisecond,
	}

	service, err := New(cfg)
	assert.Error(t, err)
	assert.Nil(t, service)
	assert.Contains(t, err.Error(), "i/o timeout")
}

func TestNew_AuthFailure_NonexistentKeyfile(t *testing.T) {
	setupTest(t)
	cfg := ConnectionConfig{
		Host:    "localhost",
		Port:    "22",
		User:    "testuser",
		KeyFile: "/path/to/nonexistent/keyfile",
	}

	service, err := New(cfg)
	assert.Error(t, err)
	assert.Nil(t, service)
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestNew_AuthFailure_InvalidKeyFormat(t *testing.T) {
	setupTest(t)
	// Create a temporary file with invalid key content
	tmpFile, err := os.CreateTemp("", "invalid_key")
	require.NoError(t, err)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Logf("Failed to remove temporary file: %v", err)
		}
	}(tmpFile.Name())

	_, err = tmpFile.WriteString("not a valid ssh key")
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)

	cfg := ConnectionConfig{
		Host:    "localhost",
		Port:    "22",
		User:    "testuser",
		KeyFile: tmpFile.Name(),
	}

	service, err := New(cfg)
	assert.Error(t, err)
	assert.Nil(t, service)
	assert.Contains(t, err.Error(), "unable to parse private key")
}

func TestNew_MultipleAuthMethods(t *testing.T) {
	setupTest(t)
	// Create a temporary file with invalid key content to ensure key-based auth is attempted
	tmpFile, err := os.CreateTemp("", "dummy_key")
	require.NoError(t, err)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Logf("Failed to remove temporary file: %v", err)
		}
	}(tmpFile.Name())

	_, err = tmpFile.WriteString("not a valid ssh key")
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)

	cfg := ConnectionConfig{
		Host:     "localhost",
		Port:     "22",
		User:     "testuser",
		Password: "password",
		KeyFile:  tmpFile.Name(),
	}

	service, err := New(cfg)
	assert.Error(t, err)
	assert.Nil(t, service)
	// We expect the key file to be prioritized, so we should get a parse error, not a "no auth method" error.
	assert.Contains(t, err.Error(), "unable to parse private key")
}
