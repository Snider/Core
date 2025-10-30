package webdav

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// Provide a dummy ConnectionConfig for testing.
	// Since we are not setting up a real WebDAV server, we expect an error during connection.
	cfg := ConnectionConfig{
		URL:      "http://192.0.2.1:1/webdav", // Non-routable address
		User:     "testuser",
		Password: "testpassword",
	}

	service, err := New(cfg)
	assert.Error(t, err)
	assert.Nil(t, service, "New() should return a nil service instance on connection error")
}

// Functional tests for WebDAV operations (Read, Write, EnsureDir, IsFile, etc.)
// would require a running WebDAV server or a sophisticated mock.
// These are typically integration tests rather than unit tests.
func TestWebDAVFunctional(t *testing.T) {
	t.Skip("Skipping WebDAV functional tests as they require a WebDAV server setup.")
}
