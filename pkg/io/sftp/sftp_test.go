package sftp

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
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
