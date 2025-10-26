// package io provides a unified interface for interacting with different filesystems.
package io

import (
	// Import the internal implementation with an alias.
	impl "github.com/Snider/Core/pkg/io"
	"github.com/Snider/Core/pkg/io/sftp"
	"github.com/Snider/Core/pkg/io/webdav"
)

// Medium is the standard interface for a storage backend.
type Medium = impl.Medium

// Expose the factory functions for creating different media.
var (
	NewSFTPMedium   = impl.NewSFTPMedium
	NewWebDAVMedium = impl.NewWebDAVMedium
)

// Expose the connection config structs for convenience.
type (
	SFTPConnectionConfig   = sftp.ConnectionConfig
	WebDAVConnectionConfig = webdav.ConnectionConfig
)
