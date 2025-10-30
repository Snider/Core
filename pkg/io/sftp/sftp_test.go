package sftp

import (
	"testing"

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

// Functional tests for SFTP operations (Read, Write, EnsureDir, IsFile, etc.)
// would require a running SFTP server or a sophisticated mock.
// These are typically integration tests rather than unit tests.
func TestSFTPFunctional(t *testing.T) {
	t.Skip("Skipping SFTP functional tests as they require an SFTP server setup.")
}
