package ansible

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewSSHClient(t *testing.T) {
	cfg := SSHConfig{
		Host: "localhost",
		Port: 2222,
		User: "root",
	}

	client, err := NewSSHClient(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "localhost", client.host)
	assert.Equal(t, 2222, client.port)
	assert.Equal(t, "root", client.user)
	assert.Equal(t, 30*time.Second, client.timeout)
}

func TestSSHConfig_Defaults(t *testing.T) {
	cfg := SSHConfig{
		Host: "localhost",
	}

	client, err := NewSSHClient(cfg)
	assert.NoError(t, err)
	assert.Equal(t, 22, client.port)
	assert.Equal(t, "root", client.user)
	assert.Equal(t, 30*time.Second, client.timeout)
}
