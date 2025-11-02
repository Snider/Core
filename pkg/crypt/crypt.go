package crypt

import (
	"github.com/Snider/Core/pkg/core"
	"github.com/Snider/Core/pkg/crypt/internal"
)

// HandleIPCEvents processes IPC messages for the crypt service.
func (s *Service) HandleIPCEvents(c *core.Core, msg core.Message) error {
	switch msg.(type) {
	case core.ActionServiceStartup:
		// Crypt is stateless, no startup needed.
		return nil
	default:
		if c.App != nil && c.App.Logger != nil {
			c.App.Logger.Debug("Crypt: Unhandled message type", "type", fmt.Sprintf("%T", msg))
		}
	}
	return nil
}

// Options holds configuration for the crypt service.
type Options = internal.Options

// Service provides cryptographic functions to the application.
type Service = internal.Service

// HashType defines the supported hashing algorithms.
type HashType = internal.HashType

const (
	LTHN   = internal.LTHN
	SHA512 = internal.SHA512
	SHA256 = internal.SHA256
	SHA1   = internal.SHA1
	MD5    = internal.MD5
)

// New is the constructor for static dependency injection.
func New() (*Service, error) {
	return internal.New()
}

// Register is the constructor for dynamic dependency injection.
func Register(c *core.Core) (any, error) {
	return internal.Register(c)
}
