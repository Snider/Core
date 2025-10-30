package sftp

import (
	"time"

	"github.com/pkg/sftp"
)

// Medium implements the io.Medium interface for the SFTP protocol.
type Medium struct {
	client *sftp.Client
}

// ConnectionConfig holds the necessary details to connect to an SFTP server.
type ConnectionConfig struct {
	Host     string
	Port     string
	User     string
	Password string // For password-based auth
	KeyFile  string // Path to a private key for key-based auth

	// Timeout specifies the duration for the network connection. If set to 0,
	// a default timeout of 30 seconds will be used.
	Timeout time.Duration
}
